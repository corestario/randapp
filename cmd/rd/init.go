package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	rcfg "github.com/dgamingfoundation/randapp/x/randapp/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"
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
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))
			chainID := viper.GetString(client.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}

			srvConfig := rcfg.DefaultRAServerConfig()

			nodeID, _, err := genutil.InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			_, pk, err := genutil.InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			var appState json.RawMessage
			genFile := config.GenesisFile()

			if !viper.GetBool(flagOverwrite) && common.FileExists(genFile) {
				stateJSON, err := ioutil.ReadFile(os.ExpandEnv("$HOME/.rd") + "/config/genesis.json")
				if err != nil {
					return err
				}

				var genesisState map[string]json.RawMessage
				err = cdc.UnmarshalJSON(stateJSON, &genesisState)
				if err != nil {
					panic(err)
				}
				var validators []tmtypes.GenesisValidator
				val, ok := genesisState["validators"]
				if !ok {
					return fmt.Errorf("no validators in genesis file")
				}
				err = cdc.UnmarshalJSON(val, &validators)
				if err != nil {
					panic(err)
				}
				//return fmt.Errorf("genesis.json file already exists: %v", genFile)
				dg := mbm.DefaultGenesis()

				appState, err = codec.MarshalJSONIndent(cdc, dg)
				if err != nil {
					return err
				}

				if err = ExportGenesisFile(genFile, chainID, validators, appState); err != nil {
					return err
				}
			} else {
				appState, err = codec.MarshalJSONIndent(cdc, mbm.DefaultGenesis())
				if err != nil {
					return err
				}
				_, _, validator, err := SimpleAppGenTx(cdc, pk)
				if err != nil {
					return err
				}
				if err = ExportGenesisFile(genFile, chainID, []tmtypes.GenesisValidator{validator}, appState); err != nil {
					return err
				}
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)
			rcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "server.toml"), srvConfig)
			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			return displayInfo(cdc, toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(client.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")

	return cmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}

// SimpleAppGenTx returns a simple GenTx command that makes the node a valdiator from the start
func SimpleAppGenTx(cdc *codec.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	addr, secret, err := server.GenerateCoinKey()
	if err != nil {
		return
	}

	bz, err := cdc.MarshalJSON(struct {
		Addr sdk.AccAddress `json:"addr"`
	}{addr})
	if err != nil {
		return
	}

	appGenTx = json.RawMessage(bz)

	bz, err = cdc.MarshalJSON(map[string]string{"secret": secret})
	if err != nil {
		return
	}

	cliPrint = json.RawMessage(bz)

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  100,
	}

	return
}

func ExportGenesisFile(
	genFile,
	chainID string,
	validators []tmtypes.GenesisValidator,
	appState json.RawMessage,
) error {
	if err := writeBLSShare(); err != nil {
		return fmt.Errorf("failed to writeBLSShare: %v", err)
	}

	genDoc := tmtypes.GenesisDoc{
		ChainID:         chainID,
		Validators:      validators,
		AppState:        appState,
		BLSThreshold:    1,
		BLSNumShares:    2,
		BLSMasterPubKey: tmtypes.DefaultBLSVerifierMasterPubKey,
		DKGNumBlocks:    1000,
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return err
	}

	return genDoc.SaveAs(genFile)
}

func writeBLSShare() error {
	blsShare := &tmtypes.BLSShareJSON{
		Pub:  tmtypes.DefaultBLSVerifierPubKey,
		Priv: tmtypes.DefaultBLSVerifierPrivKey,
	}
	usr, err := user.Current()
	if err != nil {
		return err
	}

	blsKeyFile := usr.HomeDir + "/" + ".rd/config/bls_key.json"
	// todo what should we do if bls key not exists.
	if cmn.FileExists(blsKeyFile) {
		fmt.Println("Found node key", "path", blsKeyFile)
	} else {
		f, err := os.Create(blsKeyFile)
		if err != nil {
			return err
		}
		defer f.Close()
		err = json.NewEncoder(f).Encode(blsShare)
		if err != nil {
			return err
		}

		fmt.Println("Generated node key", "path", blsKeyFile)
	}

	return nil
}

// AddGenesisAccountCmd allows users to add accounts to the genesis file
func AddGenesisAccountCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account [address] [coins[,coins]]",
		Short: "Adds an account to the genesis file",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(`
Adds accounts to the genesis file so that you can start a chain with coins in the CLI:
$ rd add-genesis-account cosmos1tse7r2fadvlrrgau3pa0ss7cqh55wrv6y9alwh 1000STAKE,1000nametoken
`),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			// add genesis account to the app state
			var genesisAccounts genaccounts.GenesisAccounts

			cdc.MustUnmarshalJSON(appState[genaccounts.ModuleName], &genesisAccounts)

			if genesisAccounts.Contains(addr) {
				return fmt.Errorf("cannot add account at existing address %v", addr)
			}

			acc := auth.NewBaseAccountWithAddress(addr)
			acc.Coins = coins

			genAcc := genaccounts.NewGenesisAccount(&acc)

			genesisAccounts = append(genesisAccounts, genAcc)

			genesisStateBz := cdc.MustMarshalJSON(genaccounts.GenesisState(genesisAccounts))
			appState[genaccounts.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			// export app state
			genDoc.AppState = appStateJSON

			return ExportGenesisFile(genFile, genDoc.ChainID, genDoc.Validators, appStateJSON)
		},
	}
	return cmd
}
