package keeper

import (
	"github.com/sideprotocol/sidechain/x/router/types"
)

var _ types.QueryServer = Keeper{}
