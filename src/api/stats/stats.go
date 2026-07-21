package apiStats

import (
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/api"
	"github.com/grantfbarnes/card-judge/database"
	"github.com/grantfbarnes/card-judge/static"
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

	var timeframe string
	var topic string
	var subject string
	for key, val := range r.Form {
		switch key {
		case "timeframe":
			timeframe = val[0]
		case "topic":
			topic = val[0]
		case "subject":
			subject = val[0]
		}
	}

	headers, rows, err := database.GetStatsLeaderboard(userId, timeframe, topic, subject)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/components/tables/stats-table.html",
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

	_ = tmpl.ExecuteTemplate(w, "stats-table", data{
		Headers: headers,
		Rows:    rows,
	})
}
