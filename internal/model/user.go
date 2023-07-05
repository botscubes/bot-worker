package model

import "github.com/goccy/go-json"

type UserStatus int

var (
	StatusUserActive UserStatus
)

type User struct {
	Id        int64   `json:"id"`
	TgId      int64   `json:"tgId"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Username  *string `json:"username"`
	StepID
	Status UserStatus `json:"-"`
}

type StepID struct {
	StepId int64 `json:"stepId"`
}

// Encode component struct to binary format (for redis)
func (c *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

// Decode component from binary format to struct (fo redis)
func (c *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &c)
}
