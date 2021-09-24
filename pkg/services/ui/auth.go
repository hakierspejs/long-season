package ui

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/session"
	"github.com/hakierspejs/long-season/pkg/services/toussaint"
	"github.com/hakierspejs/long-season/pkg/services/users"
	"github.com/hakierspejs/long-season/pkg/storage"

	"github.com/thinkofher/horror"
)

// AuthArguments holds dependencies for Auth handler.
type AuthArguments struct {
	Saver                session.Saver
	Users                storage.Users
	TwoFactor            storage.TwoFactor
	TwoFactorRedirectURI string
}

// Auth authentiates user with given nickanem and password
// as JSOn document in the payload.
func Auth(args AuthArguments) horror.HandlerFunc {
	type payload struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
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
			Storage:      args.Users,
			ErrorFactory: errFactory,
		})
		if err != nil {
			return fmt.Errorf("users.AuthenticateWithPassword: %w", err)
		}

		newSession := session.New(ctx, session.Builder{
			UserID:   match.ID,
			Nickname: match.Nickname,
		})

		methods, err := args.TwoFactor.Get(ctx, match.ID)
		if err != nil {
			return errFactory.InternalServerError(
				fmt.Errorf("tf.Get: %w", err),
				"Failed to find two factor methods.",
			)
		}

		twoFactorEnabled := toussaint.IsTwoFactorEnabled(*methods)
		fmt.Println(match.Nickname, "two factor enable: ", twoFactorEnabled)
		if err := session.WithOptions(ctx, *newSession, session.WithOptionsArguments{
			Saver:  args.Saver,
			Writer: w,
			Options: []session.Option{
				toussaint.TwoFactorRequired(twoFactorEnabled),
				toussaint.AuthenticationWithTOTP(len(methods.OneTimeCodes) > 0),
				toussaint.AuthenticationWithRecovery(len(methods.RecoveryCodes) > 0),
			},
		}); err != nil {
			return fmt.Errorf("session.WithOption: %w", err)
		}

		if twoFactorEnabled {
			// return happier.SeeOther(w, r, args.TwoFactorRedirectURI)
			http.Redirect(w, r, args.TwoFactorRedirectURI, http.StatusTemporaryRedirect)
			return nil
		}
		return happier.NoContent(w, r)
	}
}

// AuthWithCodesArguments holds dependencies for AuthWichCodes handler.
type AuthWithCodesArguments struct {
	Renewer   session.Renewer
	Saver     session.Saver
	TwoFactor storage.TwoFactor
}

// AuthWithCodes authenticates given user with one time passwords.
func AuthWithCodes(args AuthWithCodesArguments) horror.HandlerFunc {
	type payload struct {
		Code string `json:"code"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		errFactory := happier.FromRequest(r)

		s, err := args.Renewer.Renew(r)
		if err != nil {
			return errFactory.Unauthorized(
				fmt.Errorf("args.Renewer.Renew: %w", err),
				"You don't have access to given resources.",
			)
		}

		p := new(payload)
		if err := json.NewDecoder(r.Body).Decode(p); err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("json.NewDecoder.Decode: %w", err),
				"Failed to parse JSON payload.",
			)
		}

		methods, err := args.TwoFactor.Get(ctx, s.UserID)
		if err != nil {
			return errFactory.NotFound(
				fmt.Errorf("args.TwoFactor.Get: %w", err),
				"There are no two factor methods for given user.",
			)
		}

		validator := toussaint.ValidatorComposite(
			toussaint.ValidatorTOTP(*methods),
			toussaint.ValidatorRecovery(args.TwoFactor, s.UserID),
		)

		validated := validator.Validate(ctx, p.Code)
		if !validated {
			return errFactory.Unauthorized(
				fmt.Errorf("validated is false"),
				"Failed to authenticate with codes.",
			)
		}

		err = session.WithOptions(ctx, *s, session.WithOptionsArguments{
			Saver:  args.Saver,
			Writer: w,
			Options: []session.Option{
				toussaint.AuthenticationWithRecovery(len(methods.RecoveryCodes) > 0),
				toussaint.AuthenticationWithTOTP(len(methods.OneTimeCodes) > 0),
				// User is fully authenticated now, so we can
				// set required two factor authentication to false.
				toussaint.TwoFactorRequired(false),
			},
		})

		return happier.NoContent(w, r)
	}
}
