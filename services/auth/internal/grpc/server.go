package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/ybotet/pz6_csrf_xss/gen/proto/auth"
	"github.com/ybotet/pz6_csrf_xss/services/auth/internal/auth"
)

type Server struct {
    authpb.UnimplementedAuthServiceServer
    authService *auth.Service
}

func NewServer(authService *auth.Service) *Server {
    return &Server{
        authService: authService,
    }
}

func (s *Server) Verify(ctx context.Context, req *authpb.VerifyRequest) (*authpb.VerifyResponse, error) {
    log.Printf("Recibida solicitud Verify para token: %s", req.Token[:10]+"...")

    if req.Token == "" {
        return nil, status.Error(codes.Unauthenticated, "token vacío")
    }

    subject, err := s.authService.VerifyToken(req.Token)
    if err != nil {
        log.Printf("Error verificando token: %v", err)
        return nil, status.Error(codes.Unauthenticated, "token inválido")
    }

    return &authpb.VerifyResponse{
        Valid:   true,
        Subject: subject,
    }, nil
}

func Register(s *grpc.Server, authService *auth.Service) {
    authpb.RegisterAuthServiceServer(s, NewServer(authService))
}
