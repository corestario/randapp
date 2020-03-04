package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/corestario/randapp/x/randapp"
	"github.com/corestario/randapp/x/randapp/config"
	"github.com/corestario/randapp/x/randapp/metrics"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	reseeding "github.com/cosmos/modules/incubator/reseeding"
)

const appName = "randapp"

var (
	// default home directories for the application CLI
	DefaultCLIHome = os.ExpandEnv("$HOME/../rcli")

	// DefaultNodeHome sets the folder where the applcation data and configuration will be stored
	DefaultNodeHome = os.ExpandEnv("$HOME/")

	// ModuleBasicManager is in charge of setting up basic module elemnets
	ModuleBasics = module.NewBasicManager(
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		params.AppModuleBasic{},
		staking.AppModuleBasic{},
		distr.AppModuleBasic{},
		slashing.AppModuleBasic{},
		supply.AppModuleBasic{},
		reseeding.AppModule{},

		randapp.AppModule{},
	)

	// account permissions
	maccPerms = map[string][]string{
		auth.FeeCollectorName:     nil,
		distr.ModuleName:          nil,
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		reseeding.ModuleName:      nil,
	}
)

// MakeCodec generates the necessary codecs for Amino
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

type randApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	// Keys to access the substores
	keyMain      *sdk.KVStoreKey
	keyAccount   *sdk.KVStoreKey
	keySupply    *sdk.KVStoreKey
	keyStaking   *sdk.KVStoreKey
	tkeyStaking  *sdk.TransientStoreKey
	keyDistr     *sdk.KVStoreKey
	tkeyDistr    *sdk.TransientStoreKey
	keyNFT       *sdk.KVStoreKey
	keyParams    *sdk.KVStoreKey
	tkeyParams   *sdk.TransientStoreKey
	keySlashing  *sdk.KVStoreKey
	keyReseeding *sdk.KVStoreKey

	// Keepers
	accountKeeper   auth.AccountKeeper
	bankKeeper      bank.Keeper
	supplyKeeper    supply.Keeper
	stakingKeeper   staking.Keeper
	slashingKeeper  slashing.Keeper
	distrKeeper     distr.Keeper
	paramsKeeper    params.Keeper
	reseedingKeeper reseeding.Keeper

	randKeeper *randapp.Keeper

	keyPubKeys            *sdk.KVStoreKey
	keyDeals              *sdk.KVStoreKey
	keyResponses          *sdk.KVStoreKey
	keyJustifications     *sdk.KVStoreKey
	keyCommits            *sdk.KVStoreKey
	keyComplaints         *sdk.KVStoreKey
	keyReconstructCommits *sdk.KVStoreKey

	// Module Manager
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

		keyMain:      sdk.NewKVStoreKey(bam.MainStoreKey),
		keyAccount:   sdk.NewKVStoreKey(auth.StoreKey),
		keySupply:    sdk.NewKVStoreKey(supply.StoreKey),
		keyStaking:   sdk.NewKVStoreKey(staking.StoreKey),
		tkeyStaking:  sdk.NewTransientStoreKey(staking.TStoreKey),
		keyDistr:     sdk.NewKVStoreKey(distr.StoreKey),
		tkeyDistr:    sdk.NewTransientStoreKey("transient_" + distr.ModuleName),
		keyParams:    sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams:   sdk.NewTransientStoreKey(params.TStoreKey),
		keySlashing:  sdk.NewKVStoreKey(slashing.StoreKey),
		keyReseeding: sdk.NewKVStoreKey(reseeding.StoreKey),

		keyPubKeys:            sdk.NewKVStoreKey("pub_keys"),
		keyDeals:              sdk.NewKVStoreKey("deals"),
		keyResponses:          sdk.NewKVStoreKey("responses"),
		keyJustifications:     sdk.NewKVStoreKey("justifications"),
		keyCommits:            sdk.NewKVStoreKey("commits"),
		keyComplaints:         sdk.NewKVStoreKey("complaints"),
		keyReconstructCommits: sdk.NewKVStoreKey("reconstruct_commits"),
	}

	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams)
	// Set specific supspaces
	authSubspace := app.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := app.paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
	distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)

	// The AccountKeeper handles address -> account lookups
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		app.keyAccount,
		authSubspace,
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		bankSubspace,
		app.ModuleAccountAddrs(),
	)

	app.supplyKeeper = supply.NewKeeper(app.cdc, app.keySupply, app.accountKeeper,
		app.bankKeeper, maccPerms)

	// The staking keeper
	// The staking keeper
	stakingKeeper := staking.NewKeeper(
		app.cdc,
		app.keyStaking,
		app.supplyKeeper,
		stakingSubspace,
	)

	app.distrKeeper = distr.NewKeeper(
		app.cdc,
		app.keyDistr,
		distrSubspace,
		&stakingKeeper,
		app.supplyKeeper,
		auth.FeeCollectorName,
		nil,
	)

	app.slashingKeeper = slashing.NewKeeper(
		app.cdc,
		app.keySlashing,
		&stakingKeeper,
		slashingSubspace,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(
			app.distrKeeper.Hooks(),
			app.slashingKeeper.Hooks()),
	)

	srvCfg := ReadSrvConfig()
	fmt.Printf("Server Config: \n %+v \n", srvCfg)

	app.reseedingKeeper = reseeding.NewKeeper(cdc, app.keyReseeding)

	reseedingModule := reseeding.NewAppModule(app.reseedingKeeper, app.stakingKeeper)

	app.randKeeper = randapp.NewKeeper(
		app.bankKeeper,
		app.stakingKeeper,
		app.distrKeeper,

		app.keyPubKeys,
		app.keyDeals,
		app.keyResponses,
		app.keyJustifications,
		app.keyCommits,
		app.keyComplaints,
		app.keyReconstructCommits,
		app.cdc,
		srvCfg,
		metrics.NewPrometheusMsgMetrics(appName),
	)

	app.mm = module.NewManager(
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		distr.NewAppModule(app.distrKeeper, app.accountKeeper, app.supplyKeeper, app.stakingKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.supplyKeeper),

		randapp.NewAppModule(app.randKeeper, app.bankKeeper),
		reseedingModule,
	)

	app.mm.SetOrderBeginBlockers(distr.ModuleName, slashing.ModuleName)
	app.mm.SetOrderEndBlockers(staking.ModuleName, reseeding.ModuleName)

	// Sets the order of Genesis - Order matters, genutil is to always come last
	app.mm.SetOrderInitGenesis(
		distr.ModuleName,
		staking.ModuleName,
		auth.ModuleName,
		bank.ModuleName,
		slashing.ModuleName,

		randapp.ModuleName,
		reseeding.ModuleName,

		genutil.ModuleName,
	)

	// register all module routes and module queriers
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// The AnteHandler handles signature verification and transaction pre-processing
	app.SetAnteHandler(
		auth.NewAnteHandler(
			app.accountKeeper, app.supplyKeeper, auth.DefaultSigVerificationGasConsumer,
		),
	)

	app.MountStores(
		app.keyMain,
		app.keyAccount,
		app.keySupply,
		app.keyStaking,
		app.tkeyStaking,
		app.keyDistr,
		app.tkeyDistr,
		app.keySlashing,
		app.keyParams,
		app.tkeyParams,
		app.keyReseeding,

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

// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState map[string]json.RawMessage

func NewDefaultGenesisState() GenesisState {
	return ModuleBasics.DefaultGenesis()
}

func (app *randApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	err := app.cdc.UnmarshalJSON(req.AppStateBytes, &genesisState)
	if err != nil {
		panic(err)
	}
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

func (app *randApp) ExportAppStateAndValidators(forZeroHeight bool, jailWhiteList []string) (appState json.RawMessage,
	validators []tmtypes.GenesisValidator,
	err error,
) {

	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})

	genState := app.mm.ExportGenesis(ctx)
	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}

	validators = staking.WriteValidators(ctx, app.stakingKeeper)

	return appState, validators, nil
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *randApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

func ReadSrvConfig() *config.RAServerConfig {
	var cfg *config.RAServerConfig
	vCfg := viper.New()
	vCfg.SetConfigName("server")
	vCfg.AddConfigPath(DefaultNodeHome + "/config")
	err := vCfg.ReadInConfig()
	if err != nil {
		fmt.Println("ERROR: server config file not found, error:", err)
		return config.DefaultRAServerConfig()
	}
	fmt.Println(vCfg.GetString("maximum_beneficiary_commission"))
	err = vCfg.Unmarshal(&cfg)
	if err != nil {
		fmt.Println("ERROR: could not unmarshal server config file, error:", err)
		return config.DefaultRAServerConfig()
	}
	return cfg
}
