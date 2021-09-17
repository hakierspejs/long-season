package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/thinkofher/horror"

	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/session"
	"github.com/hakierspejs/long-season/pkg/services/users"
	"github.com/hakierspejs/long-season/pkg/storage"
)

// Tokenizer turns given session State into single string.
type Tokenizer interface {
	Tokenize(context.Context, session.State) (string, error)
}

func Auth(tokenizer Tokenizer, db storage.Users) horror.HandlerFunc {
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

		match, err := users.AuthenticateWithPassword(ctx, users.AuthenticationDependencies{
			Request: users.AuthenticationRequest{
				Nickname: input.Nickname,
				Password: []byte(input.Password),
			},
			Storage:      db,
			ErrorFactory: errFactory,
		})
		if err != nil {
			return fmt.Errorf("users.AuthenticateWithPassword: %w", err)
		}

		newSession := session.New(ctx, session.Builder{
			UserID:   match.ID,
			Nickname: match.Nickname,
		})

		token, err := tokenizer.Tokenize(ctx, *newSession)
		if err != nil {
			return fmt.Errorf("tokenizer.Tokenize: %w", err)
		}

		return happier.OK(w, r, &response{
			Token: token,
		})
	}
}
