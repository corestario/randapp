package randapp

import (
	"encoding/json"
	"os"

	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/dgamingfoundation/randapp/x/randapp"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/cosmos-sdk/x/auth/genaccounts"

	"github.com/cosmos/cosmos-sdk/x/bank"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	app "github.com/dgamingfoundation/randapp/x/randapp"

	logg "log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	appName = "randapp"
)

var (
	// default home directories for the application CLI
	DefaultCLIHome = os.ExpandEnv("$HOME/.rcli")

	// DefaultNodeHome sets the folder where the applcation data and configuration will be stored
	DefaultNodeHome = os.ExpandEnv("$HOME/.rd")

	// ModuleBasicManager is in charge of setting up basic module elemnets
	ModuleBasics = module.NewBasicManager(
		genaccounts.AppModuleBasic{},
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		params.AppModuleBasic{},
		app.AppModule{},
		staking.AppModuleBasic{},
		distr.AppModuleBasic{},
		slashing.AppModuleBasic{},
	)
)

type randApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	keyMain          *sdk.KVStoreKey
	keyAccount       *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey
	keyParams        *sdk.KVStoreKey
	tkeyParams       *sdk.TransientStoreKey
	keyStaking       *sdk.KVStoreKey
	tkeyStaking      *sdk.TransientStoreKey
	keyDistr         *sdk.KVStoreKey
	tkeyDistr        *sdk.TransientStoreKey
	keySlashing      *sdk.KVStoreKey

	accountKeeper       auth.AccountKeeper
	bankKeeper          bank.Keeper
	feeCollectionKeeper auth.FeeCollectionKeeper
	paramsKeeper        params.Keeper
	rKeeper             randapp.Keeper
	stakingKeeper       staking.Keeper
	slashingKeeper      slashing.Keeper
	distrKeeper         distr.Keeper

	keyPubKeys            *sdk.KVStoreKey
	keyDeals              *sdk.KVStoreKey
	keyResponses          *sdk.KVStoreKey
	keyJustifications     *sdk.KVStoreKey
	keyCommits            *sdk.KVStoreKey
	keyComplaints         *sdk.KVStoreKey
	keyReconstructCommits *sdk.KVStoreKey

	mm *module.Manager
}

// NewRandApp is a constructor function for randApp.
func NewRandApp(logger log.Logger, db dbm.DB) *randApp {

	// First define the top level codec that will be shared by the different modules
	cdc := MakeCodec()

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc))

	// Here you initialize your application with the store keys it requires
	var app = &randApp{
		BaseApp: bApp,
		cdc:     cdc,

		keyMain:          sdk.NewKVStoreKey("main"),
		keyAccount:       sdk.NewKVStoreKey("acc"),
		keyFeeCollection: sdk.NewKVStoreKey("fee_collection"),
		keyParams:        sdk.NewKVStoreKey("params"),
		tkeyParams:       sdk.NewTransientStoreKey("transient_params"),
		keyStaking:       sdk.NewKVStoreKey(staking.StoreKey),
		tkeyStaking:      sdk.NewTransientStoreKey(staking.TStoreKey),
		keyDistr:         sdk.NewKVStoreKey(distr.StoreKey),
		tkeyDistr:        sdk.NewTransientStoreKey(distr.TStoreKey),
		keySlashing:      sdk.NewKVStoreKey(slashing.StoreKey),

		keyPubKeys:            sdk.NewKVStoreKey("pub_keys"),
		keyDeals:              sdk.NewKVStoreKey("deals"),
		keyResponses:          sdk.NewKVStoreKey("responses"),
		keyJustifications:     sdk.NewKVStoreKey("justifications"),
		keyCommits:            sdk.NewKVStoreKey("commits"),
		keyComplaints:         sdk.NewKVStoreKey("complaints"),
		keyReconstructCommits: sdk.NewKVStoreKey("reconstruct_commits"),
	}

	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams, sdk.CodespaceRoot)

	stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
	distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)

	// The AccountKeeper handles address -> account lookups
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		app.keyAccount,
		app.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
	)

	// The FeeCollectionKeeper collects transaction fees and renders them to the fee distribution module.
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(cdc, app.keyFeeCollection)

	stakingKeeper := staking.NewKeeper(
		app.cdc,
		app.keyStaking,
		app.tkeyStaking,
		app.bankKeeper,
		stakingSubspace,
		staking.DefaultCodespace,
	)

	app.distrKeeper = distr.NewKeeper(
		app.cdc,
		app.keyDistr,
		distrSubspace,
		app.bankKeeper,
		&stakingKeeper,
		app.feeCollectionKeeper,
		distr.DefaultCodespace,
	)

	app.slashingKeeper = slashing.NewKeeper(
		app.cdc,
		app.keySlashing,
		&stakingKeeper,
		slashingSubspace,
		slashing.DefaultCodespace,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(
			app.distrKeeper.Hooks(),
			app.slashingKeeper.Hooks()),
	)

	app.rKeeper = randapp.NewKeeper(
		app.bankKeeper,
		app.keyPubKeys,
		app.keyDeals,
		app.keyResponses,
		app.keyJustifications,
		app.keyCommits,
		app.keyComplaints,
		app.keyReconstructCommits,
		app.cdc,
	)

	// The AnteHandler handles signature verification and transaction pre-processing.
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper, auth.DefaultSigVerificationGasConsumer))

	// The app.Router is the main transaction router where each module registers its routes.
	// Register the bank and randapp routes here.
	//app.Router().
	//	AddRoute("bank", bank.NewHandler(app.bankKeeper)).
	//AddRoute("randapp", randapp.NewHandler(app.rKeeper))
	//
	//// The app.QueryRouter is the main query router where each module registers its routes.
	//app.QueryRouter().
	//	AddRoute("randapp", randapp.NewQuerier(app.rKeeper)).
	//	AddRoute("acc", auth.NewQuerier(app.accountKeeper))

	app.mm = module.NewManager(
		genaccounts.NewAppModule(app.accountKeeper),
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper, app.feeCollectionKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		randapp.NewAppModule(app.rKeeper, app.bankKeeper),
		distr.NewAppModule(app.distrKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.feeCollectionKeeper, app.distrKeeper, app.accountKeeper),
	)

	app.mm.SetOrderInitGenesis(
		genaccounts.ModuleName,
		distr.ModuleName,
		staking.ModuleName,
		auth.ModuleName,
		bank.ModuleName,
		slashing.ModuleName,
		randapp.ModuleName,
		genutil.ModuleName,
	)

	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// The initChainer handles translating the genesis.json file into initial state for the network.
	app.SetInitChainer(app.initChainer)

	app.MountStores(
		app.keyMain,
		app.keyAccount,
		app.keyFeeCollection,
		app.keyStaking,
		app.tkeyStaking,
		app.keyDistr,
		app.tkeyDistr,
		app.keySlashing,
		app.keyParams,
		app.tkeyParams,

		app.keyPubKeys,
		app.keyDeals,
		app.keyResponses,
		app.keyJustifications,
		app.keyCommits,
		app.keyComplaints,
		app.keyReconstructCommits,
	)

	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

func (app *randApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes
	//genesisState := new(randapp.GenesisState)
	/*
			var fs []os.FileInfo
			fs, _ = ioutil.ReadDir("./")
			for _, f := range fs {
				logg.Println(f.Name())
			}
			logg.Println()
			fs, _ = ioutil.ReadDir("/root/")
			for _, f := range fs {
				logg.Println(f.Name())
			}
			logg.Println()
			fs, _ = ioutil.ReadDir("/root/.rd")
			for _, f := range fs {
				logg.Println(f.Name())
			}
			logg.Println()
			fs, _ = ioutil.ReadDir("/root/.rd/config")
			for _, f := range fs {
				logg.Println(f.Name())
			}
			logg.Println()


			$BINARY init --chain-id SO3Hlq validator${ID}
			echo 'KEYS ADD'
			$UNARY keys add validator${ID} <<< '12345678'
			echo 'ADD-GENESIS'
			echo $($UNARY keys show validator${ID} -a)
			echo $HOME
			$BINARY add-genesis-account $($UNARY keys show validator${ID} -a) 1000nametoken,100000000stake
			echo 'CONF'
			$UNARY config chain-id SO3Hlq
		    $UNARY config output json
		    $UNARY config indent true
		    $UNARY config trust-node true

	*/
	logg.Printf("STATE:\n%#+v\n", stateJSON)

	var genesisState map[string]json.RawMessage
	err := app.cdc.UnmarshalJSON(stateJSON, &genesisState)
	if err != nil {
		panic(err)
	}

	//for _, acc := range genesisState.Accounts {
	//	acc.AccountNumber = app.accountKeeper.GetNextAccountNumber(ctx)
	//	app.accountKeeper.SetAccount(ctx, acc)
	//}

	//auth.InitGenesis(ctx, app.accountKeeper, app.feeCollectionKeeper, genesisState.AuthData)
	//bank.InitGenesis(ctx, app.bankKeeper, genesisState.BankData)
	//return abci.ResponseInitChain{}
	return app.mm.InitGenesis(ctx, genesisState)
}

func (app *randApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}
func (app *randApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}
func (app *randApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keyMain)
}

// ExportAppStateAndValidators does the things
func (app *randApp) ExportAppStateAndValidators() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
	ctx := app.NewContext(true, abci.Header{})
	var accounts []*auth.BaseAccount

	appendAccountsFn := func(acc auth.Account) bool {
		account := &auth.BaseAccount{
			Address: acc.GetAddress(),
			Coins:   acc.GetCoins(),
		}

		accounts = append(accounts, account)
		return false
	}

	app.accountKeeper.IterateAccounts(ctx, appendAccountsFn)

	genState := randapp.GenesisState{
		Accounts: accounts,
		AuthData: auth.DefaultGenesisState(),
		BankData: bank.DefaultGenesisState(),
	}

	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}

	return appState, validators, err
}

// MakeCodec generates the necessary codecs for Amino
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}
