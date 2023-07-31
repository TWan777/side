package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/sideprotocol/sidechain/testutil/keeper"
	"github.com/sideprotocol/sidechain/x/router/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.RouterKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
