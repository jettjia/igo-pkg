package mgrpc

import (
	"fmt"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jettjia/igo-pkg/pkg/xerror"
)

func RecoverInterceptor() grpc_recovery.Option {
	return grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
		if err != nil {
			xerror.RecoverWithStack()
		}

		return status.ErrorProto(&spb.Status{Code: int32(codes.Internal), Message: fmt.Sprintf("%v", p), Details: nil})
	})
}
