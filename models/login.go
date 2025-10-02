package models

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Username string  `json:"username"`
	Bio      *string `json:"bio"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	Password string `json:"password"`
}
