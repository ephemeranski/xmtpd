package api_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtpd/pkg/db/queries"
	"github.com/xmtp/xmtpd/pkg/proto/xmtpv4/message_api"
	"github.com/xmtp/xmtpd/pkg/testutils"
)

func setupQueryTest(t *testing.T, db *sql.DB) []queries.InsertGatewayEnvelopeParams {
	db_rows := []queries.InsertGatewayEnvelopeParams{
		{
			OriginatorNodeID:     1,
			OriginatorSequenceID: 1,
			Topic:                []byte("topicA"),
			OriginatorEnvelope: testutils.Marshal(
				t,
				testutils.CreateOriginatorEnvelope(t, 1, 1),
			),
		},
		{
			OriginatorNodeID:     2,
			OriginatorSequenceID: 1,
			Topic:                []byte("topicA"),
			OriginatorEnvelope: testutils.Marshal(
				t,
				testutils.CreateOriginatorEnvelope(t, 2, 1),
			),
		},
		{
			OriginatorNodeID:     1,
			OriginatorSequenceID: 2,
			Topic:                []byte("topicB"),
			OriginatorEnvelope: testutils.Marshal(
				t,
				testutils.CreateOriginatorEnvelope(t, 1, 2),
			),
		},
		{
			OriginatorNodeID:     2,
			OriginatorSequenceID: 2,
			Topic:                []byte("topicB"),
			OriginatorEnvelope: testutils.Marshal(
				t,
				testutils.CreateOriginatorEnvelope(t, 2, 2),
			),
		},
		{
			OriginatorNodeID:     1,
			OriginatorSequenceID: 3,
			Topic:                []byte("topicA"),
			OriginatorEnvelope: testutils.Marshal(
				t,
				testutils.CreateOriginatorEnvelope(t, 1, 3),
			),
		},
	}
	testutils.InsertGatewayEnvelopes(t, db, db_rows)
	return db_rows
}

func TestQueryAllEnvelopes(t *testing.T) {
	api, db, cleanup := testutils.NewTestAPIClient(t)
	defer cleanup()
	db_rows := setupQueryTest(t, db)

	resp, err := api.QueryEnvelopes(
		context.Background(),
		&message_api.QueryEnvelopesRequest{
			Query: &message_api.EnvelopesQuery{},
			Limit: 0,
		},
	)
	require.NoError(t, err)
	checkRowsMatchProtos(t, db_rows, []int{0, 1, 2, 3, 4}, resp.GetEnvelopes())
}

func TestQueryPagedEnvelopes(t *testing.T) {
	api, db, cleanup := testutils.NewTestAPIClient(t)
	defer cleanup()
	db_rows := setupQueryTest(t, db)

	resp, err := api.QueryEnvelopes(
		context.Background(),
		&message_api.QueryEnvelopesRequest{
			Query: &message_api.EnvelopesQuery{},
			Limit: 2,
		},
	)
	require.NoError(t, err)
	checkRowsMatchProtos(t, db_rows, []int{0, 1}, resp.GetEnvelopes())
}

func TestQueryEnvelopesByOriginator(t *testing.T) {
	api, db, cleanup := testutils.NewTestAPIClient(t)
	defer cleanup()
	db_rows := setupQueryTest(t, db)

	resp, err := api.QueryEnvelopes(
		context.Background(),
		&message_api.QueryEnvelopesRequest{
			Query: &message_api.EnvelopesQuery{
				Filter: &message_api.EnvelopesQuery_OriginatorNodeId{
					OriginatorNodeId: 2,
				},
				LastSeen: nil,
			},
			Limit: 0,
		},
	)
	require.NoError(t, err)
	checkRowsMatchProtos(t, db_rows, []int{1, 3}, resp.GetEnvelopes())
}

func TestQueryEnvelopesByTopic(t *testing.T) {
	api, db, cleanup := testutils.NewTestAPIClient(t)
	defer cleanup()
	db_rows := setupQueryTest(t, db)

	resp, err := api.QueryEnvelopes(
		context.Background(),
		&message_api.QueryEnvelopesRequest{
			Query: &message_api.EnvelopesQuery{
				Filter:   &message_api.EnvelopesQuery_Topic{Topic: []byte("topicA")},
				LastSeen: nil,
			},
			Limit: 0,
		},
	)
	require.NoError(t, err)
	checkRowsMatchProtos(t, db_rows, []int{0, 1, 4}, resp.GetEnvelopes())
}

func TestQueryEnvelopesFromLastSeen(t *testing.T) {
	api, db, cleanup := testutils.NewTestAPIClient(t)
	defer cleanup()
	db_rows := setupQueryTest(t, db)

	resp, err := api.QueryEnvelopes(
		context.Background(),
		&message_api.QueryEnvelopesRequest{
			Query: &message_api.EnvelopesQuery{
				Filter:   nil,
				LastSeen: &message_api.VectorClock{NodeIdToSequenceId: map[uint32]uint64{1: 2}},
			},
			Limit: 0,
		},
	)
	require.NoError(t, err)
	checkRowsMatchProtos(t, db_rows, []int{1, 3, 4}, resp.GetEnvelopes())
}

func TestQueryEnvelopesWithEmptyResult(t *testing.T) {
	api, db, cleanup := testutils.NewTestAPIClient(t)
	defer cleanup()
	db_rows := setupQueryTest(t, db)

	resp, err := api.QueryEnvelopes(
		context.Background(),
		&message_api.QueryEnvelopesRequest{
			Query: &message_api.EnvelopesQuery{
				Filter: &message_api.EnvelopesQuery_Topic{
					Topic: []byte("topicC"),
				},
			},
			Limit: 0,
		},
	)
	require.NoError(t, err)
	checkRowsMatchProtos(t, db_rows, []int{}, resp.GetEnvelopes())
}

func checkRowsMatchProtos(
	t *testing.T,
	allRows []queries.InsertGatewayEnvelopeParams,
	matchingIndices []int,
	protos []*message_api.OriginatorEnvelope,
) {
	require.Len(t, protos, len(matchingIndices))
	for i, p := range protos {
		row := allRows[matchingIndices[i]]
		require.Equal(t, row.OriginatorEnvelope, testutils.Marshal(t, p))
	}
}
