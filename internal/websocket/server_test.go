package websocket

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	gamev1 "github.com/Qesy/QesyWeb/proto/game/v1"
	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"
)

func TestPingPongProtobuf(t *testing.T) {
	server := httptest.NewServer(New(nil))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+server.URL[len("http"):], nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	payload, err := proto.Marshal(&gamev1.Envelope{Version: 1, MessageId: 1, RequestId: "test", Payload: &gamev1.Envelope_Ping{Ping: &gamev1.Ping{TimestampMs: 123}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Write(ctx, websocket.MessageBinary, payload); err != nil {
		t.Fatal(err)
	}
	_, response, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var message gamev1.Envelope
	if err := proto.Unmarshal(response, &message); err != nil {
		t.Fatal(err)
	}
	if message.GetPong() == nil || message.GetPong().TimestampMs != 123 {
		t.Fatalf("response = %#v", message.GetPayload())
	}
}
