package models

type User struct {
    ID     string `json:"id,omitempty" bson:"_id,omitempty"`
    Token  string `json:"token"`
    Mobile string `json:"mobile"`
}
