package handlers

import (
	"net/http"

	"github.com/alioygur/gores"

	"github.com/hakierspejs/long-season/pkg/services/requests"
	"github.com/hakierspejs/long-season/pkg/services/result"
)

// Who is handler, which returns JSON with user data
// used for authentication.
func Who() http.HandlerFunc {
	type response struct {
		ID       int    `json:"id"`
		Nickname string `json:"nickname"`
		Private  bool   `json:"private"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := requests.JWTClaims(r)
		if err != nil {
			result.JSONError(w, &result.JSONErrorBody{
				Message: "You have to provide correct bearer token.",
				Code:    http.StatusUnauthorized,
				Type:    "unauthorized",
			})
			return
		}

		gores.JSON(w, http.StatusOK, &response{
			ID:       claims.UserID,
			Nickname: claims.Nickname,
			Private:  claims.Private,
		})
	}
}
