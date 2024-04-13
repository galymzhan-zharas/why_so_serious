package chat

import (
	"context"
	"log"
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/chat/database"
	"github.com/gorilla/mux"
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func joinBattleHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(mux.Vars(r))
	leagueId, err := strconv.ParseInt(mux.Vars(r)["leagueId"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrogn"))
		log.Println(err, "")
		return
	}

	err = addPlayerToQueue(r.Context().Value("userId").(int64), leagueId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrogn"))
		log.Println(err, "")
		return
	}
	ctx := context.WithValue(r.Context(), "leagueId", leagueId)
	ServeWs(ChatManager, w, r.WithContext(ctx))
}

func addPlayerToQueue(playerId int64, leagueId int64) error {
	query := qb.Insert("player_queue").Columns("player_id", "league_id").Values(playerId, leagueId)

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}
	db := database.DB.GetDb()

	_, err = db.Exec(context.Background(), stmt, args...)
	return err
}
