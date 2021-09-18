package handlers

import (
	"net/http"

	"github.com/alioygur/gores"

	"github.com/hakierspejs/long-season/pkg/services/result"
	"github.com/hakierspejs/long-season/pkg/services/session"
)

// Who is handler, which returns JSON with user data
// used for authentication.
func Who(renewer session.Renewer) http.HandlerFunc {
	type response struct {
		ID       string `json:"id"`
		Nickname string `json:"nickname"`
		Private  bool   `json:"priv"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		state, err := renewer.Renew(r)
		if err != nil {
			result.JSONError(w, &result.JSONErrorBody{
				Message: "Failed to authorize user.",
				Code:    http.StatusUnauthorized,
				Type:    "unauthorized",
			})
			return
		}

		gores.JSON(w, http.StatusOK, &response{
			ID:       state.UserID,
			Nickname: state.Nickname,
		})
	}
}
