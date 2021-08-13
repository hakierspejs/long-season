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
	"github.com/hakierspejs/long-season/pkg/services/devices"
	"github.com/hakierspejs/long-season/pkg/services/happier"
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
		errFactory := happier.FromRequest(r)

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
				fmt.Sprintf("Given username: %s is already taken.", p.Nickname),
			)
		}
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("api.UserCreate: creating new user failed, reason: %w", err),
				internalServerErrorResponse,
			)
		}

		return happier.OK(w, r, &models.UserPublicData{
			ID:       id,
			Nickname: p.Nickname,
		})
	}
}

func UsersAll(db storage.Users) horror.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		data, err := db.All(r.Context())

		if err != nil {
			return happier.FromRequest(r).InternalServerError(
				fmt.Errorf("db.All: %w", err),
				internalServerErrorResponse,
			)
		}

		filters := users.DefaultFilters()

		switch r.URL.Query().Get("online") {
		case "true":
			filters = append(filters, users.Online)
		case "false":
			filters = append(filters, users.Not(users.Online))
		}

		filtered := users.Filter(data, filters...)
		return happier.OK(w, r, users.PublicSlice(filtered))
	}
}

func UserRead(db storage.Users) horror.HandlerFunc {
	type response struct {
		models.UserPublicData
		Private *bool `json:"priv,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		id, err := requests.UserID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.UserID: %w", err),
				internalServerErrorResponse,
			)
		}

		user, err := db.Read(r.Context(), id)
		if errors.Is(err, serrors.ErrNoID) {
			return errFactory.NotFound(
				fmt.Errorf("db.Read: %w", err),
				fmt.Sprintf("there is no user with id: %d", id),
			)
		}
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.Read: %w", err),
				internalServerErrorResponse,
			)
		}

		var privateMode *bool = nil
		claims, err := requests.JWTClaims(r)
		if err == nil && (claims.UserID == user.ID) {
			privateMode = &user.Private
		}

		return happier.OK(w, r, &response{
			UserPublicData: user.UserPublicData,
			Private:        privateMode,
		})
	}
}

func UserRemove(db storage.Users) horror.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		id, err := requests.UserID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.UserID: %w", err),
				internalServerErrorResponse,
			)
		}

		err = db.Remove(r.Context(), id)
		if errors.Is(err, serrors.ErrNoID) {
			return errFactory.NotFound(
				fmt.Errorf("db.Remove: %w", err),
				fmt.Sprintf("there is no user with id: %d", id),
			)
		}
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.Remove: %w", err),
				internalServerErrorResponse,
			)
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

func UserUpdate(db storage.Users) horror.HandlerFunc {
	type payload struct {
		Private *bool `json:"priv,omitempty"`
	}

	type response struct {
		payload
		models.UserPublicData
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		userID, err := requests.UserID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.UserID: %w", err),
				internalServerErrorResponse,
			)
		}

		p := new(payload)
		if err := json.NewDecoder(r.Body).Decode(p); err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("json.NewDecoder().Decode: %w", err),
				fmt.Sprintf("Invalid input: %s.", err.Error()),
			)
		}

		if p.Private == nil {
			return happier.Created(w, r, struct{}{})
		}

		data, err := db.Read(r.Context(), userID)
		switch {
		case errors.As(err, &serrors.ErrNoID):
			return errFactory.NotFound(
				fmt.Errorf("db.Read: %w", err),
				fmt.Sprintf("there is no user with id: %d", userID),
			)
		case err != nil:
			return errFactory.InternalServerError(
				fmt.Errorf("db.Read: %w", err),
				internalServerErrorResponse,
			)
		}
		data.Private = *p.Private

		err = db.Update(r.Context(), *data)
		switch {
		case errors.As(err, &serrors.ErrNoID):
			return errFactory.NotFound(
				fmt.Errorf("db.Update: %w", err),
				fmt.Sprintf("there is no user with id: %d", userID),
			)
		case err != nil:
			return errFactory.InternalServerError(
				fmt.Errorf("db.Update: %w", err),
				internalServerErrorResponse,
			)
		}

		return happier.OK(w, r, &response{
			payload: payload{
				Private: p.Private,
			},
			UserPublicData: data.UserPublicData,
		})
	}
}

// UpdateStatus updates online field of every user id database
// with MAC address equal to one from slice provided by
// user in request payload.
func UpdateStatus(ch chan<- []net.HardwareAddr) horror.HandlerFunc {
	type payload struct {
		Addresses []string `json:"addresses"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		p := new(payload)
		errFactory := happier.FromRequest(r)

		err := json.NewDecoder(r.Body).Decode(p)
		if err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("json.NewDecoder().Decode: %w", err),
				fmt.Sprintf("Invalid input: %s.", err.Error()),
			)
		}

		parsedAddresses := []net.HardwareAddr{}
		for _, address := range p.Addresses {
			parsedAddress, err := net.ParseMAC(address)
			if err != nil {
				return errFactory.BadRequest(
					fmt.Errorf("net.ParseMAC: %w", err),
					fmt.Sprintf("invalid input: invalid mac address %s", address),
				)
			}
			parsedAddresses = append(parsedAddresses, parsedAddress)
		}

		// Send parsed addresses to deamon running in the background
		ch <- parsedAddresses

		return happier.Accepted(w, r)
	}
}

func Status(counters storage.StatusTx) horror.HandlerFunc {
	var response struct {
		Online  int `json:"online"`
		Unknown int `json:"unknown"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

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
			return errFactory.InternalServerError(
				fmt.Errorf("counters.DevicesStatus: %w", err),
				internalServerErrorResponse,
			)
		}

		return happier.OK(w, r, response)
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
func DeviceAdd(db storage.Devices) horror.HandlerFunc {
	type payload struct {
		Tag string `json:"tag"`
		MAC string `json:"mac"`
	}

	// TODO(thinkofher) Add Location header.
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		userID, err := requests.UserID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.UserID: %w", err),
				internalServerErrorResponse,
			)
		}

		claims, err := requests.JWTClaims(r)
		if err != nil {
			// At this point handler should have
			// been provided with JWT claims, so we
			// will just return 500.
			return errFactory.InternalServerError(
				fmt.Errorf("requests.JWTClaims: %w", err),
				internalServerErrorResponse,
			)
		}

		p := new(payload)
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("json.NewDecoder().Decode: %w", err),
				fmt.Sprintf("Invalid input: %s.", err.Error()),
			)
		}

		mac, err := net.ParseMAC(p.MAC)
		if err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("net.ParseMAC: %w", err),
				fmt.Sprintf("invalid input: invalid mac address %s", mac),
			)
		}

		hashedMac, err := bcrypt.GenerateFromPassword(mac, bcrypt.DefaultCost)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("bcrypt.GenerateFromPassword: %w", err),
				internalServerErrorResponse,
			)
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
			return errFactory.Conflict(
				fmt.Errorf("db.New: %w", err),
				fmt.Sprintf("tag already used"),
			)
		}
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.New: %w", err),
				internalServerErrorResponse,
			)
		}

		return happier.Created(w, r, &singleDevice{
			ID:  newID,
			Tag: device.Tag,
		})
	}
}

// UserDevices handler responses with list of devices owned by
// requesting user.
func UserDevices(db storage.Devices) horror.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		userID, err := requests.UserID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.UserID: %w", err),
				internalServerErrorResponse,
			)
		}

		devices, err := db.OfUser(r.Context(), userID)
		if err != nil {
			return errFactory.NotFound(
				fmt.Errorf("db.OfUser: %w", err),
				fmt.Sprintf("there is no user with id: %d", userID),
			)
		}

		result := make([]singleDevice, len(devices), cap(devices))
		for i, device := range devices {
			result[i] = singleDevice{device.ID, device.Tag}
		}

		return happier.OK(w, r, result)
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
