// Package requests contains utilities for retrieving
// data stored in http.Request reference.
package requests

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/config"
)

// UserID returns user id from url.
func UserID(r *http.Request) (int, error) {
	id, ok := r.Context().Value("user-id").(string)
	if !ok {
		return 0, errors.New("ID stored in context has inapropriate type.")
	}

	res, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func JWTClaims(r *http.Request) (*models.Claims, error) {
	claims, ok := r.Context().Value(config.JWTUserKey).(*models.Claims)
	if !ok {
		return nil, fmt.Errorf("long-season: there are no jwt claims in context")
	}
	return claims, nil
}
