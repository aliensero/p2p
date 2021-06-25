package pubsub

import (
	"context"
	"testing"
	"time"
)

func TestPubSubServer(t *testing.T) {
	ctx := context.TODO()
	ps, err := DefaultPubsub()
	if err != nil {
		t.Error(err)
	}

	ti, err := ps.Join(topic)
	if err != nil {
		log.Error(err)
		return
	}
	sub, err := ti.Subscribe()
	if err != nil {
		log.Error(err)
		return
	}
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			log.Error(err)
			continue
		}
		t.Log(string(msg.Data))
	}

}

func TestPubSubClient(t *testing.T) {
	ps, err := DefaultPubsub()
	if err != nil {
		t.Error(err)
	}
	ti, err := ps.Join(topic)
	if err != nil {
		t.Error(err)
	}
	for {
		err = ti.Publish(context.TODO(), []byte("alien test"))
		if err != nil {
			t.Log(err)
		}
		time.Sleep(10 * time.Second)
	}
}
