package mgrpc

import (
	"context"
	"errors"
	"log"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/jettjia/igo-pkg/pkg/xmiddleware/jwt"
)

// AuthInterceptor
func AuthInterceptor(ctx context.Context, jwtSecret []byte) (context.Context, error) {
	var (
		err error
	)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("AuthInterceptor:metadata:error")
		return ctx, errors.New("AuthInterceptor:metadata:error")
	}

	if len(md.Get("authorization")) == 0 {
		return ctx, err
	}

	token := md.Get("authorization")[0]
	token = strings.TrimPrefix(token, "Bearer ")
	claims, _ := jwt.ParseToken(token, jwtSecret)
	if claims == nil {
		return ctx, err
	}

	return ctx, nil
}
