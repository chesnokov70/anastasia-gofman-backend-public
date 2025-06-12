package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// @name TranslatedText
type TranslatedText struct {
	EN string `json:"en" example:"Tung, tung, tung, tung, tung, tung, tung, tung, tung, Sahur."`
	RU string `json:"ru" example:"Тун, тун, тун, Сахур."`
	ES string `json:"es" example:"Golybini Shpionini"`
}

// Scan implements the sql.Scanner interface.
func (t *TranslatedText) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &t)
}

// Value implements the driver.Valuer interface.
func (t TranslatedText) Value() (driver.Value, error) {
	return json.Marshal(t)
}
