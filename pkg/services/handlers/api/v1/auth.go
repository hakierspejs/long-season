package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"
	"github.com/thinkofher/horror"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/users"
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
		ctx := r.Context()
		errFactory := happier.FromRequest(r)

		input := new(payload)
		err := json.NewDecoder(r.Body).Decode(input)
		if err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("json.NewDecoder().Decode: %w", err),
				fmt.Sprintf("Invalid input: %s.", err.Error()),
			)
		}

		match, err := users.AuthenticateWithPassword(ctx, users.AuthenticateDependencies{
			Request: users.AuthenticateRequest{
				Nickname: input.Nickname,
				Password: []byte(input.Password),
			},
			Storage:      db,
			ErrorFactory: errFactory,
		})
		if err != nil {
			return fmt.Errorf("users.AuthenticateWithPassword: %w", err)
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
