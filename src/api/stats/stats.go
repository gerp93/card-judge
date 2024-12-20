package apiStats

import (
	"net/http"
	"text/template"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/api"
	"github.com/grantfbarnes/card-judge/database"
)

func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var topic string
	var subject string
	for key, val := range r.Form {
		if key == "topic" {
			topic = val[0]
		} else if key == "subject" {
			subject = val[0]
		}
	}

	headers, rows, err := database.GetLeaderboardStats(userId, topic, subject)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/components/table.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to parse HTML."))
		return
	}

	type data struct {
		Headers []string
		Rows    [][]string
	}

	_ = tmpl.ExecuteTemplate(w, "table", data{
		Headers: headers,
		Rows:    rows,
	})
}
