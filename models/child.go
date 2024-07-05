package models

type Child struct {
	ID          int    `json:"id,omitempty"`
	CenterID    int    `json:"center_id,omitempty"`
	ClassID     int    `json:"class_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Gender      string `json:"gender,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
}
