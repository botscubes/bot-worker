package model

import (
	"github.com/goccy/go-json"
)

type ComponentStatus int

var (
	StatusComponentActive ComponentStatus
	StatusComponentDel    ComponentStatus = 1
)

type CommandStatus int

var (
	StatusCommandActive CommandStatus
	StatusCommandDel    CommandStatus = 1
)

type Component struct {
	Id         int64     `json:"id"`
	Data       *Data     `json:"data"`
	Keyboard   *Keyboard `json:"keyboard"`
	Commands   *Commands `json:"commands"`
	NextStepId *int64    `json:"nextStepId"`
	IsMain     bool      `json:"isMain"`
}

type Commands []*Command

type Data struct {
	Type    *string     `json:"type"`
	Content *[]*Content `json:"content"`
}

type Content struct {
	Text *string `json:"text,omitempty"`
}

type Keyboard struct {
	Buttons [][]*int64 `json:"buttons"`
}

type Command struct {
	Id          *int64        `json:"id"`
	Type        *string       `json:"type"`
	Data        *string       `json:"data"`
	ComponentId *int64        `json:"componentId"`
	NextStepId  *int64        `json:"nextStepId"`
	Status      CommandStatus `json:"status"`
}

// Encode component struct to binary format (for redis)
func (c *Component) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

// Decode component from binary format to struct (fo redis)
func (c *Component) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &c)
}
