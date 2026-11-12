package main

import (
	"log"
	"os"
	"path/filepath"

	"todo-scheduler/pkg/db"
	"todo-scheduler/pkg/server"
)

func main() {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	if err := db.InitDB(dbFile); err != nil {
		log.Fatalf("не удалось инициализировать БД: %v", err)
	}
	defer db.Close()

	webDir, err := filepath.Abs("web")
	if err != nil {
		log.Fatalf("не удалось определить путь к web: %v", err)
	}

	if err := server.Start(webDir); err != nil {
		log.Fatalf("ошибка запуска сервера: %v", err)
	}
}
