package auth

import (
	"errors"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

const claimKey string = "value"

var jwtSecret = []byte(os.Getenv("CARD_JUDGE_JWT_SECRET"))

func getValueTokenString(value string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		claimKey: value,
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Println(err)
		return "", errors.New("failed to sign token")
	}
	return tokenString, nil
}

func getTokenStringValue(tokenString string) (string, error) {
	token, err := getTokenStringToken(tokenString)
	if err != nil {
		log.Println(err)
		return "", errors.New("failed to get token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims[claimKey].(string), nil
	} else {
		return "", errors.New("could not get token claims")
	}
}

func getTokenStringToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}
