package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NuEventTeam/chat/database"
	"github.com/NuEventTeam/chat/internal/features/chat"
)

var projectID = "asd"

func main() {
	log.Println("running server on 8002")
	go chat.RunChatServer(8002)

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	Stop()
	log.Println("application stopped")
}

func Stop() {
	db := database.DB.GetDb()
	query := `delete from player_queue where id > 0`

	db.Exec(context.Background(), query)
}
