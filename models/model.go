package models

type User struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Age     int64   `json:"age"`
	Address string  `json:"address"`
	Phone   string  `json:"phone"`
	Hobby   []Hobby `json:"hobby"`
}

type Hobby struct {
	Id        string `json:"id"`
	UserId    string `json:"userId"`
	HobbyName string `json:"hobbyName"`
}

type UserLogin struct {
	Id       string `json:"id"`
	UserId   string `json:"userId"`
	Username string `json:"username"`
	Password []byte `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
