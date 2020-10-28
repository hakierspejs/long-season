package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alioygur/gores"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
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
			Nickname: p.Nickname,
			Password: pass,
			Online:   false,
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

func URLParamInjection(param string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value := chi.URLParam(r, param)
			ctx := context.WithValue(r.Context(), param, value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(r *http.Request) (int, error) {
	id, ok := r.Context().Value("id").(string)
	if !ok {
		return 0, errors.New("ID stored in context has inapropriate type.")
	}

	res, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func UserRead(db storage.Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := UserID(r)
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

		gores.JSONIndent(w, http.StatusOK, user.Public(), defaultPrefix, defaultIndent)
	}
}

func UserRemove(db storage.Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := UserID(r)
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
