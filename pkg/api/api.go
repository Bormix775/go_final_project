package api

import (
	"net/http"
	"time"
)

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	var now time.Time
	var err error

	if nowStr != "" {
		now, err = time.Parse("20060102", nowStr)
		if err != nil {
			w.Write([]byte(""))
			return
		}
	} else {
		now = time.Now()
	}

	if dateStr == "" || repeat == "" {
		w.Write([]byte(""))
		return
	}

	next, err := NextDate(now, dateStr, repeat)
	if err != nil {
		w.Write([]byte(""))
		return
	}

	w.Write([]byte(next))
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		AddTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)

	}
}

func Init(mux *http.ServeMux) {
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.HandleFunc("/api/task", TaskHandler)
	mux.HandleFunc("/api/tasks", tasksHandler)
	mux.HandleFunc("/api/task/done", doneHandler)
	mux.HandleFunc("/api/signin", singInHandler)
}
