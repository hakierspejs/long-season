package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alioygur/gores"
	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/result"
	"github.com/hakierspejs/long-season/pkg/storage"
)

func ApiAuth(config models.Config, db storage.Users) http.HandlerFunc {
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
			result.JSONError(w, &result.JSONErrorBody{
				Message: "could not understand payload",
				Code:    http.StatusBadRequest,
				Type:    "bad-request",
			})
			return
		}

		users, err := db.All(r.Context())
		if err != nil {
			result.JSONError(w, &result.JSONErrorBody{
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
			result.JSONError(w, &result.JSONErrorBody{
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
			result.JSONError(w, &result.JSONErrorBody{
				Message: "given password does not match",
				Code:    http.StatusUnauthorized,
				Type:    "unauthorized",
			})
			return
		}

		signer, err := jwt.NewSignerHS(jwt.HS256, []byte(config.JWTSecret))
		if err != nil {
			result.JSONError(w, &result.JSONErrorBody{
				Message: "ooops! things are not going that great after all",
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		builder := jwt.NewBuilder(signer)

		now := time.Now()
		id := uuid.New()

		token, err := builder.Build(&models.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    config.AppName,
				Audience:  []string{"ls-apiv1"},
				Subject:   "auth",
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 48)),
				IssuedAt:  jwt.NewNumericDate(now),
				ID:        id.String(),
			},
			Nickname: match.Nickname,
			UserID:   match.ID,
		})
		if err != nil {
			result.JSONError(w, &result.JSONErrorBody{
				Message: "ooops! things are not going that great after all",
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		gores.JSONIndent(w, http.StatusOK, &response{
			Token: token.String(),
		}, defaultPrefix, defaultIndent)
	}
}

func JWTClaims(r *http.Request) (*models.Claims, error) {
	claims, ok := r.Context().Value("jwt-user").(*models.Claims)
	if !ok {
		return nil, fmt.Errorf("failed")
	}
	return claims, nil
}

func AuthResource() http.HandlerFunc {
	type response struct {
		Message string `json:"msg"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := JWTClaims(r)
		if err != nil {
			fmt.Println(err)
			result.JSONError(w, &result.JSONErrorBody{
				Message: "You have to provide correct bearer token.",
				Code:    http.StatusUnauthorized,
				Type:    "unauthorized",
			})
			return
		}

		gores.JSONIndent(w, http.StatusOK, &response{
			Message: fmt.Sprintf("You are authenticated as %s.", claims.Nickname),
		}, defaultPrefix, defaultIndent)
	}
}
