package keeper_test

import (
	"testing"

	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/sideprotocol/side/app"
	sample "github.com/sideprotocol/side/testutil/sample"
	"github.com/sideprotocol/side/x/gmm/types"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	initChain = true
)

const (
	balAlice = 500000000000
	balBob   = 200000000000
	balCarol = 100000000000
)

type KeeperTestSuite struct {
	suite.Suite

	legacyAmino *codec.LegacyAmino
	ctx         sdk.Context
	app         *simapp.App
	msgServer   types.MsgServer
	queryClient types.QueryClient
}

var gmmModuleAddress string

func (suite *KeeperTestSuite) SetupTest() {
	// app := simapp.InitSideTestApp(initChain)
	logger := log.NewTestLogger(suite.T())
	db := db.NewMemDB()

	app, _ := simapp.New(
		logger,
		db,
		nil,
		true,
		nil,
	)
	ctx := app.BaseApp.NewContext(false)
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.MinCommissionRate = sdkmath.LegacyOneDec()
	app.StakingKeeper.SetParams(ctx, stakingtypes.DefaultParams())

	gmmModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.AppCodec().InterfaceRegistry())
	//types.RegisterQueryServer(queryHelper, app.GmmKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.legacyAmino = app.LegacyAmino()
	suite.ctx = ctx
	suite.app = app
	suite.queryClient = queryClient
	//suite.msgServer = keeper.NewMsgServerImpl(app.GmmKeeper)

	// Set Coins
	suite.setupSuiteWithBalances()
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func makeBalance(address string, balance int64, denom string) banktypes.Balance {
	return banktypes.Balance{
		Address: address,
		Coins: sdk.Coins{
			sdk.Coin{
				Denom:  denom,
				Amount: sdkmath.NewInt(balance),
			},
		},
	}
}

func getBankGenesis() *banktypes.GenesisState {
	coins := []banktypes.Balance{
		makeBalance(types.Alice, balAlice, sample.DefaultBondDenom),
		makeBalance(types.Alice, balAlice, sample.USDC),
		makeBalance(types.Alice, balAlice, sample.WBTC),
		makeBalance(types.Alice, balAlice, sample.WDAI),
		makeBalance(types.Alice, balAlice, sample.WUSDT),

		makeBalance(types.Bob, balBob, sample.DefaultBondDenom),
		makeBalance(types.Bob, balBob, sample.USDC),
		makeBalance(types.Bob, balAlice, sample.WBTC),
		makeBalance(types.Bob, balAlice, sample.WDAI),
		makeBalance(types.Bob, balAlice, sample.WUSDT),

		makeBalance(types.Carol, balCarol, sample.DefaultBondDenom),
		makeBalance(types.Carol, balCarol, sample.USDC),
		makeBalance(types.Carol, balAlice, sample.WBTC),
		makeBalance(types.Carol, balAlice, sample.WDAI),
		makeBalance(types.Carol, balAlice, sample.WUSDT),
	}

	params := banktypes.DefaultParams()
	params.DefaultSendEnabled = true
	state := banktypes.NewGenesisState(
		params,
		coins,
		addAll(coins),
		[]banktypes.Metadata{}, []banktypes.SendEnabled{
			{Denom: sample.DefaultBondDenom, Enabled: true},
			{Denom: sample.USDC, Enabled: true},
		})

	return state
}

func addAll(balances []banktypes.Balance) sdk.Coins {
	total := sdk.NewCoins()
	for _, balance := range balances {
		total = total.Add(balance.Coins...)
	}
	return total
}

func (suite *KeeperTestSuite) SetupStableCoinPrices() {
	// prices set for USDT and USDC
	// provider := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	// suite.app.OracleKeeper.SetAssetInfo(suite.ctx, oracletypes.AssetInfo{
	// 	Denom:   "uusdc",
	// 	Display: "USDC",
	// 	Decimal: 6,
	// })
	// suite.app.OracleKeeper.SetAssetInfo(suite.ctx, oracletypes.AssetInfo{
	// 	Denom:   "uusdt",
	// 	Display: "USDT",
	// 	Decimal: 6,
	// })
	// suite.app.OracleKeeper.SetPrice(suite.ctx, oracletypes.Price{
	// 	Asset:     "USDC",
	// 	Price:     sdk.NewDec(1),
	// 	Source:    "elys",
	// 	Provider:  provider.String(),
	// 	Timestamp: uint64(suite.ctx.BlockTime().Unix()),
	// })
	// suite.app.OracleKeeper.SetPrice(suite.ctx, oracletypes.Price{
	// 	Asset:     "USDT",
	// 	Price:     sdk.NewDec(1),
	// 	Source:    "elys",
	// 	Provider:  provider.String(),
	// 	Timestamp: uint64(suite.ctx.BlockTime().Unix()),
	// })
}

// func getStakeGenesis() *stakingtypes.GenesisState {
// 	state := stakingtypes.DefaultGenesisState()
// 	state.Params.BondDenom = simapp.DefaultBondDenom
// 	state.Params.MinCommissionRate = sdk.OneDec()
// 	return state
// }

func (suite *KeeperTestSuite) setupSuiteWithBalances() {
	// suite.app.StakingKeeper.InitGenesis(suite.ctx, getStakeGenesis())
	suite.app.BankKeeper.InitGenesis(suite.ctx, getBankGenesis())
}

// func (suite *KeeperTestSuite) RequireBankBalance(expected int, atAddress string) {
// 	suite.RequireBankBalanceWithDenom(expected, "aside", atAddress)
// }

// func (suite *KeeperTestSuite) RequireBankBalanceWithDenom(expected int, denom string, atAddress string) {
// 	sdkAdd, err := sdk.AccAddressFromBech32(atAddress)
// 	suite.Require().Nil(err, "Failed to parse address: %s", atAddress)
// 	suite.Require().Equal(
// 		int64(expected),
// 		suite.app.BankKeeper.GetBalance(suite.ctx, sdkAdd, denom).Amount.Int64())
// }
