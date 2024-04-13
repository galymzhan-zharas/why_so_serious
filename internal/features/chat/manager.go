package chat

import (
	"encoding/json"
	"log"
	"sync"
)

var ChatManager *Manager

type Leagues map[int64]ClientList

type Manager struct {
	Leagues       Leagues
	Battle        Leagues
	messageBattle chan Message
	messageChan   chan Message
	sync.RWMutex
}

type Message struct {
	LeagueId int64
	Payload  json.RawMessage
	ClientId int64
	BattleId int64
}

func NewManager() *Manager {
	return &Manager{
		Leagues:       make(Leagues),
		Battle:        make(Leagues),
		messageBattle: make(chan Message, 1000),
		messageChan:   make(chan Message, 1000),
	}
}

func (m *Manager) LeaveQueue(client *Client) {
	defer m.Unlock()
	m.Lock()

	if event, ok := m.Leagues[client.LeagueId]; ok {
		if client, ok := event[client.ClientId]; ok {
			delete(m.Leagues[client.LeagueId], client.ClientId)
			close(client.SendMsgChan)
		}
	}
}

func (m *Manager) Run() {
	for {
		select {
		case message := <-m.messageChan:
			if client, ok := m.Leagues[message.LeagueId][message.ClientId]; ok {
				client := client
				go func() {
					client.SendMsgChan <- message
				}()
			}
		case message := <-m.messageBattle:
			{
				log.Println("messageing battle")
				for key, val := range m.Battle[message.BattleId] {
					if key != message.ClientId {
						log.Println(key)
						val.SendMsgChan <- message
					}
				}
			}
		}
	}
}
