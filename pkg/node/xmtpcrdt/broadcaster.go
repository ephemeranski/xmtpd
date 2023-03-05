package xmtpcrdtnode

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/xmtp/xmtpd/pkg/context"
	"github.com/xmtp/xmtpd/pkg/crdt/types"
)

type broadcaster struct {
	topic *pubsub.Topic
	C     chan *types.Event
}

func newBroadcaster(topic *pubsub.Topic, ch chan *types.Event) (*broadcaster, error) {
	return &broadcaster{
		topic: topic,
		C:     ch,
	}, nil
}

func (b *broadcaster) Broadcast(ctx context.Context, ev *types.Event) error {
	evB, err := ev.ToBytes()
	if err != nil {
		return err
	}
	return b.topic.Publish(ctx, evB)
}

func (b *broadcaster) Next(ctx context.Context) (*types.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case ev := <-b.C:
		return ev, nil
	}
}

func (b *broadcaster) Close() error {
	close(b.C)
	return nil
}
