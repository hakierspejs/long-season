package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/alioygur/gores"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/utils"
	"github.com/hakierspejs/long-season/pkg/storage"
)

const idLen = 16

func ApiAuth(config models.Config, db storage.Users, rnd *rand.Rand) http.HandlerFunc {
	type payload struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	type response struct {
		Token string `json:"token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		input := new(payload)
		err := json.NewDecoder(r.Body).Decode(input)
		if err != nil {
			jsonError(w, &jsonErrorBody{
				Message: "could not understand payload",
				Code:    http.StatusBadRequest,
				Type:    "bad-request",
			})
			return
		}

		users, err := db.All(r.Context())
		if err != nil {
			jsonError(w, &jsonErrorBody{
				Message: "ooops! things are not going that great after all",
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		// Search for user with exactly same nickname.
		var match *models.User = nil
		for _, user := range users {
			if user.Nickname == input.Nickname {
				match = &user
			}
		}

		// Check if there is the user with given nickname
		// in the database.
		if match == nil {
			jsonError(w, &jsonErrorBody{
				Message: "there is no user with given nickname",
				Code:    http.StatusNotFound,
				Type:    "not-found",
			})
			return
		}

		// Check if passwords do match.
		if err := bcrypt.CompareHashAndPassword(
			match.Password,
			[]byte(input.Password),
		); err != nil {
			jsonError(w, &jsonErrorBody{
				Message: "given password does not match",
				Code:    http.StatusUnauthorized,
				Type:    "unauthorized",
			})
			return
		}

		// Prepare JWT token.
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
			Issuer:    input.Nickname,
			ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
			IssuedAt:  time.Now().Unix(),
			Id:        utils.RandString(idLen, rnd),
		})

		// Sign token.
		tokenString, err := token.SignedString([]byte(config.JWTSecret))
		if err != nil {
			jsonError(w, &jsonErrorBody{
				Message: "ooops! things are not going that great after all",
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		gores.JSONIndent(w, http.StatusOK, &response{
			Token: tokenString,
		}, defaultPrefix, defaultIndent)
	}
}

func jwtUser(r *http.Request) (string, error) {
	token, ok := r.Context().Value("jwt-user").(*jwt.Token)
	if !ok {
		return "", fmt.Errorf("failed to retrieve claims from token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to retrieve standard claims")
	}

	return claims["iss"].(string), nil
}

func AuthResource() http.HandlerFunc {
	type response struct {
		Message string `json:"msg"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		nickname, err := jwtUser(r)
		if err != nil {
			fmt.Println(err)
			jsonError(w, &jsonErrorBody{
				Message: "You have to provide correct bearer token.",
				Code:    http.StatusUnauthorized,
				Type:    "unauthorized",
			})
			return
		}

		gores.JSONIndent(w, http.StatusOK, &response{
			Message: fmt.Sprintf("You are authenticated as %s.", nickname),
		}, defaultPrefix, defaultIndent)
	}
}
