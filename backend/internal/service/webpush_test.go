package service_test

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	webpush "github.com/SherClockHolmes/webpush-go"

	"duekeep/internal/model"
	"duekeep/internal/service"
)

func TestWebPushSenderSees410(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusGone)
	}))
	t.Cleanup(srv.Close)
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	p256dh, auth := testSubKeys(t)
	sender := service.NewWebPushSender(pub, priv, "dev@duekeep.local")
	status, err := sender.Send(t.Context(), model.PushSubscription{
		Endpoint: srv.URL, P256dh: p256dh, Auth: auth,
	}, []byte(`{"title":"`+itemTitleDomain+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusGone {
		t.Fatalf("status %d", status)
	}
}

func testSubKeys(t *testing.T) (p256dh, auth string) {
	t.Helper()
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()),
		base64.RawURLEncoding.EncodeToString(secret)
}
