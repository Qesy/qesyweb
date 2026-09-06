package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	gamev1 "github.com/Qesy/QesyWeb/proto/game/v1"
	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
	room string
}
type Server struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]struct{}
}

func New() *Server {
	return &Server{rooms: make(map[string]map[*Client]struct{})}
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"localhost", "127.0.0.1", "[::1]"}})
	if err != nil {
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 16)}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	go s.writeLoop(ctx, client)
	s.readLoop(ctx, client)
	s.remove(client)
	_ = conn.Close(websocket.StatusNormalClosure, "")
}
func (s *Server) readLoop(ctx context.Context, client *Client) {
	client.conn.SetReadLimit(1 << 20)
	for {
		_, data, err := client.conn.Read(ctx)
		if err != nil {
			return
		}
		var incoming gamev1.Envelope
		if err := proto.Unmarshal(data, &incoming); err != nil {
			s.sendError(client, "invalid_message", "Protobuf 消息格式错误")
			continue
		}
		switch payload := incoming.Payload.(type) {
		case *gamev1.Envelope_Ping:
			s.sendMessage(client, &gamev1.Envelope{Version: 1, MessageId: 2, RequestId: incoming.RequestId, Payload: &gamev1.Envelope_Pong{Pong: &gamev1.Pong{TimestampMs: payload.Ping.TimestampMs}}})
		case *gamev1.Envelope_JoinRoom:
			s.join(client, payload.JoinRoom.RoomId)
		case *gamev1.Envelope_LeaveRoom:
			s.leave(client)
		case *gamev1.Envelope_RoomBroadcast:
			s.broadcast(client, payload.RoomBroadcast)
		}
	}
}
func (s *Server) writeLoop(ctx context.Context, client *Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-client.send:
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := client.conn.Write(writeCtx, websocket.MessageBinary, data)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
func (s *Server) sendMessage(client *Client, message *gamev1.Envelope) {
	data, err := proto.Marshal(message)
	if err == nil {
		select {
		case client.send <- data:
		default:
		}
	}
}
func (s *Server) sendError(client *Client, code, message string) {
	s.sendMessage(client, &gamev1.Envelope{Version: 1, MessageId: 100, Payload: &gamev1.Envelope_Error{Error: &gamev1.Error{Code: code, Message: message}}})
}
func (s *Server) join(client *Client, roomID string) {
	if roomID == "" {
		s.sendError(client, "room_required", "房间 ID 不能为空")
		return
	}
	s.leave(client)
	s.mu.Lock()
	room := s.rooms[roomID]
	if room == nil {
		room = make(map[*Client]struct{})
		s.rooms[roomID] = room
	}
	room[client] = struct{}{}
	client.room = roomID
	s.mu.Unlock()
	s.sendMessage(client, &gamev1.Envelope{Version: 1, MessageId: 3, Payload: &gamev1.Envelope_RoomBroadcast{RoomBroadcast: &gamev1.RoomBroadcast{RoomId: roomID, Text: "已加入房间"}}})
}
func (s *Server) leave(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if client.room == "" {
		return
	}
	room := s.rooms[client.room]
	delete(room, client)
	if len(room) == 0 {
		delete(s.rooms, client.room)
	}
	client.room = ""
}
func (s *Server) remove(client *Client) { s.leave(client) }
func (s *Server) broadcast(sender *Client, message *gamev1.RoomBroadcast) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room := s.rooms[sender.room]
	for client := range room {
		s.sendMessage(client, &gamev1.Envelope{Version: 1, MessageId: 4, Payload: &gamev1.Envelope_RoomBroadcast{RoomBroadcast: message}})
	}
}
