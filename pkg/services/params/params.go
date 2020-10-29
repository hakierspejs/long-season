package params

import (
	"errors"
	"net/http"
	"strconv"
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
