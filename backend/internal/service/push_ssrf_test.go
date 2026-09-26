package service_test

import (
	"context"
	"net"
	"testing"

	"duekeep/internal/model"
	"duekeep/internal/service"
)

func TestSubscribeRejectsPrivateEndpoint(t *testing.T) {
	t.Parallel()
	p := service.NewPush(nil, nil, "pub")
	for _, ep := range []string{
		"http://127.0.0.1/sub",
		"https://127.0.0.1/sub",
		"https://10.1.2.3/sub",
		"https://localhost/sub",
	} {
		err := p.Subscribe(t.Context(), "u1", model.PushSubscribe{
			Endpoint: ep,
			Keys:     model.PushKeys{P256dh: "p", Auth: "a"},
		}, "")
		if err == nil {
			t.Fatalf("accepted %s", ep)
		}
	}
}

func TestSubscribeRejectsResolvedPrivate(t *testing.T) {
	t.Parallel()
	p := service.NewPush(nil, nil, "pub")
	p.SetLookupForTest(func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("192.168.1.1")}, nil
	})
	err := p.Subscribe(t.Context(), "u1", model.PushSubscribe{
		Endpoint: "https://push.internal/sub",
		Keys:     model.PushKeys{P256dh: "p", Auth: "a"},
	}, "")
	if err == nil {
		t.Fatal("accepted private resolution")
	}
}
