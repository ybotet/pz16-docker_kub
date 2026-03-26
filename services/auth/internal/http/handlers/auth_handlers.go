package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus" // IMPORTANTE: importar logrus
	"github.com/ybotet/pz6_csrf_xss/services/auth/internal/auth"
	httpModels "github.com/ybotet/pz6_csrf_xss/services/auth/internal/http/models"
)

type AuthHandler struct {
    authService  *auth.Service
    logger       *logrus.Logger  // Cambiado a *logrus.Logger
    sessionStore map[string]string
}

func NewAuthHandler(authService *auth.Service, logger *logrus.Logger) *AuthHandler {
    return &AuthHandler{
        authService:  authService,
        logger:       logger,
        sessionStore: make(map[string]string),
    }
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // 1. Decodificar request
    var req httpModels.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.WithError(err).Error("Error decodificando login request")
        h.respondWithError(w, http.StatusBadRequest, "Invalid request")
        return
    }

    // 2. Validar credenciales (temporal)
    if req.Username == "" {
        h.respondWithError(w, http.StatusBadRequest, "Username required")
        return
    }

    // 3. Generar JWT
    token, err := h.authService.GenerateToken(req.Username, 24*time.Hour)
    if err != nil {
        h.logger.WithError(err).Error("Error generando token")
        h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
        return
    }

    // 4. Generar tokens para cookies
    sessionID := h.generateSecureToken(32)
    csrfToken := h.generateSecureToken(32)
    
    // Guardar sesión
    h.sessionStore[sessionID] = req.Username

    // 5. Establecer cookies
    h.setSessionCookie(w, sessionID)
    h.setCSRFCookie(w, csrfToken)

    // 6. Responder
    h.respondWithJSON(w, http.StatusOK, httpModels.LoginResponse{
        Token:     token,
        CSRFToken: csrfToken,
    })
    
    h.logger.WithFields(logrus.Fields{
        "username": req.Username,
    }).Info("Login exitoso")
}

// Health check
func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
    h.respondWithJSON(w, http.StatusOK, map[string]string{
        "status":  "ok",
        "service": "auth",
    })
}

// Métodos privados
func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, sessionID string) {
    http.SetCookie(w, &http.Cookie{
        Name:     "session",
        Value:    sessionID,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
        Path:     "/",
        MaxAge:   86400,
    })
}

func (h *AuthHandler) setCSRFCookie(w http.ResponseWriter, csrfToken string) {
    http.SetCookie(w, &http.Cookie{
        Name:     "csrf_token",
        Value:    csrfToken,
        HttpOnly: false,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
        Path:     "/",
        MaxAge:   86400,
    })
}

func (h *AuthHandler) generateSecureToken(length int) string {
    b := make([]byte, length)
    if _, err := rand.Read(b); err != nil {
        h.logger.WithError(err).Error("Error generando token seguro")
        return ""
    }
    return hex.EncodeToString(b)
}

func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(payload)
}

func (h *AuthHandler) respondWithError(w http.ResponseWriter, code int, message string) {
    h.respondWithJSON(w, code, httpModels.ErrorResponse{Error: message})
}