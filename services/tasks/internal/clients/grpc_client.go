package clients

import (
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/ybotet/pz6_csrf_xss/gen/proto/auth"
)

type AuthClient struct {
    client authpb.AuthServiceClient
    conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
    conn, err := grpc.Dial(addr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
        grpc.WithTimeout(5*time.Second),
    )
    if err != nil {
        return nil, err
    }

    client := authpb.NewAuthServiceClient(conn)
    log.Printf("Cliente gRPC conectado a Auth service en %s", addr)

    return &AuthClient{
        client: client,
        conn:   conn,
    }, nil
}

func (c *AuthClient) Close() error {
    return c.conn.Close()
}

func (c *AuthClient) GetClient() authpb.AuthServiceClient {
    return c.client
}
