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

	r.HandleFunc("/login", LoginUser).Methods("POST")

	r.HandleFunc("/register", RegisterUser).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("."+"/static/"))))

	return r
}
