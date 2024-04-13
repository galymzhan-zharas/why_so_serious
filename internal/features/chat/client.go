package chat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 30 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type ClientList map[int64]*Client

type Client struct {
	ClientId    int64
	LeagueId    int64
	Playing     bool
	Conn        *websocket.Conn
	Manager     *Manager
	SendMsgChan chan Message
}

func NewClient(userId, leagueId int64, m *Manager, conn *websocket.Conn) *Client {
	return &Client{
		ClientId: userId,
		LeagueId: leagueId,
		Conn:     conn,
		Manager:  m,

		SendMsgChan: make(chan Message, 100),
	}
}

func (c *Client) writeMessage(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.SendMsgChan:
			log.Println("hererere-----------\n", string(message.Payload))
			log.Println("---------------")

			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, message.Payload)
			if err != nil {
				log.Printf("[%d] err :%v\n", c.ClientId, err)
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (c *Client) readMessage(ctx context.Context) {

	defer func() {
		c.Manager.LeaveQueue(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var battleReq BattleRequest
		err = json.Unmarshal(payload, &battleReq)
		if err != nil {
			log.Println(err)
		}

		c.Manager.messageBattle <- Message{
			BattleId: battleReq.BattleId,
			Payload:  payload,
			ClientId: c.ClientId,
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func ServeWs(manager *Manager, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	userId := r.Context().Value("userId").(int64)
	eventId := r.Context().Value("leagueId").(int64)

	client := NewClient(userId, eventId, manager, conn)

	client.Manager.JoinQueue(client)

	go client.writeMessage(r.Context())
	go client.readMessage(r.Context())
}

type BattleRequest struct {
	BattleId int64    `json:"battleId"`
	Command  string   `json:"command"`
	LeagueId int64    `json:"laguesId"`
	Correct  bool     `json:"correct`
	Question Question `json:"question"`
}

type BattleResponse struct {
	BattleId  int64      `json:"battleId"`
	Questions []Question `json:"question"`
}

type Question struct {
	Id        int64           `json:"questionId"`
	Module    string          `json:"module"`
	Payload   json.RawMessage `json:"payload"`
	Correct   int64           `json:"correct"`
	IsCorrect bool            `json:"isCorrect"`
}

var (
	CommandStart  = "start"
	CommandAnswer = "answer"
	CommandFinish = "Finish"
)
