package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/thinkofher/horror"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/devices"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/requests"
	"github.com/hakierspejs/long-season/pkg/services/result"
	"github.com/hakierspejs/long-season/pkg/services/session"
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

		id, err := users.Add(r.Context(), users.AddUserRequest{
			Nickname: p.Nickname,
			Password: []byte(p.Password),
			Storage:  db,
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

func UsersAll(db storage.Users, adapter storage.UserAdapter) horror.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		data, err := db.All(ctx)
		if err != nil {
			return happier.FromRequest(r).InternalServerError(
				fmt.Errorf("db.All: %w", err),
				internalServerErrorResponse,
			)
		}

		adaptedData, err := adapter.Users(ctx, data)
		if err != nil {
			return happier.FromRequest(r).InternalServerError(
				fmt.Errorf("adapter.Users: %w", err),
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

		filtered := users.Filter(adaptedData, filters...)
		return happier.OK(w, r, users.PublicSlice(filtered))
	}
}

func UserRead(renewer session.Renewer, db storage.Users, adapter storage.UserAdapter) horror.HandlerFunc {
	type response struct {
		models.UserPublicData
		Private *bool `json:"priv,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		errFactory := happier.FromRequest(r)

		id, err := requests.UserID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.UserID: %w", err),
				internalServerErrorResponse,
			)
		}

		user, err := db.Read(ctx, id)
		if errors.Is(err, serrors.ErrNoID) {
			return errFactory.NotFound(
				fmt.Errorf("db.Read: %w", err),
				fmt.Sprintf("there is no user with id: %s", id),
			)
		}
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.Read: %w", err),
				internalServerErrorResponse,
			)
		}

		var privateMode *bool = nil
		state, err := renewer.Renew(r)
		if err == nil && (state.UserID == user.ID) {
			privateMode = &user.Private
		}

		adapted, err := adapter.User(ctx, *user)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("adapter.User: %w", err),
				internalServerErrorResponse,
			)
		}

		return happier.OK(w, r, &response{
			UserPublicData: adapted.UserPublicData,
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
				fmt.Sprintf("there is no user with id: %s", id),
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

func UserUpdate(db storage.Users, onlineUsers storage.OnlineUsers) horror.HandlerFunc {
	type payload struct {
		Private *bool `json:"priv,omitempty"`
	}

	type response struct {
		payload
		models.UserPublicData
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
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

		res := new(response)
		res.payload = *p

		err = db.Update(ctx, userID, func(u *storage.UserEntry) error {
			u.Private = *p.Private
			res.UserPublicData = models.UserPublicData{
				ID:       u.ID,
				Nickname: u.Nickname,
			}
			return nil
		})
		switch {
		case errors.Is(err, serrors.ErrNoID):
			return errFactory.NotFound(
				fmt.Errorf("db.Read: %w", err),
				fmt.Sprintf("there is no user with id: %s", userID),
			)
		case err != nil:
			return errFactory.InternalServerError(
				fmt.Errorf("db.Read: %w", err),
				internalServerErrorResponse,
			)
		}

		isOnline, err := onlineUsers.IsOnline(ctx, userID)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("onlineUsers.IsOnline: %w", err),
				internalServerErrorResponse,
			)
		}
		res.Online = isOnline

		return happier.OK(w, r, res)
	}
}

// UpdateUserPassword updates password of given user after
// successfully authentication of previous one.
func UpdateUserPassword(db storage.Users) horror.HandlerFunc {
	type payload struct {
		Old string `json:"old"`
		New string `json:"new"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		errFactory := happier.FromRequest(r)

		p := new(payload)
		if err := json.NewDecoder(r.Body).Decode(p); err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("json.NewDecoder.Decode: %w", err),
				fmt.Sprintf("Invalid input: %s.", err.Error()),
			)
		}

		if ok := users.VerifyPassword(p.New); !ok {
			return errFactory.BadRequest(
				users.ErrInvaliPassword,
				fmt.Sprintf("Invalid password."),
			)
		}

		userID, err := requests.UserID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.UserID: %w", err),
				internalServerErrorResponse,
			)
		}

		_, err = users.AuthenticateWithPassword(ctx, users.AuthenticationDependencies{
			Request: users.AuthenticationRequest{
				UserID:   userID,
				Password: []byte(p.Old),
			},
			Storage:      db,
			ErrorFactory: errFactory,
		})
		if err != nil {
			return errFactory.Unauthorized(
				fmt.Errorf("users.AuthenticateWithPassword: %w", err),
				"Invalid old password.",
			)
		}

		newPass, err := bcrypt.GenerateFromPassword([]byte(p.New), bcrypt.DefaultCost)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("api.UserCreate: hashing password failed: %w", err),
				internalServerErrorResponse,
			)
		}

		err = db.Update(ctx, userID, func(u *storage.UserEntry) error {
			u.HashedPassword = newPass
			return nil
		})
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.Update: %w", err),
				internalServerErrorResponse,
			)
		}

		return happier.NoContent(w, r)
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
	ID  string `json:"id"`
	Tag string `json:"tag"`
}

// DeviceAdd handles creation of new device for requesting user.
func DeviceAdd(renewer session.Renewer, db storage.Devices) horror.HandlerFunc {
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

		state, err := renewer.Renew(r)
		if err != nil {
			// At this point handler should have
			// been provided with session, so we
			// will just return 500.
			return errFactory.InternalServerError(
				fmt.Errorf("renewer.Renew: %w", err),
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

		newID, err := devices.Add(r.Context(), devices.AddDeviceRequest{
			OwnerID: userID,
			Owner:   state.Nickname,
			Tag:     p.Tag,
			MAC:     p.MAC,
			Storage: db,
		})
		if err != nil {
			return fmt.Errorf("devices.Add: %w", err)
		}

		return happier.Created(w, r, &singleDevice{
			ID:  newID,
			Tag: p.Tag,
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
				fmt.Sprintf("there is no user with id: %s", userID),
			)
		}

		result := make([]singleDevice, len(devices), cap(devices))
		for i, device := range devices {
			result[i] = singleDevice{device.ID, device.Tag}
		}

		return happier.OK(w, r, result)
	}
}

func sameOwner(userID, deviceOwnerID, stateUserID string) bool {
	return (userID == deviceOwnerID) && (deviceOwnerID == stateUserID)
}

func DeviceRead(renewer session.Renewer, db storage.Devices) horror.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		deviceID, err := requests.DeviceID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.DeviceID: %w", err),
				internalServerErrorResponse,
			)
		}

		userID, err := requests.UserID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.UserID: %w", err),
				internalServerErrorResponse,
			)
		}

		state, err := renewer.Renew(r)
		if err != nil {
			// At this point handler should have
			// been provided with session, so we
			// will just return 500.
			if err != nil {
				return errFactory.InternalServerError(
					fmt.Errorf("renewer.Renew: %w", err),
					internalServerErrorResponse,
				)
			}
		}

		device, err := db.Read(r.Context(), deviceID)
		if errors.Is(err, serrors.ErrNoID) {
			return errFactory.NotFound(
				fmt.Errorf("db.Read: %w", err),
				fmt.Sprintf("there is no device with given id: %s", deviceID),
			)
		}
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.Read: %w", err),
				internalServerErrorResponse,
			)
		}

		// Check if requesting user owns resources.
		if !sameOwner(userID, device.OwnerID, state.UserID) {
			return errFactory.NotFound(
				fmt.Errorf("sameOwner error: userID=%s, deviceOwnerID=%s, stateUserID=%s",
					userID, device.OwnerID, state.UserID),
				fmt.Sprintf("you don't have device with id=%s", deviceID),
			)
		}

		return happier.OK(w, r, &singleDevice{
			ID:  device.ID,
			Tag: device.Tag,
		})
	}
}

func DeviceRemove(renewer session.Renewer, db storage.Devices) horror.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		deviceID, err := requests.DeviceID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.DeviceID: %w", err),
				internalServerErrorResponse,
			)
		}

		userID, err := requests.UserID(r)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("requests.UserID: %w", err),
				internalServerErrorResponse,
			)
		}

		state, err := renewer.Renew(r)
		if err != nil {
			// At this point handler should have
			// been provided with session, so we
			// will just return 500.
			if err != nil {
				return errFactory.InternalServerError(
					fmt.Errorf("renewer.Renew: %w", err),
					internalServerErrorResponse,
				)
			}
		}

		device, err := db.Read(r.Context(), deviceID)
		if errors.Is(err, serrors.ErrNoID) {
			return errFactory.NotFound(
				fmt.Errorf("db.Read: %w", err),
				fmt.Sprintf("there is no device with given id: %s", deviceID),
			)
		}
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.Read: %w", err),
				internalServerErrorResponse,
			)
		}

		// Check if requesting user owns resources.
		if !sameOwner(userID, device.OwnerID, state.UserID) {
			return errFactory.NotFound(
				fmt.Errorf("sameOwner error: userID=%s, deviceOwnerID=%s, stateUserID=%s",
					userID, device.OwnerID, state.UserID),
				fmt.Sprintf("you don't have device with id=%s", deviceID),
			)
		}

		err = db.Remove(r.Context(), deviceID)
		if errors.Is(err, serrors.ErrNoID) {
			notFound(w)
			return errFactory.NotFound(
				fmt.Errorf("db.Remove: %w", err),
				fmt.Sprintf("there is no device with given id: %s", deviceID),
			)
		}
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.Remove: %w", err),
				internalServerErrorResponse,
			)
		}

		return happier.NoContent(w, r)
	}
}
