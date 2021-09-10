package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"
	"github.com/thinkofher/horror"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/result"
	"github.com/hakierspejs/long-season/pkg/storage"
)

func ApiAuth(config models.Config, db storage.Users) horror.HandlerFunc {
	type payload struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	type response struct {
		Token string `json:"token"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		input := new(payload)
		err := json.NewDecoder(r.Body).Decode(input)
		if err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("json.NewDecoder().Decode: %w", err),
				fmt.Sprintf("Invalid input: %s.", err.Error()),
			)
		}

		users, err := db.All(r.Context())
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.All: %w", err),
				internalServerErrorResponse,
			)
		}

		// Search for user with exactly same nickname.
		var match *models.User = nil
		for _, user := range users {
			if user.Nickname == input.Nickname {
				match = &user
				break
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
			return errFactory.NotFound(
				fmt.Errorf("match == nil, user given nickname: %s, not found", input.Nickname),
				fmt.Sprintf("there is no user with given nickname: \"%s\"", input.Nickname),
			)
		}

		// Check if passwords do match.
		if err := bcrypt.CompareHashAndPassword(
			match.Password,
			[]byte(input.Password),
		); err != nil {
			return errFactory.Unauthorized(
				fmt.Errorf("bcrypt.CompareHashAndPassword: %w", err),
				fmt.Sprintf("given password does not match"),
			)
		}

		signer, err := jwt.NewSignerHS(jwt.HS256, []byte(config.JWTSecret))
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("jwt.NewSignerHS: %w", err),
				internalServerErrorResponse,
			)
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
			Private:  match.Private,
		})
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("builder.Build: %w", err),
				internalServerErrorResponse,
			)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "jwt-token",
			Expires:  now.Add(time.Hour * 4),
			Value:    token.String(),
			HttpOnly: true,
			Path:     "/",
		})

		return happier.OK(w, r, &response{
			Token: token.String(),
		})
	}
}
