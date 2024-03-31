package data

import (
	"encoding/json"
	"io"

	"github.com/go-playground/validator/v10"
)

type URL struct {
	ID        int    `json:"id"`
	Link      string `json:"url" validate:"url,required"`
	ShortHash string `json:"short_hash,omitempty" validate:"omitempty,len=5,alphanum"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func (u *URL) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(u)
}

func (u *URL) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(u)
}

func (u *URL) Validate() error {
	v := validator.New()
	return v.Struct(u)
}
