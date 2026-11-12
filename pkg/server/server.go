package server

import (
	"log"
	"net/http"
	"os"

	"todo-scheduler/pkg/api"
)

func Start(webDir string) error {
	mux := http.NewServeMux()

	api.Init(mux)

	fs := http.FileServer(http.Dir(webDir))
	mux.Handle("/", fs)

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Сервер запущен на порту :%s, раздача статики из: %s", port, webDir)

	return http.ListenAndServe(":"+port, mux)
}
