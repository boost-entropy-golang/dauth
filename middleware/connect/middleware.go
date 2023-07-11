package server

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"
	"github.com/streamingfast/dauth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ connect.Interceptor = (*AuthInterceptor)(nil)

func NewAuthInterceptor(check dauth.Authenticator, logger *zap.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		check:  check,
		logger: logger,
	}
}

type AuthInterceptor struct {
	check  dauth.Authenticator
	logger *zap.Logger
}

// WrapUnary implements [Interceptor] by applying the interceptor function.
func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {

		peerAddr := req.Peer().Addr
		headers := req.Header()
		path := req.Spec().Procedure

		childCtx, err := validateAuth(ctx, path, headers, peerAddr, i.check)
		if err != nil {
			return nil, obfuscateErrorMessage(err, i.logger)
		}

		return next(childCtx, req)
	}
}

func (i *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {

		peerAddr := conn.Peer().Addr
		headers := conn.RequestHeader()
		path := conn.Spec().Procedure

		childCtx, err := validateAuth(ctx, path, headers, peerAddr, i.check)
		if err != nil {
			return obfuscateErrorMessage(err, i.logger)
		}

		return next(childCtx, conn)
	}
}

// Noop
func (i *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func obfuscateErrorMessage(err error, logger *zap.Logger) error {
	if st, ok := status.FromError(err); ok {
		msg := st.Message()
		switch st.Code() {
		case codes.Internal, codes.Unavailable, codes.Unknown:
			logger.Error("authentication service via Connect-Web middleware fatal error", zap.Error(err))
			msg = "error with authentication service, please try again later"
		}
		return connect.NewError(connect.Code(st.Code()), errors.New(msg))
	} else {
		logger.Error("authentication service via Connect-Web middleware non-gRPC error", zap.Error(err))
	}

	return err
}
