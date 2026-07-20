package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
	"github.com/grantfbarnes/card-judge/database"
)

type RequestContextKey string

const basePageDataRequestContextKey RequestContextKey = "basePageDataRequestContextKey"
const userIdRequestContextKey RequestContextKey = "userIdRequestContextKey"

type BasePageData struct {
	PageTitle string
	User      database.User
	LoggedIn  bool
}

func MiddlewareForPages(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		basePageData := BasePageData{
			PageTitle: "Card Judge",
			User:      database.User{},
			LoggedIn:  false,
		}

		userId, err := auth.GetUserId(r)
		if err == nil {
			user, err := database.GetUser(userId)
			if user.Id == uuid.Nil {
				auth.RemoveUserId(w)
			} else if err == nil {
				basePageData.User = user
				basePageData.LoggedIn = true
			}
		}

		// required to be logged in
		if r.URL.Path == "/account" ||
			r.URL.Path == "/users" ||
			r.URL.Path == "/review" ||
			r.URL.Path == "/lobbies" ||
			r.URL.Path == "/decks" ||
			strings.HasPrefix(r.URL.Path, "/stats/") ||
			strings.HasPrefix(r.URL.Path, "/lobby/") ||
			strings.HasPrefix(r.URL.Path, "/deck/") {
			if !basePageData.LoggedIn {
				auth.SetRedirectUrl(w, r.URL.Path+"?"+r.URL.RawQuery)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}

		// required to not be logged in
		if r.URL.Path == "/login" {
			if basePageData.LoggedIn {
				http.Redirect(w, r, auth.GetRedirectUrl(r), http.StatusSeeOther)
				return
			}
		}

		// required to be admin
		if r.URL.Path == "/users" ||
			r.URL.Path == "/review" {
			if !basePageData.User.IsAdmin {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}

		r = r.WithContext(context.WithValue(r.Context(), basePageDataRequestContextKey, basePageData))

		next.ServeHTTP(w, r)
	})
}

func GetBasePageData(r *http.Request) BasePageData {
	return r.Context().Value(basePageDataRequestContextKey).(BasePageData)
}

func MiddlewareForAPIs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, _ := auth.GetUserId(r)
		r = r.WithContext(context.WithValue(r.Context(), userIdRequestContextKey, userId))
		next.ServeHTTP(w, r)
	})
}

func GetUserId(r *http.Request) uuid.UUID {
	return r.Context().Value(userIdRequestContextKey).(uuid.UUID)
}

func UserIsAdmin(r *http.Request) bool {
	userId := GetUserId(r)
	if userId == uuid.Nil {
		return false
	}

	isAdmin, _ := database.GetUserIsAdmin(userId)
	return isAdmin
}
