package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var cookieNameUserToken string = "CARD-JUDGE-USER-TOKEN"
var cookieNameRedirectURL string = "CARD-JUDGE-REDIRECT-URL"

func SetCookiePrefix(prefix string) {
	cookieNameUserToken = prefix + "-USER-TOKEN"
	cookieNameRedirectURL = prefix + "-REDIRECT-URL"
}

var secret []byte = getRandomBytes()

func getRandomBytes() []byte {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return bytes
}

func GetUserId(r *http.Request) (uuid.UUID, error) {
	token, err := getCookieValue(r, cookieNameUserToken)
	if err != nil {
		return uuid.Nil, errors.New("cookie not found")
	}

	userId, err := getUserIdFromToken(token)
	if err != nil {
		return uuid.Nil, errors.New("user not logged in")
	}

	return userId, nil
}

func SetUserId(w http.ResponseWriter, userId uuid.UUID) {
	token, expiry := getTokenFromUserId(userId)
	http.SetCookie(w, &http.Cookie{
		Name:    cookieNameUserToken,
		Value:   token,
		Path:    "/",
		Expires: expiry,
	})
}

func RemoveUserId(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    cookieNameUserToken,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
}

func GetRedirectUrl(r *http.Request) string {
	redirectPath, err := getCookieValue(r, cookieNameRedirectURL)
	if err != nil {
		return "/"
	}
	return redirectPath
}

func SetRedirectUrl(w http.ResponseWriter, url string) {
	http.SetCookie(w, &http.Cookie{
		Name:    cookieNameRedirectURL,
		Value:   url,
		Path:    "/",
		Expires: getExpiry(),
	})
}

func getCookieValue(r *http.Request, cookieName string) (string, error) {
	cookieFound := false
	cookieValue := ""
	for _, c := range r.Cookies() {
		if c.Name != cookieName {
			continue
		}
		cookieFound = true
		cookieValue = c.Value
		break
	}

	if !cookieFound {
		return "", errors.New("cookie not found")
	}

	return cookieValue, nil
}

func getTokenFromUserId(userId uuid.UUID) (string, time.Time) {
	expiry := getExpiry()
	value := fmt.Sprintf("%s %d", userId, expiry.Unix())
	valueBytes := []byte(value)
	valueBytesEncoded := base64.URLEncoding.EncodeToString(valueBytes)
	valueBytesHashed := getHashedBytes(valueBytes)
	valueBytesHashedEncoded := base64.URLEncoding.EncodeToString(valueBytesHashed)
	token := fmt.Sprintf("%s|%s", valueBytesEncoded, valueBytesHashedEncoded)
	return token, expiry
}

func getUserIdFromToken(token string) (uuid.UUID, error) {
	split := strings.Split(token, "|")
	if len(split) != 2 {
		return uuid.Nil, errors.New("invalid token")
	}

	valueBytesHashedEncoded := split[1]
	valueBytesHashed, err := base64.URLEncoding.DecodeString(valueBytesHashedEncoded)
	if err != nil {
		return uuid.Nil, errors.New("failed to decode token")
	}

	valueBytesEncoded := split[0]
	valueBytes, err := base64.URLEncoding.DecodeString(valueBytesEncoded)
	if err != nil {
		return uuid.Nil, errors.New("failed to decode token")
	}

	if !hmac.Equal(getHashedBytes(valueBytes), valueBytesHashed) {
		return uuid.Nil, errors.New("hash invalid")
	}

	value := string(valueBytes)
	split = strings.Split(value, " ")
	if len(split) != 2 {
		return uuid.Nil, errors.New("invalid token")
	}

	expiry, err := strconv.ParseInt(split[1], 10, 64)
	if err != nil || time.Now().Unix() > expiry {
		return uuid.Nil, errors.New("token expired")
	}

	userId, err := uuid.Parse(split[0])
	if err != nil {
		return uuid.Nil, err
	}

	return userId, nil
}

func getHashedBytes(bytes []byte) []byte {
	hash := hmac.New(sha256.New, secret)
	hash.Write(bytes)
	return hash.Sum(nil)
}

func getExpiry() time.Time {
	return time.Now().Add(12 * time.Hour)
}
