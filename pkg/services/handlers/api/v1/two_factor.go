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
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/requests"
	"github.com/hakierspejs/long-season/pkg/services/session"
	"github.com/hakierspejs/long-season/pkg/storage"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/thinkofher/horror"
)

// TwoFactorMethods handler returns list of enabled two factor methods.
func TwoFactorMethods(renewer session.Renewer, db storage.TwoFactor) horror.HandlerFunc {
	type response struct {
		Active []models.TwoFactorMethod `json:"active"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		errFactory := happier.FromRequest(r)

		userID, err := requests.UserID(r)
		if err != nil {
			errFactory.Unauthorized(
				fmt.Errorf("renewer.Renew: %w", err),
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

		if err := db.Update(r.Context(), sessionState.UserID, func(tf *models.TwoFactor) error {
			tf.OneTimeCodes = append(tf.OneTimeCodes, models.OneTimeCode{
				ID:     uuid.New().String(),
				Name:   p.Name,
				Secret: p.Secret,
			})
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
