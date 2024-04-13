package chat

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/chat/database"
)

func (m *Manager) NewBattle(players []*Client) error {
	log.Println(players)
	battleId := rand.Int63()

	m.AddBattle(battleId, players)

	questions, err := getQuestions()
	if err != nil {
		log.Println(err)
		return err
	}

	resp := BattleResponse{
		BattleId:  battleId,
		Questions: questions,
	}

	js, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		return err
	}

	m1 := Message{
		LeagueId: 1,
		Payload:  js,
		ClientId: players[0].ClientId,
	}

	m2 := Message{
		LeagueId: 1,
		Payload:  js,
		ClientId: players[1].ClientId,
	}

	players[0].Manager.messageChan <- m1
	players[1].Manager.messageChan <- m2
	return nil
}

func (m *Manager) AddBattle(battleId int64, players []*Client) {
	m.Lock()
	defer m.Unlock()
	for _, client := range players {
		if client == nil {
			log.Println("nil client")
			return
		}
		if _, ok := m.Battle[battleId]; !ok {
			m.Battle[battleId] = make(ClientList)
		}
		m.Battle[battleId][client.ClientId] = client
	}

	query := qb.Insert("battles").Columns("custom_id,player1,player2").Values(battleId, players[0].ClientId, players[1].ClientId)
	stmt, args, _ := query.ToSql()

	db := database.DB.GetDb()
	_, err := db.Exec(context.Background(), stmt, args...)
	log.Println(err)
}

func (m *Manager) JoinQueue(client *Client) {
	defer m.Unlock()
	m.Lock()
	if client == nil {
		log.Println("nil client")
		return
	}
	if _, ok := m.Leagues[client.LeagueId]; !ok {
		m.Leagues[client.LeagueId] = make(ClientList)
	}
	m.Leagues[client.LeagueId][client.ClientId] = client
}

func (m *Manager) MakeBattle() {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			playersMap, err := getReadyPlayers()
			if err != nil {
				log.Println("While searching", err)
				continue
			}
			for leagueId, players := range playersMap {
				//create battle for each pair
				if len(players)%2 != 0 {
					players = players[:len(players)-1]
				}
				wg := &sync.WaitGroup{}
				log.Println(len(players))
				for i := 0; i < len(players)-1; i = i + 2 {
					wg.Add(1)
					go func(j int, id int64) {
						defer wg.Done()
						c := []*Client{}
						if val1, ok := m.Leagues[leagueId][int64(players[j].PlayerId)]; ok {
							c = append(c, val1)
						}

						if val1, ok := m.Leagues[leagueId][int64(players[j+1].PlayerId)]; ok {
							c = append(c, val1)

						}
						if len(c) == 2 {
							err = updatePlayerQueue(c[0].ClientId, c[1].ClientId)
							if err != nil {
								log.Println(err)
								return
							}
							m.NewBattle(c)

						}

					}(i, leagueId)
				}
				wg.Wait()
			}
		default:
		}
	}
}

type Player struct {
	PlayerId uint64
	LeagueId int64
}

func getReadyPlayers() (map[int64][]Player, error) {
	query := qb.Select("player_id, league_id").From("player_queue").Where(sq.Eq{"is_active": true}).OrderBy("random()")
	stmt, args, err := query.ToSql()

	if err != nil {
		return nil, err
	}

	db := database.DB.GetDb()

	rows, err := db.Query(context.Background(), stmt, args...)
	if err != nil {
		return nil, err
	}
	players := map[int64][]Player{}
	for rows.Next() {
		var p Player

		err := rows.Scan(&p.PlayerId, &p.LeagueId)
		if err != nil {
			return nil, err
		}

		if _, ok := players[p.LeagueId]; !ok {
			players[p.LeagueId] = []Player{}
		}
		players[p.LeagueId] = append(players[p.LeagueId], p)
	}

	return players, nil
}

func updatePlayerQueue(ids ...int64) error {
	query := qb.Update("player_queue").Set("is_active", false).Where(sq.Eq{"player_id": ids})
	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}
	db := database.DB.GetDb()
	_, err = db.Exec(context.Background(), stmt, args...)

	return err
}

func getQuestions() ([]Question, error) {
	query := `SELECT *
	FROM (
		SELECT id,module,content
		FROM questions
		WHERE module = 'listening'
		ORDER BY RANDOM()
		LIMIT 1
	) AS m1
	UNION ALL
	SELECT *
	FROM (
		SELECT id,module,content
		FROM questions
		WHERE module = 'vocabular'
		ORDER BY RANDOM()
		LIMIT 1
	) AS m2
	UNION ALL
	SELECT 
	FROM (
		SELECT id,module,content
		FROM questions
		WHERE module = 'grammar'
		ORDER BY RANDOM()
		LIMIT 1
	) AS m3;`

	var questions []Question
	var questionIds []int64
	db := database.DB.GetDb()

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q Question
		err := rows.Scan(&q.Id, &q.Module, &q.Payload)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
		questionIds = append(questionIds, q.Id)
	}

	query = `insert into battle_questions(battle_id,question_id) values($1,$2),values($1,$3),values($1,$4)`
	_, err = db.Exec(context.Background(), query, questionIds[0], questionIds[1], questionIds[2])
	if err != nil {
		return nil, err
	}
	return questions, nil
}
