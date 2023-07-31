package router_test

import (
	"testing"

	keepertest "github.com/sideprotocol/sidechain/testutil/keeper"
	"github.com/sideprotocol/sidechain/testutil/nullify"
	"github.com/sideprotocol/sidechain/x/router"
	"github.com/sideprotocol/sidechain/x/router/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:	types.DefaultParams(),
		PortId: types.PortID,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.RouterKeeper(t)
	router.InitGenesis(ctx, *k, genesisState)
	got := router.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	// this line is used by starport scaffolding # genesis/test/assert
}
