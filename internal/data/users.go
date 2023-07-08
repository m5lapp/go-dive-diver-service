package data

import "github.com/m5lapp/go-service-toolkit/serialisation/jsonz"

// User represents a user's base details returned from the User service.
type User struct {
	Name         string          `json:"name"`
	FriendlyName *string         `json:"friendly_name,omitempty" mapstructure:"friendly_name"`
	BirthDate    *jsonz.DateOnly `json:"birth_date,omitempty" mapstructure:"birth_date"`
	Gender       *string         `json:"gender,omitempty"`
	CountryCode  *string         `json:"country_code,omitempty" mapstructure:"country_code"`
	TimeZone     *string         `json:"time_zone,omitempty" mapstructure:"time_zone"`
}

// UserResponse represents how a User struct is enveloped from the User service.
type UserResponse struct {
	User User `json:"user"`
}
