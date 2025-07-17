package entity

import "gorm.io/datatypes"

// Settings represents a singleton settings object in the database.
// There is always only one instance with ID=1.
type Settings struct {
	ID   uint           `json:"id" gorm:"primaryKey"`
	Data datatypes.JSON `json:"data" gorm:"type:jsonb" swaggertype:"object" example:"{\"about_us\":\"some text\"}"`
}
