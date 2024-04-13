package chat

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/NuEventTeam/chat/database"
)

var projectID = "kzlanguage"

func RunChatServer(port int) error {
	ChatManager = NewManager()
	go ChatManager.Run()
	go ChatManager.MakeBattle()
	database.NewDatabase(context.Background())

	srv := &http.Server{
		Handler:      getRouter(),
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 1 * time.Second,
		ReadTimeout:  1 * time.Second,
	}

	return srv.ListenAndServe()
}
