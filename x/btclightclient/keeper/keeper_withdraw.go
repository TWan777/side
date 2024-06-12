package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/btclightclient/types"
)

// GetRequestSeqence returns the request sequence
func (k Keeper) GetRequestSeqence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SequenceKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// IncrementRequestSequence increments the request sequence and returns the new sequence
func (k Keeper) IncrementRequestSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	seq := k.GetRequestSeqence(ctx) + 1
	store.Set(types.SequenceKey, sdk.Uint64ToBigEndian(seq))
	return seq
}

// New signing request
// sender: the address of the sender
// txBytes: the transaction bytes
// vault: the address of the vault, default is empty.
// If empty, the vault will be Bitcoin vault, otherwise it will be Ordinals or Runes vault
func (k Keeper) NewSigningRequest(ctx sdk.Context, sender string, coin sdk.Coin, vault string) *types.BitcoinSigningRequest {

	// create a new bitcoin transaction
	// tx := wire.NewMsgTx(wire.TxVersion)

	// outScript, err := txscript.PayToAddrScript(sender)

	signingRequest := &types.BitcoinSigningRequest{
		Address:      sender,
		TxBytes:      "",
		Status:       types.SigningStatus_SIGNING_STATUS_CREATED,
		Sequence:     k.IncrementRequestSequence(ctx),
		VaultAddress: vault,
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(signingRequest)
	store.Set(types.BtcSigningRequestKey(signingRequest.Sequence), bz)

	return signingRequest
}

// GetSigningRequest returns the signing request
func (k Keeper) GetSigningRequest(ctx sdk.Context, hash string) *types.BitcoinSigningRequest {
	store := ctx.KVStore(k.storeKey)
	var signingRequest types.BitcoinSigningRequest
	// TODO replace the key with the hash
	bz := store.Get(types.BtcSigningRequestKey(1))
	k.cdc.MustUnmarshal(bz, &signingRequest)
	return &signingRequest
}

// SetSigningRequest sets the signing request
func (k Keeper) SetSigningRequest(ctx sdk.Context, txHash string, signingRequest *types.BitcoinSigningRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(signingRequest)
	// TODO replace the key with the hash
	store.Set(types.BtcSigningRequestKey(1), bz)
}

// IterateSigningRequests iterates through all signing requests
func (k Keeper) IterateSigningRequests(ctx sdk.Context, process func(signingRequest types.BitcoinSigningRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.BtcSigningRequestPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var signingRequest types.BitcoinSigningRequest
		k.cdc.MustUnmarshal(iterator.Value(), &signingRequest)
		if process(signingRequest) {
			break
		}
	}
}

// filter SigningRequest by status with pagination
func (k Keeper) FilterSigningRequestsByStatus(ctx sdk.Context, req *types.QuerySigningRequestRequest) []*types.BitcoinSigningRequest {
	var signingRequests []*types.BitcoinSigningRequest
	k.IterateSigningRequests(ctx, func(signingRequest types.BitcoinSigningRequest) (stop bool) {
		if signingRequest.Status == req.Status {
			signingRequests = append(signingRequests, &signingRequest)
		}
		// pagination TODO: limit the number of signing requests
		if len(signingRequests) >= 100 {
			return true
		}
		return false
	})
	return signingRequests
}