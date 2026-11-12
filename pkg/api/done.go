package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"todo-scheduler/pkg/db"
)

func doneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "не указан идентификатор задачи"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "задача не найдена"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, struct{}{})
		return
	}

	now := time.Now()
	next, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if err := db.UpdateDate(next, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, struct{}{})
}
