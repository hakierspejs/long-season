// Package toussaint implements service logic for manipulating two
// factor methods.
package toussaint

import "github.com/hakierspejs/long-season/pkg/models"

// Method implements method function which returns
// query-able two factor data.
type Method interface {
	Method(userID string) models.TwoFactorMethod
}

// CollectMethods build slice of valeus that implements
// Method interface from multiple slices of Methods.
func CollectMethods(userID string, tf models.TwoFactor) []Method {
	res := []Method{}

	for _, v := range tf.OneTimeCodes {
		res = append(res, v)
	}

	return res
}

// Find returns first two factor method for user with given user id
// which is predictable by given function.
func Find(s []Method, userID string, f func(m models.TwoFactorMethod) bool) *models.TwoFactorMethod {
	methods := make([]models.TwoFactorMethod, len(s), len(s))
	for i, v := range s {
		methods[i] = v.Method(userID)
	}

	for _, v := range methods {
		if f(v) {
			return &v
		}
	}

	return nil
}
