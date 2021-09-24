package api

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/models/set"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/requests"
	"github.com/hakierspejs/long-season/pkg/services/session"
	"github.com/hakierspejs/long-season/pkg/services/toussaint"
	"github.com/hakierspejs/long-season/pkg/storage"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/thinkofher/horror"
)

// TwoFactorMethods handler returns list of enabled two factor methods.
// Make sure to make this resource private befour mounting to some mux or
// router.
func TwoFactorMethods(db storage.TwoFactor) horror.HandlerFunc {
	type response struct {
		Active []models.TwoFactorMethod `json:"active"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		userID, err := requests.UserID(r)
		if err != nil {
			errFactory.Unauthorized(
				fmt.Errorf("requests.UserID: %w", err),
				"You don't have access to given resources.",
			)
		}

		methods, err := db.Get(r.Context(), userID)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("db.Get: %w", err),
				internalServerErrorResponse,
			)
		}

		res := new(response)
		for _, m := range methods.OneTimeCodes {
			res.Active = append(res.Active, m.Method(userID))
		}

		return happier.OK(w, r, res)
	}
}

// userAndTwoFactorID returns respectively user ID, two factor ID and
// error if there was a failure in the process of parsing.
//
// Returns output safe errors for horror Handlers.
func userAndTwoFactorID(r *http.Request) (string, string, error) {
	errFactory := happier.FromRequest(r)

	userID, err := requests.UserID(r)
	if err != nil {
		return "", "", errFactory.BadRequest(
			fmt.Errorf("requests.UserID: %w", err),
			"Missing user ID.",
		)
	}

	twoFactorID, err := requests.TwoFactorID(r)
	if err != nil {
		return "", "", errFactory.BadRequest(
			fmt.Errorf("requests.TwoFactorID: %w", err),
			"Missing TwoFactor method's ID",
		)
	}

	return userID, twoFactorID, nil
}

// TwoFactorMethod return two factor method resource with given
// two factor ID for user with given user ID.
//
// Make sure to make this resource private befour mounting to some mux or
// router.
func TwoFactorMethod(db storage.TwoFactor) horror.HandlerFunc {
	type response struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Location string `json:"location"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		userID, twoFactorID, err := userAndTwoFactorID(r)
		if err != nil {
			return fmt.Errorf("userAndTwoFactorID: %w", err)
		}

		twoFactor, err := db.Get(r.Context(), userID)
		if err != nil {
			return errFactory.NotFound(
				fmt.Errorf("db.Get: %w", err),
				"Failed to fetch two factor methods for user.",
			)
		}

		methods := toussaint.CollectMethods(userID, *twoFactor)
		res := toussaint.Find(methods, userID, func(tf models.TwoFactorMethod) bool {
			return tf.ID == twoFactorID
		})
		if res == nil {
			return errFactory.NotFound(
				fmt.Errorf("There is no two factor method with the given ID."),
				"There is no two factor method with the given ID.",
			)
		}

		return happier.OK(w, r, &response{
			ID:       res.ID,
			Name:     res.Name,
			Type:     string(res.Type),
			Location: res.Location,
		})
	}
}

// TwoFactorMethodRemove deletes two factor method with given ID.
//
// Make sure to make this resource private befour mounting to some mux or
// router.
func TwoFactorMethodRemove(db storage.TwoFactor) horror.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		userID, twoFactorID, err := userAndTwoFactorID(r)
		if err != nil {
			return fmt.Errorf("userAndTwoFactorID: %w", err)
		}

		err = db.Update(r.Context(), userID, func(tf *models.TwoFactor) error {
			methods := toussaint.CollectMethods(userID, *tf)
			res := toussaint.Find(methods, userID, func(tf models.TwoFactorMethod) bool {
				return tf.ID == twoFactorID
			})
			if res == nil {
				return errFactory.NotFound(
					fmt.Errorf("There is no two factor method with the given ID."),
					"There is no two factor method with the given ID.",
				)
			}

			delete(tf.OneTimeCodes, twoFactorID)
			delete(tf.RecoveryCodes, twoFactorID)
			return nil
		})
		if err != nil {
			return fmt.Errorf("db.Update: %w", err)
		}

		return happier.NoContent(w, r)
	}
}

// OptionsOTP returns json payload with secret string for
// OTP codes and data URI with image for authentication apps.
func OptionsOTP(config models.Config, renewer session.Renewer) horror.HandlerFunc {
	type response struct {
		Secret string `json:"secret"`
		Image  string `json:"image"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		sessionState, err := renewer.Renew(r)
		if err != nil {
			errFactory.Unauthorized(
				fmt.Errorf("renewer.Renew: %w", err),
				"You don't have access to given resources.",
			)
		}

		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      config.AppName,
			Digits:      otp.DigitsSix,
			Rand:        rand.Reader,
			Algorithm:   otp.AlgorithmSHA1,
			AccountName: sessionState.Nickname,
		})
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("totp.Generate: %w", err),
				internalServerErrorResponse,
			)
		}

		var buf bytes.Buffer
		img, err := key.Image(200, 200)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("key.Image: %w", err),
				internalServerErrorResponse,
			)
		}
		png.Encode(&buf, img)

		imageNode := fmt.Sprintf(
			"data:image/png;base64,%s",
			base64.StdEncoding.EncodeToString(buf.Bytes()),
		)

		return happier.OK(w, r, &response{
			Secret: key.Secret(),
			Image:  imageNode,
		})
	}
}

// AddOTP enables otp codes based two factor authentication with provided secret
// string and name.
func AddOTP(renewer session.Renewer, db storage.TwoFactor) horror.HandlerFunc {
	type payload struct {
		Name   string `json:"name"`
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		sessionState, err := renewer.Renew(r)
		if err != nil {
			return errFactory.Unauthorized(
				fmt.Errorf("renewer.Renew: %w", err),
				"You don't have access to given resources.",
			)
		}

		p := new(payload)
		if err := json.NewDecoder(r.Body).Decode(p); err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("json.NewDecoder.Decode: %w", err),
				internalServerErrorResponse,
			)
		}

		validated := totp.Validate(p.Code, p.Secret)
		if !validated {
			return errFactory.BadRequest(
				fmt.Errorf("failed to validate totp code"),
				"Failed to validate totp code with given secret.",
			)
		}

		id := uuid.New().String()
		if err := db.Update(r.Context(), sessionState.UserID, func(tf *models.TwoFactor) error {
			tf.OneTimeCodes[id] = models.OneTimeCode{
				ID:     id,
				Name:   p.Name,
				Secret: p.Secret,
			}
			return nil
		}); err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("json.NewDecoder.Decode: %w", err),
				internalServerErrorResponse,
			)
		}

		return happier.NoContent(w, r)
	}
}

func AddRecovery(renewer session.Renewer, db storage.TwoFactor) horror.HandlerFunc {
	type payload struct {
		Name  string   `json:"name"`
		Codes []string `json:"codes"`
	}

	const (
		maxCodesLength int = 10
		maxCodeLength  int = 20
	)

	verifyPayload := func(p *payload) (string, bool) {
		if p.Name == "" {
			return "Missing Name.", false
		}
		if len(p.Codes) > maxCodesLength {
			return "Too many codes has been sent.", false
		}
		for _, c := range p.Codes {
			if len(c) > maxCodeLength {
				return fmt.Sprintf("Maximum length of single code is %d.", maxCodeLength), false
			}
		}
		return "", true
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)
		p := new(payload)

		state, err := renewer.Renew(r)
		if err != nil {
			return errFactory.Unauthorized(
				fmt.Errorf("renewer.Renew: %w", err),
				"You don't have access to given resources.",
			)
		}

		err = json.NewDecoder(r.Body).Decode(p)
		if err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("json.NewDecoder.Decode: %w", err),
				"Failed to parse payload.",
			)
		}

		msg, ok := verifyPayload(p)
		if !ok {
			return errFactory.BadRequest(
				fmt.Errorf("verifyPayload is not ok: %s", msg),
				msg,
			)
		}

		id := uuid.New().String()
		err = db.Update(r.Context(), state.UserID, func(tf *models.TwoFactor) error {
			tf.RecoveryCodes[id] = models.Recovery{
				ID:    id,
				Name:  p.Name,
				Codes: set.StringFromSlice(p.Codes),
			}
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
