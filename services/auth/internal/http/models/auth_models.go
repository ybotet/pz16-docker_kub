package models

// LoginRequest representa la petición de login
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

// LoginResponse representa la respuesta de login
type LoginResponse struct {
    Token     string `json:"token"`
    CSRFToken string `json:"csrf_token"`
}

// ErrorResponse representa un error HTTP
type ErrorResponse struct {
    Error string `json:"error"`
}