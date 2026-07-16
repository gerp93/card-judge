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
	BrandName string
	User      database.User
	LoggedIn  bool
}

type PagePolicy struct {
	LoginPaths        []string
	LoginPathPrefixes []string
	AdminPaths        []string
}

var brandName = "Card Judge"
var pagePolicy PagePolicy

func SetBrandName(name string) {
	brandName = name
}

func SetPagePolicy(policy PagePolicy) {
	pagePolicy = policy
}

func pageRequiresLogin(path string) bool {
	for _, p := range pagePolicy.LoginPaths {
		if path == p {
			return true
		}
	}
	for _, p := range pagePolicy.LoginPathPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func pageRequiresAdmin(path string) bool {
	for _, p := range pagePolicy.AdminPaths {
		if path == p {
			return true
		}
	}
	return false
}

func MiddlewareForPages(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		basePageData := BasePageData{
			PageTitle: brandName,
			BrandName: brandName,
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
		if pageRequiresLogin(r.URL.Path) {
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
		if pageRequiresAdmin(r.URL.Path) {
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
