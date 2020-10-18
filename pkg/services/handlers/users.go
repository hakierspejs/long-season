package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alioygur/gores"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/users"
	"github.com/hakierspejs/long-season/pkg/storage"
)

func UserCreate(db storage.Users) http.HandlerFunc {
	type payload struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var p payload

		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			// TODO(dudekb) Implement proper error handling.
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("decoding payload failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		id, err := db.New(r.Context(), models.User{
			Nickname: p.Nickname,
			Password: []byte(p.Password),
			Online:   false,
		})
		if err != nil {
			// TODO(dudekb) Implement proper error handling.
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("creating new user failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		gores.JSONIndent(w, http.StatusOK, &models.UserPublicData{
			ID:       id,
			Nickname: p.Nickname,
		}, defaultPrefix, defaultIndent)

	}
}

func UsersAll(db storage.Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := db.All(r.Context())

		if err != nil {
			// TODO(dudekb) Implement proper error handling.
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("reading all users failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		result := users.PublicSlice(data)
		gores.JSONIndent(w, http.StatusOK, result, defaultPrefix, defaultIndent)
	}
}
