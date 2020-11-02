package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/alioygur/gores"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/requests"
	"github.com/hakierspejs/long-season/pkg/services/result"
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
			result.JSONError(w, &result.JSONErrorBody{
				Message: fmt.Sprintf("decoding payload failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		pass, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
		if err != nil {
			result.JSONError(w, &result.JSONErrorBody{
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
			result.JSONError(w, &result.JSONErrorBody{
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
			result.JSONError(w, &result.JSONErrorBody{
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
		id, err := requests.UserID(r)
		if err != nil {
			// TODO(thinkofher) Implement proper error handling.
			result.JSONError(w, &result.JSONErrorBody{
				Message: fmt.Sprintf("reading user id failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		user, err := db.Read(r.Context(), id)
		if err != nil {
			// TODO(thinkofher) Implement proper error handling.
			result.JSONError(w, &result.JSONErrorBody{
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
		id, err := requests.UserID(r)
		if err != nil {
			// TODO(thinkofher) Implement proper error handling.
			result.JSONError(w, &result.JSONErrorBody{
				Message: fmt.Sprintf("reading user id failed, error: %s", err.Error()),
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		err = db.Remove(r.Context(), id)
		if err != nil {
			// TODO(thinkofher) Implement proper error handling.
			result.JSONError(w, &result.JSONErrorBody{
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
func UpdateStatus(dbu storage.Users, dbd storage.Devices) http.HandlerFunc {
	type payload struct {
		Addresses []string `json:"addresses"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		p := new(payload)

		err := json.NewDecoder(r.Body).Decode(p)
		if err != nil {
			result.JSONError(w, &result.JSONErrorBody{
				Message: fmt.Sprintf("invalid input: %s", err.Error()),
				Code:    http.StatusBadRequest,
				Type:    "bad-request",
			})
			return
		}

		err = storage.UpdateStatuses(r.Context(), p.Addresses, dbd, dbu)
		if err != nil {
			result.JSONError(w, &result.JSONErrorBody{
				Message: "ooops! things are not going that great after all",
				Code:    http.StatusInternalServerError,
				Type:    "internal-server-error",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}

func internalServerError(w http.ResponseWriter) {
	result.JSONError(w, &result.JSONErrorBody{
		Message: "ooops! things are not going that great after all",
		Code:    http.StatusInternalServerError,
		Type:    "internal-server-error",
	})
}

func notFound(w http.ResponseWriter) {
	result.JSONError(w, &result.JSONErrorBody{
		Message: "cannot find requested resources",
		Code:    http.StatusNotFound,
		Type:    "not-found",
	})
}

func badRequest(msg string, w http.ResponseWriter) {
	result.JSONError(w, &result.JSONErrorBody{
		Message: msg,
		Code:    http.StatusBadRequest,
		Type:    "bad-request",
	})
}

// DeviceAdd handles creation of new device for requesting user.
// Make sure to use with middleware.JWT (or another middleware that
// appends models.Claims to request), because this handler has
// to know some arbitrary user data.
func DeviceAdd(db storage.Devices) http.HandlerFunc {
	type payload struct {
		Tag string `json:"tag"`
		MAC string `json:"mac"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := requests.UserID(r)
		if err != nil {
			notFound(w)
			return
		}

		claims, err := requests.JWTClaims(r)
		if err != nil {
			// At this point handler should have
			// been provided with JWT claims, so we
			// will just return 500.
			internalServerError(w)
			return
		}

		p := new(payload)
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			internalServerError(w)
			return
		}

		device := models.Device{
			DevicePublicData: models.DevicePublicData{
				Owner: claims.Nickname,
				Tag:   p.Tag,
			},
			MAC:     []byte(p.MAC),
			OwnerID: claims.UserID,
		}

		_, err = db.New(r.Context(), userID, device)
		if errors.Is(err, serrors.ErrDeviceDuplication) {
			result.JSONError(w, &result.JSONErrorBody{
				Code:    http.StatusConflict,
				Type:    "conflict",
				Message: "tag already used",
			})
			return
		}
		if err != nil {
			internalServerError(w)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

type singleDevice struct {
	ID  int    `json:"id"`
	Tag string `json:"tag"`
}

// UserDevices handler responses with list of devices owned by
// requesting user.
func UserDevices(db storage.Devices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := requests.UserID(r)
		if err != nil {
			notFound(w)
			return
		}

		devices, err := db.OfUser(r.Context(), userID)
		if err != nil {
			notFound(w)
			return
		}

		result := make([]singleDevice, len(devices), cap(devices))
		for i, device := range devices {
			result[i] = singleDevice{device.ID, device.Tag}
		}

		gores.JSONIndent(w, http.StatusOK, result, defaultPrefix, defaultIndent)
	}
}

func sameOwner(userID, deviceOwnerID, claimsID int) bool {
	return (userID == deviceOwnerID) && (deviceOwnerID == claimsID)
}

func DeviceRead(db storage.Devices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID, err := requests.DeviceID(r)
		if err != nil {
			badRequest("invalid device id. please fix your request", w)
			return
		}

		userID, err := requests.UserID(r)
		if err != nil {
			badRequest("invalid user id. please fix your request", w)
			return
		}

		claims, err := requests.JWTClaims(r)
		if err != nil {
			// At this point handler should have
			// been provided with JWT claims, so we
			// will just return 500.
			internalServerError(w)
			return
		}

		device, err := db.Read(r.Context(), deviceID)
		if errors.As(err, &serrors.ErrNoID) {
			notFound(w)
			return
		}
		if err != nil {
			internalServerError(w)
			return
		}

		// Check if requesting user owns resources.
		if !sameOwner(userID, device.OwnerID, claims.UserID) {
			notFound(w)
			return
		}

		result := &singleDevice{
			ID:  device.ID,
			Tag: device.Tag,
		}
		gores.JSONIndent(w, http.StatusOK, result, defaultPrefix, defaultIndent)
	}
}

func DeviceRemove(db storage.Devices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID, err := requests.DeviceID(r)
		if err != nil {
			badRequest("invalid device id. please fix your request", w)
			return
		}

		userID, err := requests.UserID(r)
		if err != nil {
			badRequest("invalid user id. please fix your request", w)
			return
		}

		claims, err := requests.JWTClaims(r)
		if err != nil {
			// At this point handler should have
			// been provided with JWT claims, so we
			// will just return 500.
			internalServerError(w)
			return
		}

		device, err := db.Read(r.Context(), deviceID)
		if errors.As(err, &serrors.ErrNoID) {
			notFound(w)
			return
		}
		if err != nil {
			internalServerError(w)
			return
		}

		// Check if requesting user owns resources.
		if !sameOwner(userID, device.OwnerID, claims.UserID) {
			notFound(w)
			return
		}

		err = db.Remove(r.Context(), deviceID)
		if errors.As(err, &serrors.ErrNoID) {
			notFound(w)
			return
		}
		if err != nil {
			internalServerError(w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
