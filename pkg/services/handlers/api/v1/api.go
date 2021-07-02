package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/alioygur/gores"
	"github.com/thinkofher/horror"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/apierr"
	"github.com/hakierspejs/long-season/pkg/services/devices"
	"github.com/hakierspejs/long-season/pkg/services/requests"
	"github.com/hakierspejs/long-season/pkg/services/result"
	"github.com/hakierspejs/long-season/pkg/services/users"
	"github.com/hakierspejs/long-season/pkg/storage"
	serrors "github.com/hakierspejs/long-season/pkg/storage/errors"
)

func conflict(msg string, w http.ResponseWriter) {
	result.JSONError(w, &result.JSONErrorBody{
		Message: msg,
		Code:    http.StatusConflict,
		Type:    "conflict",
	})
}

const internalServerErrorResponse = "Internal server error. Please try again later."

func UserCreate(db storage.Users) horror.HandlerFunc {
	type payload struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var p payload
		errFactory := apierr.FromRequest(r)

		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("api.UserCreate: decoding payload failed: %w", err),
				internalServerErrorResponse,
			)
		}

		if err := users.VerifyRegisterData(p.Nickname, p.Password); err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("api.UserCreate: invalid input: %w", err),
				fmt.Sprintf("Invalid input: %s.", err.Error()),
			)
		}

		pass, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("api.UserCreate: hashing password failed: %w", err),
				internalServerErrorResponse,
			)
		}

		id, err := db.New(r.Context(), models.User{
			UserPublicData: models.UserPublicData{
				Nickname: p.Nickname,
				Online:   false,
			},
			Password: pass,
		})
		if errors.Is(err, serrors.ErrNicknameTaken) {
			return errFactory.Conflict(
				fmt.Errorf("api.UserCreate: %w", err),
				fmt.Sprintf("Given username: %w is already taken.", p.Nickname),
			)
		}
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("api.UserCreate: creating new user failed, reason: %w", err),
				internalServerErrorResponse,
			)
		}

		gores.JSONIndent(w, http.StatusOK, &models.UserPublicData{
			ID:       id,
			Nickname: p.Nickname,
		}, defaultPrefix, defaultIndent)
		return nil
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

		filters := users.DefaultFilters()

		switch r.URL.Query().Get("online") {
		case "true":
			filters = append(filters, users.Online)
		case "false":
			filters = append(filters, users.Not(users.Online))
		}

		filtered := users.Filter(data, filters...)
		gores.JSONIndent(
			w, http.StatusOK, users.PublicSlice(filtered),
			defaultPrefix, defaultIndent,
		)
	}
}

func UserRead(db storage.Users) http.HandlerFunc {
	type response struct {
		models.UserPublicData
		Private *bool `json:"priv,omitempty"`
	}
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

		var privateMode *bool = nil
		claims, err := requests.JWTClaims(r)
		if err == nil && (claims.UserID == user.ID) {
			privateMode = &user.Private
		}

		gores.JSONIndent(w, http.StatusOK, &response{
			UserPublicData: user.UserPublicData,
			Private:        privateMode,
		}, defaultPrefix, defaultIndent)
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

func UserUpdate(db storage.Users) http.HandlerFunc {
	type payload struct {
		Private *bool `json:"priv,omitempty"`
	}

	type response struct {
		payload
		models.UserPublicData
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := requests.UserID(r)
		if err != nil {
			notFound(w)
			return
		}

		p := new(payload)
		if err := json.NewDecoder(r.Body).Decode(p); err != nil {
			result.JSONError(w, &result.JSONErrorBody{
				Message: fmt.Sprintf("invalid input: %s", err.Error()),
				Code:    http.StatusBadRequest,
				Type:    "bad-request",
			})
			return
		}

		if p.Private == nil {
			gores.JSONIndent(w, http.StatusCreated, struct{}{},
				defaultPrefix, defaultIndent)
			return
		}

		data, err := db.Read(r.Context(), userID)
		switch {
		case errors.As(err, &serrors.ErrNoID):
			notFound(w)
			return
		case err != nil:
			internalServerError(w)
			return
		}
		data.Private = *p.Private

		err = db.Update(r.Context(), *data)
		switch {
		case errors.As(err, &serrors.ErrNoID):
			notFound(w)
			return
		case err != nil:
			internalServerError(w)
			return
		}

		gores.JSONIndent(w, http.StatusOK, &response{
			payload: payload{
				Private: p.Private,
			},
			UserPublicData: data.UserPublicData,
		}, defaultPrefix, defaultIndent)
	}
}

// UpdateStatus updates online field of every user id database
// with MAC address equal to one from slice provided by
// user in request payload.
func UpdateStatus(ch chan<- []net.HardwareAddr) http.HandlerFunc {
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

		parsedAddresses := []net.HardwareAddr{}
		for _, address := range p.Addresses {
			parsedAddress, err := net.ParseMAC(address)
			if err != nil {
				badRequest("invalid mac address", w)
				return
			}
			parsedAddresses = append(parsedAddresses, parsedAddress)
		}

		// Send parsed addresses to deamon running in the background
		ch <- parsedAddresses

		w.WriteHeader(http.StatusAccepted)
		return
	}
}

func Status(counters storage.StatusTx) http.HandlerFunc {
	var response struct {
		Online  int `json:"online"`
		Unknown int `json:"unknown"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		err := counters.DevicesStatus(
			r.Context(),
			func(ctx context.Context, s storage.Status) error {
				online, err := s.OnlineUsers(ctx)
				if err != nil {
					return fmt.Errorf("failed to read online users: %w", err)
				}

				unknown, err := s.UnknownDevices(ctx)
				if err != nil {
					return fmt.Errorf("failed to read unknown devices: %w", err)
				}

				response.Online = online
				response.Unknown = unknown
				return nil
			},
		)
		if err != nil {
			internalServerError(w)
			return
		}

		gores.JSONIndent(w, http.StatusOK, response, defaultPrefix, defaultIndent)
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

type singleDevice struct {
	ID  int    `json:"id"`
	Tag string `json:"tag"`
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

	// TODO(thinkofher) Add Location header.
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

		mac, err := net.ParseMAC(p.MAC)
		if err != nil {
			badRequest("invalid mac address", w)
			return
		}

		hashedMac, err := bcrypt.GenerateFromPassword(mac, bcrypt.DefaultCost)
		if err != nil {
			internalServerError(w)
			return
		}

		device := models.Device{
			DevicePublicData: models.DevicePublicData{
				Owner: claims.Nickname,
				Tag:   p.Tag,
			},
			MAC:     hashedMac,
			OwnerID: claims.UserID,
		}

		newID, err := db.New(r.Context(), userID, device)
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

		gores.JSONIndent(w, http.StatusCreated, &singleDevice{
			ID:  newID,
			Tag: device.Tag,
		}, defaultPrefix, defaultIndent)
	}
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

func DeviceUpdate(db storage.Devices) http.HandlerFunc {
	type payload struct {
		MAC string `json:"mac"`
		Tag string `json:"tag"`
	}

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

		input := new(payload)
		err = json.NewDecoder(r.Body).Decode(input)
		if err != nil {
			badRequest("invalid input data", w)
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

		mac, err := net.ParseMAC(input.MAC)
		if err != nil {
			badRequest("invalid mac address", w)
			return
		}

		hashedMac, err := bcrypt.GenerateFromPassword(mac, bcrypt.DefaultCost)
		if err != nil {
			internalServerError(w)
			return
		}

		updated := devices.Update(*device, &devices.Changes{
			MAC: hashedMac,
			Tag: input.Tag,
		})

		err = db.Update(r.Context(), updated)
		if err != nil {
			internalServerError(w)
			return
		}

		gores.JSONIndent(w, http.StatusOK, &updated.DevicePublicData, defaultPrefix, defaultIndent)
	}
}
