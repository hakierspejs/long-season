package models

import "fmt"

// Config represents configuration that is
// being used by server.
type Config struct {
	Host         string
	Port         string
	DatabasePath string
}

// Address returns address string that is compatible
// with http.ListenAndServe function.
func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
