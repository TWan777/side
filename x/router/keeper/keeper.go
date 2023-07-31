package keeper

import (
	"fmt"
	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	//"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/tendermint/tendermint/libs/log"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"
	atomicswaptypes "github.com/sideprotocol/ibcswap/v6/modules/apps/100-atomic-swap/types"
	interchainswaptypes "github.com/sideprotocol/ibcswap/v6/modules/apps/101-interchain-swap/types"
	"github.com/sideprotocol/sidechain/x/router/types"
)

var (
	// DefaultPacketTimeoutHeight is the timeout height following IBC defaults
	DefaultPacketTimeoutHeight = clienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: 0,
	}

	// DefaultForwardTransferPacketTimeoutTimestamp is the timeout timestamp following IBC defaults
	DefaultForwardTransferPacketTimeoutTimestamp = time.Duration(atomicswaptypes.DefaultRelativePacketTimeoutTimestamp) * time.Nanosecond

	// DefaultRefundTransferPacketTimeoutTimestamp is a 28-day timeout for refund packets since funds are stuck in router module otherwise.
	DefaultRefundTransferPacketTimeoutTimestamp = 28 * 24 * time.Hour
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		channelKeeper        types.ChannelKeeper
		portKeeper           types.PortKeeper
		scopedKeeper         exported.ScopedKeeper
		ics4Wrapper          porttypes.ICS4Wrapper
		atomicswapKeeper     types.AtomicSwapKeeper
		interchainswapKeeper types.InterchainSwapKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper exported.ScopedKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
	atomicswapKeeper types.AtomicSwapKeeper,
	interchainswapKeeper types.InterchainSwapKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,

		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
		scopedKeeper:  scopedKeeper,
	}
}

// ----------------------------------------------------------------------------
// IBC Keeper Logic
// ----------------------------------------------------------------------------

// ChanCloseInit defines a wrapper function for the channel Keeper's function.
func (k Keeper) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	capName := host.ChannelCapabilityPath(portID, channelID)
	chanCap, ok := k.scopedKeeper.GetCapability(ctx, capName)
	if !ok {
		return sdkerrors.Wrapf(channeltypes.ErrChannelCapabilityNotFound, "could not retrieve channel capability at: %s", capName)
	}
	return k.channelKeeper.ChanCloseInit(ctx, portID, channelID, chanCap)
}

// IsBound checks if the IBC app module is already bound to the desired port
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

// BindPort defines a wrapper function for the port Keeper's function in
// order to expose it to module's InitGenesis function
func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	cap := k.portKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, cap, host.PortPath(portID))
}

// GetPort returns the portID for the IBC app module. Used in ExportGenesis
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.PortKey))
}

// SetPort sets the portID for the IBC app module. Used in InitGenesis
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PortKey, []byte(portID))
}

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability allows the IBC app module to claim a capability that core IBC
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetAtomicswapKeeper(atomicswapKeeper types.AtomicSwapKeeper) {
	k.atomicswapKeeper = atomicswapKeeper
}

func (k Keeper) SetInterchainswapKeeper(interchainswapKeeper types.InterchainSwapKeeper) {
	k.interchainswapKeeper = interchainswapKeeper
}

func (k *Keeper) WriteAcknowledgementForForwardedPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	//data atomicswaptypes.AtomicSwapPacketData,
	routingPacket *types.RouterPacketData,
	ack channeltypes.Acknowledgement,
) error {
	// Lookup module by channel capability
	_, cap, err := k.channelKeeper.LookupModuleByChannel(ctx, routingPacket.RefundPortId, routingPacket.RefundChannelId)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// for forwarded packets, the funds were moved into an escrow account if the denom originated on this chain.
	// On an ack error or timeout on a forwarded packet, the funds in the escrow account
	// should be moved to the other escrow account on the other side or burned.
	if !ack.Success() {
		// If this packet is non-refundable due to some action that took place between the initial ibc transfer and the forward
		// we write a successful ack containing details on what happened regardless of ack error or timeout
		if routingPacket.Nonrefundable {
			ackResult := fmt.Sprintf("packet forward failed after point of no return: %s", ack.GetError())
			newAck := channeltypes.NewResultAcknowledgement([]byte(ackResult))

			return k.ics4Wrapper.WriteAcknowledgement(ctx, cap, channeltypes.Packet{
				Data:               routingPacket.PacketData,
				Sequence:           routingPacket.RefundSequence,
				SourcePort:         routingPacket.PacketSrcPortId,
				SourceChannel:      routingPacket.PacketSrcChannelId,
				DestinationPort:    routingPacket.RefundPortId,
				DestinationChannel: routingPacket.RefundChannelId,
				TimeoutHeight:      clienttypes.MustParseHeight(routingPacket.PacketTimeoutHeight),
				TimeoutTimestamp:   routingPacket.PacketTimeoutTimestamp,
			}, newAck)
		}

		// atomicswaptypes.SenderChainIsSource(
		// 	packet.SourcePort, packet.SourceChannel, data.Path,
		// )

		// if transfertypes.SenderChainIsSource(packet.SourcePort, packet.SourceChannel, fullDenomPath) {
		// 	// funds were moved to escrow account for transfer, so they need to either:
		// 	// - move to the other escrow account, in the case of native denom
		// 	// - burn

		// 	amount, ok := sdk.NewIntFromString(data.Amount)
		// 	if !ok {
		// 		return fmt.Errorf("failed to parse amount from packet data for forward refund: %s", data.Amount)
		// 	}
		// 	denomTrace := transfertypes.ParseDenomTrace(fullDenomPath)
		// 	token := sdk.NewCoin(denomTrace.IBCDenom(), amount)

		// 	escrowAddress := transfertypes.GetEscrowAddress(packet.SourcePort, packet.SourceChannel)

		// 	if transfertypes.SenderChainIsSource(inFlightPacket.RefundPortId, inFlightPacket.RefundChannelId, fullDenomPath) {
		// 		// transfer funds from escrow account for forwarded packet to escrow account going back for refund.

		// 		refundEscrowAddress := transfertypes.GetEscrowAddress(inFlightPacket.RefundPortId, inFlightPacket.RefundChannelId)

		// 		if err := k.bankKeeper.SendCoins(
		// 			ctx, escrowAddress, refundEscrowAddress, sdk.NewCoins(token),
		// 		); err != nil {
		// 			return fmt.Errorf("failed to send coins from escrow account to refund escrow account: %w", err)
		// 		}
		// 	} else {
		// 		// transfer the coins from the escrow account to the module account and burn them.

		// 		if err := k.bankKeeper.SendCoinsFromAccountToModule(
		// 			ctx, escrowAddress, transfertypes.ModuleName, sdk.NewCoins(token),
		// 		); err != nil {
		// 			return fmt.Errorf("failed to send coins from escrow to module account for burn: %w", err)
		// 		}

		// 		if err := k.bankKeeper.BurnCoins(
		// 			ctx, transfertypes.ModuleName, sdk.NewCoins(token),
		// 		); err != nil {
		// 			// NOTE: should not happen as the module account was
		// 			// retrieved on the step above and it has enough balace
		// 			// to burn.
		// 			panic(fmt.Sprintf("cannot burn coins after a successful send from escrow account to module account: %v", err))
		// 		}
		// 	}
		// }
	}

	return k.ics4Wrapper.WriteAcknowledgement(ctx, cap, channeltypes.Packet{
		Data:               routingPacket.PacketData,
		Sequence:           routingPacket.RefundSequence,
		SourcePort:         routingPacket.PacketSrcPortId,
		SourceChannel:      routingPacket.PacketSrcChannelId,
		DestinationPort:    routingPacket.RefundPortId,
		DestinationChannel: routingPacket.RefundChannelId,
		TimeoutHeight:      clienttypes.MustParseHeight(routingPacket.PacketTimeoutHeight),
		TimeoutTimestamp:   routingPacket.PacketTimeoutTimestamp,
	}, ack)
}

// func (k *Keeper) ForwardTransferPacket(
// 	ctx sdk.Context,
// 	routingPacket *types.RouterPacketData,
// 	srcPacket channeltypes.Packet,
// 	srcPacketSender string,
// 	receiver string,
// 	metadata *types.ForwardMetadata,
// 	token sdk.Coin,
// 	maxRetries uint8,
// 	timeout time.Duration,
// 	labels []metrics.Label,
// 	nonrefundable bool,
// ) error {
// 	var err error
// 	feeAmount := sdk.NewDecFromInt(token.Amount).Mul(k.GetFeePercentage(ctx)).RoundInt()
// 	packetAmount := token.Amount.Sub(feeAmount)
// 	feeCoins := sdk.Coins{sdk.NewCoin(token.Denom, feeAmount)}
// 	packetCoin := sdk.NewCoin(token.Denom, packetAmount)

// 	// pay fees
// 	if feeAmount.IsPositive() {
// 		hostAccAddr, err := sdk.AccAddressFromBech32(receiver)
// 		if err != nil {
// 			return err
// 		}
// 		err = k.distrKeeper.FundCommunityPool(ctx, feeCoins, hostAccAddr)
// 		if err != nil {
// 			k.Logger(ctx).Error("packetForwardMiddleware error funding community pool",
// 				"error", err,
// 			)
// 			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
// 		}
// 	}

// 	memo := ""

// 	// set memo for next transfer with next from this transfer.
// 	if metadata.Next != nil {
// 		memo = *metadata.Next
// 	}

// 	msgTransfer := transfertypes.NewMsgTransfer(
// 		metadata.Port,
// 		metadata.Channel,
// 		packetCoin,
// 		receiver,
// 		metadata.Receiver,
// 		DefaultPacketTimeoutHeight,
// 		uint64(ctx.BlockTime().UnixNano())+uint64(timeout.Nanoseconds()),
// 		memo,
// 	)

// 	k.Logger(ctx).Debug("packetForwardMiddleware ForwardTransferPacket",
// 		"port", metadata.Port, "channel", metadata.Channel,
// 		"sender", receiver, "receiver", metadata.Receiver,
// 		"amount", packetCoin.Amount.String(), "denom", packetCoin.Denom,
// 	)

// 	// send tokens to destination
// 	res, err := k.transferKeeper.Transfer(
// 		sdk.WrapSDKContext(ctx),
// 		msgTransfer,
// 	)
// 	if err != nil {
// 		k.Logger(ctx).Error("packetForwardMiddleware ForwardTransferPacket error",
// 			"port", metadata.Port, "channel", metadata.Channel,
// 			"sender", receiver, "receiver", metadata.Receiver,
// 			"amount", packetCoin.Amount.String(), "denom", packetCoin.Denom,
// 			"error", err,
// 		)
// 		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
// 	}

// 	// Store the following information in keeper:
// 	// key - information about forwarded packet: src_channel (parsedReceiver.Channel), src_port (parsedReceiver.Port), sequence
// 	// value - information about original packet for refunding if necessary: retries, srcPacketSender, srcPacket.DestinationChannel, srcPacket.DestinationPort

// 	if inFlightPacket == nil {
// 		inFlightPacket = &types.InFlightPacket{
// 			PacketData:            srcPacket.Data,
// 			OriginalSenderAddress: srcPacketSender,
// 			RefundChannelId:       srcPacket.DestinationChannel,
// 			RefundPortId:          srcPacket.DestinationPort,
// 			RefundSequence:        srcPacket.Sequence,
// 			PacketSrcPortId:       srcPacket.SourcePort,
// 			PacketSrcChannelId:    srcPacket.SourceChannel,

// 			PacketTimeoutTimestamp: srcPacket.TimeoutTimestamp,
// 			PacketTimeoutHeight:    srcPacket.TimeoutHeight.String(),

// 			RetriesRemaining: int32(maxRetries),
// 			Timeout:          uint64(timeout.Nanoseconds()),
// 			Nonrefundable:    nonrefundable,
// 		}
// 	} else {
// 		inFlightPacket.RetriesRemaining--
// 	}

// 	key := types.RefundPacketKey(metadata.Channel, metadata.Port, res.Sequence)
// 	store := ctx.KVStore(k.storeKey)
// 	bz := k.cdc.MustMarshal(inFlightPacket)
// 	store.Set(key, bz)

// 	defer func() {
// 		telemetry.SetGaugeWithLabels(
// 			[]string{"tx", "msg", "ibc", "transfer"},
// 			float32(token.Amount.Int64()),
// 			[]metrics.Label{telemetry.NewLabel(coretypes.LabelDenom, token.Denom)},
// 		)

// 		telemetry.IncrCounterWithLabels(
// 			[]string{"ibc", types.ModuleName, "send"},
// 			1,
// 			labels,
// 		)
// 	}()
// 	return nil
// }

func (k *Keeper) ForwardAtomicSwapPacket(
	ctx sdk.Context,
	routingPacket *types.RouterPacketData,
	srcPacket channeltypes.Packet,
	data atomicswaptypes.AtomicSwapPacketData,
	metadata *types.ForwardMetadata,
	maxRetries uint8,
	timeout time.Duration,
	labels []metrics.Label,
	nonrefundable bool,
) error {
	// var err error
	// feeAmount := sdk.NewDecFromInt(token.Amount).Mul(k.GetFeePercentage(ctx)).RoundInt()
	// packetAmount := token.Amount.Sub(feeAmount)
	// feeCoins := sdk.Coins{sdk.NewCoin(token.Denom, feeAmount)}
	// packetCoin := sdk.NewCoin(token.Denom, packetAmount)

	// pay fees
	// if feeAmount.IsPositive() {
	// 	hostAccAddr, err := sdk.AccAddressFromBech32(receiver)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = k.distrKeeper.FundCommunityPool(ctx, feeCoins, hostAccAddr)
	// 	if err != nil {
	// 		k.Logger(ctx).Error("packetForwardMiddleware error funding community pool",
	// 			"error", err,
	// 		)
	// 		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	// 	}
	// }

	memo := ""

	// set memo for next transfer with next from this transfer.
	if metadata.Next != nil {
		memo = *metadata.Next
	}
	data.Memo = memo

	// msgTransfer := transfertypes.NewMsgTransfer(
	// 	metadata.Port,
	// 	metadata.Channel,
	// 	packetCoin,
	// 	receiver,
	// 	metadata.Receiver,
	// 	DefaultTransferPacketTimeoutHeight,
	// 	uint64(ctx.BlockTime().UnixNano())+uint64(timeout.Nanoseconds()),
	// 	memo,
	// )

	k.Logger(ctx).Debug("packetForwardMiddleware ForwardTransferPacket",
		"port", metadata.Port, "channel", metadata.Channel,
		//"sender", receiver, "receiver", metadata.Receiver,
		//"amount", packetCoin.Amount.String(), "denom", packetCoin.Denom,
	)

	sequence, err := k.atomicswapKeeper.SendSwapPacket(
		ctx, metadata.Port, metadata.Channel, DefaultPacketTimeoutHeight, 0, data,
	)
	// send tokens to destination
	// res, err := k.transferKeeper.Transfer(
	// 	sdk.WrapSDKContext(ctx),
	// 	msgTransfer,
	// )
	if err != nil {
		k.Logger(ctx).Error("packetForwardMiddleware ForwardTransferPacket error",
			"port", metadata.Port, "channel", metadata.Channel,
			//"sender", receiver, "receiver", metadata.Receiver,
			//"amount", packetCoin.Amount.String(), "denom", packetCoin.Denom,
			"error", err,
		)
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	// Store the following information in keeper:
	// key - information about forwarded packet: src_channel (parsedReceiver.Channel), src_port (parsedReceiver.Port), sequence
	// value - information about original packet for refunding if necessary: retries, srcPacketSender, srcPacket.DestinationChannel, srcPacket.DestinationPort

	if routingPacket == nil {
		routingPacket = &types.RouterPacketData{
			PacketData: srcPacket.Data,
			//OriginalSenderAddress: srcPacketSender,
			RefundChannelId:    srcPacket.DestinationChannel,
			RefundPortId:       srcPacket.DestinationPort,
			RefundSequence:     srcPacket.Sequence,
			PacketSrcPortId:    srcPacket.SourcePort,
			PacketSrcChannelId: srcPacket.SourceChannel,

			PacketTimeoutTimestamp: srcPacket.TimeoutTimestamp,
			PacketTimeoutHeight:    srcPacket.TimeoutHeight.String(),

			RetriesRemaining: int32(maxRetries),
			Timeout:          uint64(timeout.Nanoseconds()),
			Nonrefundable:    nonrefundable,
		}
	} else {
		routingPacket.RetriesRemaining--
	}

	key := types.RefundPacketKey(metadata.Channel, metadata.Port, *sequence)
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(routingPacket)
	store.Set(key, bz)

	// defer func() {
	// 	telemetry.SetGaugeWithLabels(
	// 		[]string{"tx", "msg", "ibc", "atomicswap"},
	// 		float32(token.Amount.Int64()),
	// 		[]metrics.Label{telemetry.NewLabel(coretypes.LabelDenom, token.Denom)},
	// 	)

	// 	telemetry.IncrCounterWithLabels(
	// 		[]string{"ibc", types.ModuleName, "send"},
	// 		1,
	// 		labels,
	// 	)
	// }()
	return nil
}

func (k *Keeper) ForwardInterchainSwapPacket(
	ctx sdk.Context,
	routingPacket *types.RouterPacketData,
	srcPacket channeltypes.Packet,
	data interchainswaptypes.IBCSwapPacketData,
	metadata *types.ForwardMetadata,
	maxRetries uint8,
	timeout time.Duration,
	labels []metrics.Label,
	nonrefundable bool,
) error {
	var err error

	memo := ""

	// set memo for next transfer with next from this transfer.
	if metadata.Next != nil {
		memo = *metadata.Next
	}

	data.Memo = memo

	k.Logger(ctx).Debug("packetForwardMiddleware ForwardTransferPacket",
		"port", metadata.Port, "channel", metadata.Channel,
		//"sender", receiver, "receiver", metadata.Receiver,
		//"amount", packetCoin.Amount.String(), "denom", packetCoin.Denom,
	)

	// send tokens to destination
	sequence, err := k.interchainswapKeeper.SendIBCSwapPacket(
		ctx,
		metadata.Port,
		metadata.Channel,
		DefaultPacketTimeoutHeight,
		0,
		data,
	)
	if err != nil {
		k.Logger(ctx).Error("packetForwardMiddleware ForwardTransferPacket error",
			"port", metadata.Port, "channel", metadata.Channel,
			//"sender", receiver, "receiver", metadata.Receiver,
			//"amount", packetCoin.Amount.String(), "denom", packetCoin.Denom,
			"error", err,
		)
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	// Store the following information in keeper:
	// key - information about forwarded packet: src_channel (parsedReceiver.Channel), src_port (parsedReceiver.Port), sequence
	// value - information about original packet for refunding if necessary: retries, srcPacketSender, srcPacket.DestinationChannel, srcPacket.DestinationPort

	if routingPacket == nil {
		routingPacket = &types.RouterPacketData{
			PacketData: srcPacket.Data,
			//OriginalSenderAddress: srcPacketSender,
			RefundChannelId:    srcPacket.DestinationChannel,
			RefundPortId:       srcPacket.DestinationPort,
			RefundSequence:     srcPacket.Sequence,
			PacketSrcPortId:    srcPacket.SourcePort,
			PacketSrcChannelId: srcPacket.SourceChannel,

			PacketTimeoutTimestamp: srcPacket.TimeoutTimestamp,
			PacketTimeoutHeight:    srcPacket.TimeoutHeight.String(),

			RetriesRemaining: int32(maxRetries),
			Timeout:          uint64(timeout.Nanoseconds()),
			Nonrefundable:    nonrefundable,
		}
	} else {
		routingPacket.RetriesRemaining--
	}

	key := types.RefundPacketKey(metadata.Channel, metadata.Port, *sequence)
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(routingPacket)
	store.Set(key, bz)

	// defer func() {
	// 	telemetry.SetGaugeWithLabels(
	// 		[]string{"tx", "msg", "ibc", "transfer"},
	// 		float32(token.Amount.Int64()),
	// 		[]metrics.Label{telemetry.NewLabel(coretypes.LabelDenom, token.Denom)},
	// 	)

	// 	telemetry.IncrCounterWithLabels(
	// 		[]string{"ibc", types.ModuleName, "send"},
	// 		1,
	// 		labels,
	// 	)
	// }()
	return nil
}

// TimeoutShouldRetry returns inFlightPacket and no error if retry should be attempted. Error is returned if IBC refund should occur.
func (k *Keeper) TimeoutShouldRetry(
	ctx sdk.Context,
	packet channeltypes.Packet,
) (*types.RouterPacketData, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.RefundPacketKey(packet.SourceChannel, packet.SourcePort, packet.Sequence)

	if !store.Has(key) {
		// not a forwarded packet, ignore.
		return nil, nil
	}

	bz := store.Get(key)
	var routingPacket types.RouterPacketData
	k.cdc.MustUnmarshal(bz, &routingPacket)

	if routingPacket.RetriesRemaining <= 0 {
		k.Logger(ctx).Error("packetForwardMiddleware reached max retries for packet",
			"key", string(key),
			"original-sender-address", routingPacket.OriginalSenderAddress,
			"refund-channel-id", routingPacket.RefundChannelId,
			"refund-port-id", routingPacket.RefundPortId,
		)
		return &routingPacket, fmt.Errorf("giving up on packet on channel (%s) port (%s) after max retries",
			routingPacket.RefundChannelId, routingPacket.RefundPortId)
	}

	return &routingPacket, nil
}

func (k *Keeper) RetryAtomicSwapTimeout(
	ctx sdk.Context,
	channel, port string,
	data atomicswaptypes.AtomicSwapPacketData,
	routingPacket *types.RouterPacketData,
) error {
	// send transfer again
	metadata := &types.ForwardMetadata{
		//Receiver: data.Receiver,
		Channel: channel,
		Port:    port,
	}

	if data.Memo != "" {
		metadata.Next = &data.Memo
	}

	k.Logger(ctx).Error("packetForwardMiddleware error parsing amount from string for router retry on timeout",
		"original-sender-address", routingPacket.OriginalSenderAddress,
		"refund-channel-id", routingPacket.RefundChannelId,
		"refund-port-id", routingPacket.RefundPortId,
		"retries-remaining", routingPacket.RetriesRemaining,
		//"amount", data.Amount,
	)
	//denom := transfertypes.ParseDenomTrace(data.Denom).IBCDenom()

	//var token = sdk.NewCoin(denom, amount)

	// srcPacket and srcPacketSender are empty because inFlightPacket is non-nil.
	return k.ForwardAtomicSwapPacket(
		ctx,
		routingPacket,
		channeltypes.Packet{},
		data,
		metadata,
		uint8(routingPacket.RetriesRemaining),
		time.Duration(routingPacket.Timeout)*time.Nanosecond,
		nil,
		routingPacket.Nonrefundable,
	)
}

func (k *Keeper) RetryInterchainSwapTimeout(
	ctx sdk.Context,
	channel, port string,
	data interchainswaptypes.IBCSwapPacketData,
	routingPacket *types.RouterPacketData,
) error {
	// send transfer again
	metadata := &types.ForwardMetadata{
		//Receiver: data.Receiver,
		Channel: channel,
		Port:    port,
	}

	if data.Memo != "" {
		metadata.Next = &data.Memo
	}

	k.Logger(ctx).Error("packetForwardMiddleware error parsing amount from string for router retry on timeout",
		"original-sender-address", routingPacket.OriginalSenderAddress,
		"refund-channel-id", routingPacket.RefundChannelId,
		"refund-port-id", routingPacket.RefundPortId,
		"retries-remaining", routingPacket.RetriesRemaining,
		//"amount", data.Amount,
	)
	//denom := transfertypes.ParseDenomTrace(data.Denom).IBCDenom()

	//var token = sdk.NewCoin(denom, amount)

	// srcPacket and srcPacketSender are empty because inFlightPacket is non-nil.
	return k.ForwardInterchainSwapPacket(
		ctx,
		routingPacket,
		channeltypes.Packet{},
		data,
		metadata,
		uint8(routingPacket.RetriesRemaining),
		time.Duration(routingPacket.Timeout)*time.Nanosecond,
		nil,
		routingPacket.Nonrefundable,
	)
}


func (k *Keeper) RemoveRoutingPacket(ctx sdk.Context, packet channeltypes.Packet) {
	store := ctx.KVStore(k.storeKey)
	key := types.RefundPacketKey(packet.SourceChannel, packet.SourcePort, packet.Sequence)
	if !store.Has(key) {
		// not a forwarded packet, ignore.
		return
	}

	// done with packet key now, delete.
	store.Delete(key)
}

// GetAndClearInFlightPacket will fetch an InFlightPacket from the store, remove it if it exists, and return it.
func (k *Keeper) GetAndClearRoutingPacket(
	ctx sdk.Context,
	channel string,
	port string,
	sequence uint64,
) *types.RouterPacketData {
	store := ctx.KVStore(k.storeKey)
	key := types.RefundPacketKey(channel, port, sequence)
	if !store.Has(key) {
		// this is either not a forwarded packet, or it is the final destination for the refund.
		return nil
	}

	bz := store.Get(key)

	// done with packet key now, delete.
	store.Delete(key)

	var routingPacket types.RouterPacketData
	k.cdc.MustUnmarshal(bz, &routingPacket)
	return &routingPacket
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (k Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string, sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return k.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper WriteAcknowledgement function.
// ICS29 WriteAcknowledgement is used for asynchronous acknowledgements.
func (k *Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper GetAppVersion function.
func (k *Keeper) GetAppVersion(
	ctx sdk.Context,
	portID,
	channelID string,
) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}

// LookupModuleByChannel wraps ChannelKeeper LookupModuleByChannel function.
func (k *Keeper) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error) {
	return k.channelKeeper.LookupModuleByChannel(ctx, portID, channelID)
}
