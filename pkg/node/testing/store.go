package testing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"github.com/xmtp/xmtpd/pkg/context"
	crdttest "github.com/xmtp/xmtpd/pkg/crdt/testing"
	"github.com/xmtp/xmtpd/pkg/node"
	test "github.com/xmtp/xmtpd/pkg/testing"
)

func TestTopicBootstrap(t *testing.T, storeMaker func(t *testing.T, ctx context.Context) node.NodeStore) {
	topics := []string{"topic1", "topic2", "topic3"}
	ctx := test.NewContext(t)
	store := storeMaker(t, ctx)
	node := NewTestNodeWithNameAndStore(t, ctx, "", store)
	for i, topic := range topics {
		node.PublishRandom(t, topic, i+1)
	}
	res, err := store.Topics()
	require.NoError(t, err)
	require.ElementsMatch(t, res, topics)
	node.Close()

	ctx = test.NewContext(t)
	store = storeMaker(t, ctx)
	node = NewTestNodeWithNameAndStore(t, ctx, "", store)
	for i, topic := range topics {
		resp, err := node.Query(ctx, &v1.QueryRequest{ContentTopics: []string{topic}})
		require.NoError(t, err)
		require.Len(t, resp.Envelopes, i+1)
	}
	node.Close()
}

func BenchmarkQuery(b *testing.B, ctx context.Context, topic *crdttest.TestStore, start, end uint64, pageSize uint32) {
	b.Run(fmt.Sprintf("%d/%d/%d", start, end, pageSize), func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			var cursor *v1.Cursor
			var resp *v1.QueryResponse
			var err error
			for count := end - start + 1; count > 0; count -= uint64(len(resp.Envelopes)) {
				resp, err = topic.Query(ctx, &v1.QueryRequest{
					StartTimeNs: start,
					EndTimeNs:   end,
					PagingInfo: &v1.PagingInfo{
						Limit:  pageSize,
						Cursor: cursor,
					},
				})
				require.NoError(b, err)
				cursor = resp.PagingInfo.Cursor
			}
		}
	})
}