package e2e_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/types"
)

// Tests that we can set a value and retrieve it.
func TestApp_Tx(t *testing.T) {
	t.Helper()
	testNode(t, func(t *testing.T, node e2e.Node) {
		t.Helper()
		client, err := node.Client()
		require.NoError(t, err)

		// Generate a random value, to prevent duplicate tx errors when
		// manually running the test multiple times for a testnet.
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		bz := make([]byte, 32)
		_, err = r.Read(bz)
		require.NoError(t, err)

		key := fmt.Sprintf("testapp-tx-%v", node.Name)
		value := hex.EncodeToString(bz)
		tx := types.Tx(fmt.Sprintf("%v=%v", key, value))

		res, err := client.BroadcastTxSync(ctx, tx)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Zero(t, res.Code)

		hash := tx.Hash()
		require.Equal(t, res.Hash, cmtbytes.HexBytes(hash))
		waitTime := 30 * time.Second
		require.Eventuallyf(t, func() bool {
			txResp, err := client.Tx(ctx, hash, false)
			return err == nil && bytes.Equal(txResp.Tx, tx)
		}, waitTime, time.Second,
			"submitted tx wasn't committed after %v", waitTime,
		)

		// NOTE: we don't test abci query of the light client
		if node.Mode == e2e.ModeLight {
			return
		}

		abciResp, err := client.ABCIQuery(ctx, "", []byte(key))
		require.NoError(t, err)
		assert.Equal(t, key, string(abciResp.Response.Key))
		assert.Equal(t, value, string(abciResp.Response.Value))
	})
}
