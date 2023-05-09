package keeper

import (
	"math/big"
	"sidechain/x/devearn/types"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ethereum/go-ethereum/common"
)

// DistributeRewards transfers the allocated rewards to the participants
//   - allocates the amount to be distributed from the inflation pool
//   - distributes the rewards to all participants
//   - deletes all gas meters
//   - updates the remaining epochs of each incentive
//   - sets the cumulative totalGas to zero
func (k Keeper) DistributeRewards(ctx sdk.Context) error {
	logger := k.Logger(ctx)
	devEarnGasMeters := make(map[string]uint64)
	devEarnRewardReceivers := make(map[string]string)
	k.IterateDevEarnInfos(ctx, func(devEarnInfo types.DevEarnInfo) (stop bool) {
		devEarnGasMeters[devEarnInfo.GetContract()] = devEarnInfo.GetGasMeter()
		devEarnRewardReceivers[devEarnInfo.GetContract()] = devEarnInfo.GetOwnerAddress()
		devEarnInfo.Epochs--

		// Update dev_earn info and reset its total gas count. Remove dev_earn info if it
		// has no remaining epochs left.
		if devEarnInfo.IsActive() {
			devEarnInfo.GasMeter = 0
			k.SetDevEarnInfo(ctx, devEarnInfo)
		} else {
			k.DeleteDevEarnInfo(ctx, devEarnInfo)
			logger.Info(
				"devEarn finalized",
				"contract", devEarnInfo.Contract,
			)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDistributeRewards,
				sdk.NewAttribute(types.AttributeKeyContract, devEarnInfo.Contract),
				sdk.NewAttribute(
					types.AttributeKeyEpochs,
					strconv.FormatUint(uint64(devEarnInfo.Epochs), 10),
				),
			),
		)
		return false
	})
	k.SendReward(ctx, devEarnGasMeters, devEarnRewardReceivers)

	return nil
}

func (k Keeper) SendReward(
	ctx sdk.Context,
	devEarnGasMeters map[string]uint64,
	devEarnRewardReceivers map[string]string,
) (rewards sdk.Coins, count uint64) {
	logger := k.Logger(ctx)
	var totalGasDec sdk.Dec = sdk.NewDec(0)
	// Check if participants spent gas on interacting with incentive
	for _, gasMeter := range devEarnGasMeters {
		totalGasDec = totalGasDec.Add(sdk.NewDecFromBigInt(new(big.Int).SetUint64(gasMeter)))
	}
	if totalGasDec.Equal(sdk.NewDec(0)) {
		logger.Debug(
			"no gas spent on dev earn during epoch",
		)
		return sdk.Coins{}, 0
	}
	mouduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	denom, err := sdk.GetBaseDenom()
	if err != nil {
		logger.Debug("could not get the denom of smallest unit registered", "error", err.Error())
	}
	totalReward := k.bankKeeper.GetBalance(ctx, mouduleAddr, denom)
	for contract, gasmeter := range devEarnGasMeters {
		cumulativeGas := sdk.NewDecFromBigInt(new(big.Int).SetUint64(gasmeter))
		gasRatio := cumulativeGas.Quo(totalGasDec)
		reward := gasRatio.MulInt(totalReward.Amount.Quo(sdk.NewInt(2)))

		tvlRatio, tvlErr := k.TvlReward(ctx, contract)
		if tvlErr != nil {
			logger.Debug("could not get tvl ratio", "error", tvlErr.Error())
		}
		reward.Add(tvlRatio.MulInt(totalReward.Amount.Quo(sdk.NewInt(2))))

		if !reward.IsPositive() {
			continue
		}
		coin := sdk.Coin{Denom: denom, Amount: reward.TruncateInt()}
		coins := sdk.Coins{}
		coins = coins.Add(coin)

		participant := common.HexToAddress(devEarnRewardReceivers[contract])
		err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			sdk.AccAddress(participant.Bytes()),
			coins,
		)
		if err != nil {
			logger.Debug("failed to distribute developer's rewards",
				"address", devEarnRewardReceivers[contract],
				"allocation", coins.String(),
				"contract_addr", contract,
				"error", err.Error(),
			)
		}
	}
	return rewards, count
}

// TvlReward function calculates TVL rewards using native token balance
// TODO: Add support for erc20 tokens value estimation
func (k Keeper) TvlReward(ctx sdk.Context, contractAddress string) (sdk.Dec, error) {
	// Query total supply of native token
	totalDenomSupply, _, err := k.bankKeeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{
		Key:        nil,
		Offset:     0,
		Limit:      100,
		CountTotal: false,
		Reverse:    false,
	})
	if err != nil {
		ctx.Logger().Error("get total supply err happen, err :", err)
		return sdk.NewDec(0), err
	}
	denom, err := sdk.GetBaseDenom()
	if err != nil {
		ctx.Logger().Error("get base denom err happen, err :", err)
		return sdk.NewDec(0), err
	}

	totalSupply := sdk.NewDecFromBigInt(new(big.Int).SetUint64(totalDenomSupply.AmountOf(denom).Uint64()))
	bal := k.bankKeeper.GetBalance(ctx, sdk.AccAddress(contractAddress), denom)
	balD := sdk.NewDecFromBigInt(new(big.Int).SetUint64(bal.Amount.Uint64()))
	tvlRatio := balD.Quo(totalSupply)

	return tvlRatio, nil
}
