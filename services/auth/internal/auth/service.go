package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
    jwtSecret []byte
}

func NewService(secret string) *Service {
    return &Service{
        jwtSecret: []byte(secret),
    }
}

// VerifyToken проверяет JWT токен и возвращает subject если он действителен
func (s *Service) VerifyToken(tokenString string) (string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("недопустимый метод подписи")
        }
        return s.jwtSecret, nil
    })

    if err != nil {
        return "", err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if exp, ok := claims["exp"].(float64); ok {
            if time.Now().Unix() > int64(exp) {
                return "", errors.New("срок действия токена истек")
            }
        }
        subject, _ := claims["sub"].(string)
        if subject == "" {
            return "", errors.New("subject не найден в токене")
        }
        return subject, nil
    }

    return "", errors.New("недействительный токен")
}

// GenerateToken создает новый JWT токен для тестирования
func (s *Service) GenerateToken(subject string, duration time.Duration) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub": subject,
        "exp": time.Now().Add(duration).Unix(),
        "iat": time.Now().Unix(),
    })
    return token.SignedString(s.jwtSecret)
}