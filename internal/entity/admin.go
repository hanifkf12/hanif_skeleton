package entity

import "time"

type Admin struct {
	Id        int       `json:"id,omitempty" db:"id"`
	FirstName string    `json:"first_name,omitempty" db:"firstName"`
	LastName  string    `json:"last_name,omitempty" db:"lastName"`
	Email     string    `json:"email,omitempty" db:"email"`
	Gender    string    `json:"gender,omitempty" db:"gender"`
	BirthDate time.Time `json:"birth_date,omitempty" db:"birthDate"`
	Password  string    `json:"-" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"createdAt"`
	UpdatedAt time.Time `json:"updated_at" db:"updatedAt"`
}
