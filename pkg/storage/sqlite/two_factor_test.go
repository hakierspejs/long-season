package sqlite

import (
	"context"
	"testing"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/models/set"
	"github.com/hakierspejs/long-season/pkg/storage"
	"github.com/matryer/is"
)

func TestTwoFactor(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()

	f, closer, err := NewFactory(":memory:")
	is.NoErr(err)
	defer closer()

	usersData := map[string]storage.UserEntry{
		"1": {
			ID:             "1",
			Nickname:       "johnny",
			HashedPassword: []byte("71Hk4Rt2WY8xqgYoKxPm"),
			Private:        false,
		},
		"2": {
			ID:             "2",
			Nickname:       "marco",
			HashedPassword: []byte("u8dXHRi0JNo23JVeHkjh"),
			Private:        true,
		},
	}

	su := f.Users()
	for _, u := range usersData {
		id, err := su.New(ctx, u)
		is.NoErr(err)
		is.Equal(id, u.ID)
	}

	twoFactorMethods := map[string]models.TwoFactor{
		"1": {
			OneTimeCodes: map[string]models.OneTimeCode{
				"1": {
					ID:     "1",
					Name:   "otp1",
					Secret: "otps1",
				},
				"2": {
					ID:     "2",
					Name:   "otp2",
					Secret: "otps2",
				},
			},
			RecoveryCodes: map[string]models.Recovery{
				"1": {
					ID:    "1",
					Name:  "rec1",
					Codes: set.StringFromSlice([]string{"code1", "code2", "code3"}),
				},
			},
		},
		"2": {
			OneTimeCodes: map[string]models.OneTimeCode{
				"3": {
					ID:     "3",
					Name:   "otp3",
					Secret: "otps3",
				},
			},
			RecoveryCodes: map[string]models.Recovery{
				"2": {
					ID:    "2",
					Name:  "rec2",
					Codes: set.StringFromSlice([]string{"code1", "code2", "code3"}),
				},
			},
		},
	}

	stf := f.TwoFactor()
	for userID, methods := range twoFactorMethods {
		err = stf.Update(ctx, userID, func(tf *models.TwoFactor) error {
			for id, otp := range methods.OneTimeCodes {
				tf.OneTimeCodes[id] = otp
			}

			for id, rec := range methods.RecoveryCodes {
				tf.RecoveryCodes[id] = rec
			}
			return nil
		})
		is.NoErr(err)
	}

	johnnyTwoFactor, err := stf.Get(ctx, "1")
	is.NoErr(err)

	for otpID, otp := range johnnyTwoFactor.OneTimeCodes {
		currOtp, ok := twoFactorMethods["1"].OneTimeCodes[otpID]
		is.True(ok)
		is.Equal(otp.ID, currOtp.ID)
		is.Equal(otp.Name, currOtp.Name)
		is.Equal(otp.Secret, currOtp.Secret)
	}

	for recoveryID, recovery := range johnnyTwoFactor.RecoveryCodes {
		currRecovery, ok := twoFactorMethods["1"].RecoveryCodes[recoveryID]
		is.True(ok)
		is.Equal(recovery.ID, currRecovery.ID)
		is.Equal(recovery.Name, currRecovery.Name)
		is.True(recovery.Codes.Equals(*currRecovery.Codes))
	}

	marcoTwoFactor, err := stf.Get(ctx, "2")
	is.NoErr(err)

	for otpID, otp := range marcoTwoFactor.OneTimeCodes {
		currOtp, ok := twoFactorMethods["2"].OneTimeCodes[otpID]
		is.True(ok)
		is.Equal(otp.ID, currOtp.ID)
		is.Equal(otp.Name, currOtp.Name)
		is.Equal(otp.Secret, currOtp.Secret)
	}

	for recoveryID, recovery := range marcoTwoFactor.RecoveryCodes {
		currRecovery, ok := twoFactorMethods["2"].RecoveryCodes[recoveryID]
		is.True(ok)
		is.Equal(recovery.ID, currRecovery.ID)
		is.Equal(recovery.Name, currRecovery.Name)
		is.True(recovery.Codes.Equals(*currRecovery.Codes))
	}

	err = stf.Remove(ctx, "2")
	is.NoErr(err)

	marcoTwoFactor, err = stf.Get(ctx, "2")
	is.NoErr(err)
	is.Equal(len(marcoTwoFactor.OneTimeCodes), 0)
	is.Equal(len(marcoTwoFactor.RecoveryCodes), 0)

	err = stf.Update(ctx, "1", func(tf *models.TwoFactor) error {
		tf.OneTimeCodes["n1"] = models.OneTimeCode{
			ID:     "n1",
			Name:   "new1",
			Secret: "new1s",
		}
		tf.RecoveryCodes["n2"] = models.Recovery{
			ID:    "n2",
			Name:  "new2",
			Codes: set.StringFromSlice([]string{"code1", "code2", "code3"}),
		}
		return nil
	})
	is.NoErr(err)

	johnnyTwoFactor, err = stf.Get(ctx, "1")
	is.NoErr(err)

	new1, ok := johnnyTwoFactor.OneTimeCodes["n1"]
	is.True(ok)
	is.Equal(new1.ID, "n1")
	is.Equal(new1.Name, "new1")
	is.Equal(new1.Secret, "new1s")

	new2, ok := johnnyTwoFactor.RecoveryCodes["n2"]
	is.True(ok)
	is.Equal(new2.ID, "n2")
	is.Equal(new2.Name, "new2")
	is.True(new2.Codes.Contains("code1"))
	is.True(new2.Codes.Contains("code2"))
	is.True(new2.Codes.Contains("code3"))
}
