package main

import (
	"encoding/json"
	"fmt"
	lg "log"
	"os"
	"path/filepath"

	"github.com/corestario/dkglib/lib/blsShare"
	mpcfg "github.com/corestario/randapp/x/randapp/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	flagOverwrite  = "overwrite"
	flagClientHome = "home-client"
)

type printInfo struct {
	Moniker    string          `json:"moniker"`
	ChainID    string          `json:"chain_id"`
	NodeID     string          `json:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string,
	appMessage json.RawMessage) printInfo {

	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(cdc *codec.Codec, info printInfo) error {
	out, err := codec.MarshalJSONIndent(cdc, info)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s\n", string(out)) // nolint: errcheck
	return nil
}

// InitCmd returns a command that initializes all files needed for Tendermint
// and the respective application.
func InitCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager,
	defaultNodeHome string) *cobra.Command { // nolint: golint
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			lg.Println("INIIIIIIIIIIIIIIIITTTT")
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))
			chainID := viper.GetString(flags.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}

			srvConfig := mpcfg.DefaultRAServerConfig()

			nodeID, _, err := genutil.InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			genFile := config.GenesisFile()
			if !viper.GetBool(flagOverwrite) && common.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}
			appState, err := codec.MarshalJSONIndent(cdc, mbm.DefaultGenesis())
			if err != nil {
				return err
			}

			genDoc := &types.GenesisDoc{}
			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				genDoc, err = types.GenesisDocFromFile(genFile)
				if err != nil {
					return err
				}
			}

			genDoc.ChainID = chainID
			genDoc.Validators = nil
			genDoc.AppState = appState
			//genDoc, err = fillGenDocConfig(config, genDoc)
			//if err != nil {
			//	return err
			//}
			if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
				return err
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)
			mpcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "server.toml"), srvConfig)

			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			return displayInfo(cdc, toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")

	return cmd
}

func fillGenDocConfig(config *cfg.Config, genDoc types.GenesisDoc) (types.GenesisDoc, error) {
	// private validator
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	if cmn.FileExists(privValKeyFile) {
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
		logger.Info("Found private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	} else {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		logger.Info("Generated private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	}

	nodeKeyFile := config.NodeKeyFile()
	if cmn.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return types.GenesisDoc{}, err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// todo what should we do if bls key not exsists
	blsKeyFile := config.BLSKeyFile()
	//if cmn.FileExists(blsKeyFile) {
	//	logger.Info("Found node key", "path", blsKeyFile)
	//} else {
	f, err := os.Create(blsKeyFile)
	if err != nil {
		return types.GenesisDoc{}, err
	}
	defer f.Close()
	share, ok := blsShare.TestnetShares[config.NodeID]
	if !ok {
		return types.GenesisDoc{}, fmt.Errorf("node id #%d is unexpected", config.NodeID)
	}
	err = json.NewEncoder(f).Encode(share)
	if err != nil {
		return types.GenesisDoc{}, err
	}

	logger.Info("Generated node key", "path", blsKeyFile)
	//}

	// genesis file

	genDoc.GenesisTime = tmtime.Now()
	genDoc.ConsensusParams = types.DefaultConsensusParams()

	key := pv.GetPubKey()
	genDoc.Validators = []types.GenesisValidator{{
		Address: key.Address(),
		PubKey:  key,
		Power:   10,
	}}

	// This keypair allows for single-node execution, e.g. $ tendermint node.
	genDoc.BLSMasterPubKey = blsShare.DefaultBLSVerifierMasterPubKey
	genDoc.BLSThreshold = 2
	genDoc.BLSNumShares = 4
	genDoc.DKGNumBlocks = 10
	logger.Info("Generated geneAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAsis file", "path", genDoc.BLSShare)

	return genDoc, nil
}
