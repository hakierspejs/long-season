package models

import (
	"encoding/json"
	"testing"

	"github.com/matryer/is"
)

func TestParsingTwoFactor(t *testing.T) {
	is := is.New(t)
	dat := []byte(`{
		"oneTimeCodes": {
		"1": {
			"id": "someid",
			"name": "somename",
			"randomKey": "randomValue"
		},
		"2": {
			"name": "sda",
			"secret": "somesecret"
		}
		}
	}`)
	res := new(TwoFactor)
	is.NoErr(json.Unmarshal(dat, res))
	is.Equal(res.OneTimeCodes["1"].ID, "someid")
	is.Equal(res.OneTimeCodes["1"].Name, "somename")
	is.Equal(res.OneTimeCodes["2"].Name, "sda")
	is.Equal(res.OneTimeCodes["2"].Secret, "somesecret")
}
