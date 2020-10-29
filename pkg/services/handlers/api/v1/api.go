package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alioygur/gores"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/params"
	"github.com/hakierspejs/long-season/pkg/services/users"
	"github.com/hakierspejs/long-season/pkg/storage"
	serrors "github.com/hakierspejs/long-season/pkg/storage/errors"
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
			// TODO(thinkofher) Implement proper error handling.
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("decoding payload failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		pass, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
		if err != nil {
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("hashing password failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		id, err := db.New(r.Context(), models.User{
			UserPublicData: models.UserPublicData{
				Nickname: p.Nickname,
				Online:   false,
			},
			Password: pass,
		})
		if err != nil {
			// TODO(thinkofher) Implement proper error handling.
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
			// TODO(thinkofher) Implement proper error handling.
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("reading all users failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		result := users.PublicSlice(data)

		// Filter only online users.
		if r.URL.Query().Get("online") == "true" {
			filtered := make([]models.UserPublicData, 0, len(result))
			for _, u := range result {
				if u.Online {
					filtered = append(filtered, u)
				}
			}
			result = filtered
		}

		gores.JSONIndent(w, http.StatusOK, result, defaultPrefix, defaultIndent)
	}
}

func UserRead(db storage.Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := params.UserID(r)
		if err != nil {
			// TODO(thinkofher) Implement proper error handling.
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("reading user id failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		user, err := db.Read(r.Context(), id)
		if err != nil {
			// TODO(thinkofher) Implement proper error handling.
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("reading user failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		gores.JSONIndent(w, http.StatusOK, user.UserPublicData, defaultPrefix, defaultIndent)
	}
}

func UserRemove(db storage.Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := params.UserID(r)
		if err != nil {
			// TODO(thinkofher) Implement proper error handling.
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("reading user id failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		err = db.Remove(r.Context(), id)
		if err != nil {
			// TODO(thinkofher) Implement proper error handling.
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("removing user id failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateStatus updates online field of every user id database
// with MAC address equal to one from slice provided by
// user in request payload.
func UpdateStatus(db storage.Users) http.HandlerFunc {
	type payload struct {
		Addresses []string `json:"addresses"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		p := new(payload)

		err := json.NewDecoder(r.Body).Decode(p)
		if err != nil {
			jsonError(w, &jsonErrorBody{
				Message: fmt.Sprintf("invalid input: %s", err.Error()),
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

		updateded := make([]models.User, len(users))
		for _, user := range users {
			for _, address := range p.Addresses {
				if bytes.Equal([]byte(address), user.MAC) {
					user.Online = true
					updateded = append(updateded, user)
				}
			}
		}

		err = db.UpdateMany(r.Context(), updateded)
		if err != nil {
			switch err.(type) {
			case serrors.NoID:
				errNoID := err.(serrors.NoID)
				jsonError(w, &jsonErrorBody{
					Message: fmt.Sprintf("there is no user with id equal to %d", errNoID.ID()),
					Code:    http.StatusNotFound,
					Type:    "status-not-found",
				})
				return
			default:
				jsonError(w, &jsonErrorBody{
					Message: "ooops! things are not going that great after all",
					Code:    http.StatusInternalServerError,
					Type:    "internal-server-error",
				})
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}
