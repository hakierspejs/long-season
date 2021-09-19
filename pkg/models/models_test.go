package models

import (
	"encoding/json"
	"testing"

	"github.com/matryer/is"
)

func TestParsingTwoFactor(t *testing.T) {
	is := is.New(t)
	dat := []byte(`{
		"oneTimeCodes": [{
			"id": "someid",
			"name": "somename",
			"randomKey": "randomValue"
		},{
			"name": "sda",
			"secret": "somesecret"
		}]
	}`)
	res := new(TwoFactor)
	is.NoErr(json.Unmarshal(dat, res))
	is.Equal(res.OneTimeCodes[0].ID, "someid")
	is.Equal(res.OneTimeCodes[0].Name, "somename")
	is.Equal(res.OneTimeCodes[1].Name, "sda")
	is.Equal(res.OneTimeCodes[1].Secret, "somesecret")
}
