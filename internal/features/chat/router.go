package chat

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func getRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/ws/{leagueId}", Authorize(joinBattleHandler)).Methods("GET")

	r.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "Ok")
	}).Methods("GET")

	return r
}
