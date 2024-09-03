package api

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/xmtp/xmtpd/pkg/db"
	"github.com/xmtp/xmtpd/pkg/db/queries"
	"github.com/xmtp/xmtpd/pkg/proto/xmtpv4/message_api"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	subscriptionBufferSize    = 1024
	maxSubscriptionsPerClient = 10000
)

type subscriber = chan<- []*message_api.OriginatorEnvelope

type subscribeWorker struct {
	ctx context.Context
	log *zap.Logger

	dbSubscription <-chan []queries.GatewayEnvelope
	// Assumption: listeners cannot be in multiple slices
	globalListeners     []subscriber
	originatorListeners map[uint32][]subscriber
	topicListeners      map[string][]subscriber
}

func startSubscribeWorker(
	ctx context.Context,
	log *zap.Logger,
	store *sql.DB,
) (*subscribeWorker, error) {
	q := queries.New(store)
	// Get vector clock from DB
	query := func(ctx context.Context, lastSeen db.VectorClock, numRows int32) ([]queries.GatewayEnvelope, db.VectorClock, error) {
		envs, err := q.
			SelectGatewayEnvelopes(
				ctx,
				*db.SetVectorClock(&queries.SelectGatewayEnvelopesParams{}, lastSeen),
			)
		// TODO(rich) log size of envs
		if err != nil {
			return nil, lastSeen, err
		}
		for _, env := range envs {
			lastSeen[uint32(env.OriginatorNodeID)] = uint64(env.OriginatorSequenceID)
		}
		return envs, lastSeen, nil
	}
	subscription := db.NewDBSubscription(
		ctx,
		log,
		query,
		db.VectorClock{}, // TODO(rich) fetch from DB
		db.PollingOptions{
			Interval: 100 * time.Millisecond,
			NumRows:  100,
		}, // TODO(rich) Make numRows nullable
	)
	dbChan, err := subscription.Start()
	if err != nil {
		return nil, err
	}
	worker := &subscribeWorker{
		ctx:                 ctx,
		log:                 log.Named("subscribeWorker"),
		dbSubscription:      dbChan,
		globalListeners:     make([]subscriber, 0),
		originatorListeners: make(map[uint32][]subscriber),
		topicListeners:      make(map[string][]subscriber),
	}

	go worker.start()

	return worker, nil
}

func (s *subscribeWorker) start() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case new_batch := <-s.dbSubscription:
			for _, row := range new_batch {
				s.dispatch(&row)
			}
		}
	}
}

func (s *subscribeWorker) dispatch(
	row *queries.GatewayEnvelope,
) {
	// TODO(rich) log how long this takes
	bytes := row.OriginatorEnvelope
	env := &message_api.OriginatorEnvelope{}
	err := proto.Unmarshal(bytes, env)
	if err != nil {
		s.log.Error("Failed to unmarshal envelope", zap.Error(err))
		return
	}
	for _, listener := range s.originatorListeners[uint32(row.OriginatorNodeID)] {
		select {
		case listener <- []*message_api.OriginatorEnvelope{env}:
		default: // TODO(rich) log here
		}
	}
	for _, listener := range s.topicListeners[hex.EncodeToString(row.Topic)] {
		select {
		case listener <- []*message_api.OriginatorEnvelope{env}:
		default:
		}
	}
	for _, listener := range s.globalListeners {
		select {
		case listener <- []*message_api.OriginatorEnvelope{env}:
		default:
		}
	}
}

// TODO(rich) clearer naming - broadcast/listen, publish/subscribe, in/out
func (s *subscribeWorker) subscribe(
	requests []*message_api.BatchSubscribeEnvelopesRequest_SubscribeEnvelopesRequest,
) (<-chan []*message_api.OriginatorEnvelope, error) {
	// TODO(rich) count how many subscriptions the server has
	subscribeAll := false
	topics := make(map[string]bool, len(requests))
	originators := make(map[uint32]bool, len(requests))

	if len(requests) > maxSubscriptionsPerClient {
		// When a client subscribes to too many originators or topics, we treat it as a request to
		// subscribe to all instead of throwing an error. We rely on the client's existing
		// filtering logic rather than forcing clients to respond to an error.
		subscribeAll = true
	} else {
		for _, req := range requests {
			enum := req.GetQuery().GetFilter()
			if enum == nil {
				subscribeAll = true
			}
			switch filter := enum.(type) {
			case *message_api.EnvelopesQuery_Topic:
				if len(filter.Topic) == 0 {
					return nil, status.Errorf(codes.InvalidArgument, "missing topic")
				}
				topics[hex.EncodeToString(filter.Topic)] = true
			case *message_api.EnvelopesQuery_OriginatorNodeId:
				// TODO(rich) validate filter
				originators[filter.OriginatorNodeId] = true
			default:
				subscribeAll = true
			}
		}
	}

	ch := make(chan []*message_api.OriginatorEnvelope, subscriptionBufferSize)

	if subscribeAll {
		if len(topics) > 0 || len(originators) > 0 {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"cannot filter by topic or originator when subscribing to all",
			)
		}
		// TODO(rich) thread safety
		s.globalListeners = append(s.globalListeners, ch)
	} else if len(topics) > 0 {
		if len(originators) > 0 {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"cannot filter by both topic and originator in same subscription request",
			)
		}
		for topic := range topics {
			// TODO(rich) Handle uncreated slice
			s.topicListeners[topic] = append(s.topicListeners[topic], ch)
		}
	} else if len(originators) > 0 {
		for originator := range originators {
			// TODO(rich) Handle uncreated slice
			s.originatorListeners[originator] = append(s.originatorListeners[originator], ch)
		}
	}

	return ch, nil
}