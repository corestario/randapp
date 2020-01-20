package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/corestario/dkglib/lib/blsShare"
	cfg "github.com/tendermint/tendermint/config"
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

func initFilesWithConfig(config *cfg.Config, logger log.Logger) error {
	// private validator

	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	err := os.MkdirAll(path.Join(config.RootDir, "config"), nodeDirPerm)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(path.Join(config.RootDir, "data"), nodeDirPerm)
	if err != nil {
		panic(err)
	}
	logger.Info("CONFIG private validator", "rootDir", config.RootDir, "keyFile", privValKeyFile,
		"stateFile", privValStateFile)
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
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// todo what should we do if bls key not exsists
	blsKeyFile := config.BLSKeyFile()
	if cmn.FileExists(blsKeyFile) {
		logger.Info("Found node key", "path", blsKeyFile)
	} else {
		f, err := os.Create(blsKeyFile)
		if err != nil {
			return err
		}
		defer f.Close()
		share, ok := blsShare.TestnetShares[config.NodeID]
		if !ok {
			return fmt.Errorf("node id #%d is unexpected", config.NodeID)
		}
		err = json.NewEncoder(f).Encode(share)
		if err != nil {
			return err
		}

		logger.Info("Generated node key", "path", blsKeyFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if cmn.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {
		genDoc := types.GenesisDoc{
			ChainID:         fmt.Sprintf("test-chain-%v", cmn.RandStr(6)),
			GenesisTime:     tmtime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
		}
		key := pv.GetPubKey()
		genDoc.Validators = []types.GenesisValidator{{
			Address: key.Address(),
			PubKey:  key,
			Power:   10,
		}}

		// This keypair allows for single-node execution, e.g. $ tendermint node.
		genDoc.BLSMasterPubKey = blsShare.DefaultBLSVerifierMasterPubKey
		genDoc.BLSThreshold = 3
		genDoc.BLSNumShares = 4
		genDoc.DKGNumBlocks = 10
		logger.Info("Generated geneAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAsis file", "path", genDoc)

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		logger.Info("Generated genesis file", "path", genFile)
		logger.Info("Generated geneAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAsis file", "path", genDoc)
	}

	return nil
}
