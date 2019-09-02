package main

import (
	"encoding/json"
	"fmt"
	"io"
	l "log"
	"net/http"
	"os"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	genaccscli "github.com/cosmos/cosmos-sdk/x/genaccounts/client/cli"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
	app "github.com/dgamingfoundation/randapp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

// DefaultNodeHome sets the folder where the application data and configuration will be stored
var DefaultNodeHome = os.ExpandEnv("$HOME/.rd")

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "rd",
		Short:             "randapp App Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}
	// CLI commands to initialize the chain
	rootCmd.AddCommand(
		InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(ctx, cdc, genaccounts.AppModuleBasic{}, app.DefaultNodeHome),
		genutilcli.GenTxCmd(ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{}, genaccounts.AppModuleBasic{}, app.DefaultNodeHome, app.DefaultCLIHome),
		genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics),
		// AddGenesisAccountCmd allows users to add accounts to the genesis file
		genaccscli.AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
	)

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)
	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "NS", app.DefaultNodeHome)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9080", nil); err != nil {
			l.Fatalf("failed to run prometheus: %v", err)
		}
	}()
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
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
		return nsApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	nsApp := app.NewRandApp(logger, db)

	return nsApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

func ExportGenesisFile(
	genFile,
	chainID string,
	validators []types.GenesisValidator,
	appState json.RawMessage,
) error {
	if err := writeBLSShare(); err != nil {
		return fmt.Errorf("failed to writeBLSShare: %v", err)
	}

	genDoc := types.GenesisDoc{
		ChainID:         chainID,
		Validators:      validators,
		AppState:        appState,
		BLSThreshold:    1,
		BLSNumShares:    2,
		BLSMasterPubKey: types.DefaultBLSVerifierMasterPubKey,
		DKGNumBlocks:    1000,
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return err
	}

	return genDoc.SaveAs(genFile)
}

func writeBLSShare() error {
	blsShare := &types.BLSShareJSON{
		Pub:  types.DefaultBLSVerifierPubKey,
		Priv: types.DefaultBLSVerifierPrivKey,
	}

	blsKeyFile := "/home/andrey/.rd/config/bls_key.json"
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
