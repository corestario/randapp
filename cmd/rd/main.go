package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/server"
	tx "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	genaccscli "github.com/cosmos/cosmos-sdk/x/auth/genaccounts/client/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	app "github.com/dgamingfoundation/randapp"
	xapp "github.com/dgamingfoundation/randapp/x/randapp"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	storeAcc      = "acc"
	storeNS       = "randapp"
	flagOverwrite = "overwrite"
)

func main() {
	cobra.EnableCommandSorting = false

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "rd",
		Short:             "randapp App Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(
		genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(ctx, cdc, genaccounts.AppModuleBasic{}, app.DefaultNodeHome),
		genutilcli.GenTxCmd(ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{}, genaccounts.AppModuleBasic{}, app.DefaultNodeHome, app.DefaultCLIHome),
		genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics),
		// AddGenesisAccountCmd allows users to add accounts to the genesis file
		genaccscli.AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		InitCmd(ctx, cdc),
		AddGenesisAccountCmd(ctx, cdc),
	)

	server.AddCommands(ctx, cdc, rootCmd, newApp, appExporter())

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "NS", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func appExporter() server.AppExporter {
	return func(logger log.Logger, db dbm.DB, _ io.Writer, _ int64, _ bool, _ []string) (
		json.RawMessage, []tmtypes.GenesisValidator, error) {
		dapp := app.NewRandApp(logger, db)
		return dapp.ExportAppStateAndValidators()
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewRandApp(logger, db)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		nsApp := app.NewRandApp(logger, db)
		err := nsApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return nsApp.ExportAppStateAndValidators()
	}

	nsApp := app.NewRandApp(logger, db)

	return nsApp.ExportAppStateAndValidators()
}

// InitCmd initializes all files for tendermint and application
func InitCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis config, priv-validator file, and p2p-node file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			chainID := viper.GetString(client.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}

			_, pk, err := genutil.InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			var appState json.RawMessage
			genFile := config.GenesisFile()

			if !viper.GetBool(flagOverwrite) && common.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			genesis := xapp.GenesisState{
				AuthData: auth.DefaultGenesisState(),
				BankData: bank.DefaultGenesisState(),
			}

			appState, err = codec.MarshalJSONIndent(cdc, genesis)
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

			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			fmt.Printf("Initialized rd configuration and bootstrapping files in %s...\n", viper.GetString(cli.HomeFlag))
			return nil
		},
	}

	cmd.Flags().String(cli.HomeFlag, app.DefaultNodeHome, "node's home directory")
	cmd.Flags().String(client.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")

	return cmd
}

func queryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		tx.GetQueryCmd(cdc),
		client.LineBreak,
		tx.GetAccountCmd(cdc),
	)

	return queryCmd
}

func txCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		tx.GetSignCommand(cdc),
		tx.GetBroadcastCommand(cdc),
		client.LineBreak,
	)

	return txCmd
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
		Power:  10,
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
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			coins.Sort()

			var genDoc tmtypes.GenesisDoc
			config := ctx.Config
			genFile := config.GenesisFile()
			if !common.FileExists(genFile) {
				return fmt.Errorf("%s does not exist, run `gaiad init` first", genFile)
			}
			genContents, err := ioutil.ReadFile(genFile)
			if err != nil {
			}

			if err = cdc.UnmarshalJSON(genContents, &genDoc); err != nil {
				return err
			}

			var appState xapp.GenesisState
			if err = cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
				return err
			}

			for _, stateAcc := range appState.Accounts {
				if stateAcc.Address.Equals(addr) {
					return fmt.Errorf("the application state already contains account %v", addr)
				}
			}

			acc := auth.NewBaseAccountWithAddress(addr)
			acc.Coins = coins
			appState.Accounts = append(appState.Accounts, &acc)
			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			return ExportGenesisFile(genFile, genDoc.ChainID, genDoc.Validators, appStateJSON)
		},
	}
	return cmd
}
