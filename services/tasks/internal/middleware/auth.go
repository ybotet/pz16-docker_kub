package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	authpb "github.com/ybotet/pz6_csrf_xss/gen/proto/auth"
)

type contextKey string

const (
    UserIDKey contextKey = "user_id"
)

// AuthMiddleware autentica por JWT (del header Authorization)
type AuthMiddleware struct {
    authClient authpb.AuthServiceClient
    logger     *logrus.Logger
}

func NewAuthMiddleware(authClient authpb.AuthServiceClient) *AuthMiddleware {
    return &AuthMiddleware{
        authClient: authClient,
    }
}

// Authenticate valida el token JWT
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Obtener token del header Authorization
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            http.Error(w, `{"error":"Token no proporcionado"}`, http.StatusUnauthorized)
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        // Verificar token con auth service
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        resp, err := m.authClient.Verify(ctx, &authpb.VerifyRequest{
            Token: token,
        })
        if err != nil || !resp.Valid {
            http.Error(w, `{"error":"Token inválido"}`, http.StatusUnauthorized)
            return
        }

        // Añadir user ID al contexto
        ctx = context.WithValue(r.Context(), UserIDKey, resp.Subject)
        next(w, r.WithContext(ctx))
    }
}

// GetUserID obtiene el user ID del contexto
func GetUserID(ctx context.Context) string {
    if id, ok := ctx.Value(UserIDKey).(string); ok {
        return id
    }
    return ""
}