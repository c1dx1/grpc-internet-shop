package interceptors

import (
	"context"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"internet-shop/services/user-service/sessions"
)

func SessionAuthInterceptor(redisClient *redis.Client) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if info.FullMethod == "/proto.UserService/SignInUser" || info.FullMethod == "/proto.UserService/SignUpUser" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}

		sessionIDs := md["session-id"]
		if len(sessionIDs) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing session id")
		}

		userID, err := sessions.GetUserIDFromSession(ctx, redisClient, sessionIDs[0])
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid session id")
		}

		valid, err := ValidateSession(ctx, redisClient, sessionIDs[0])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error validating session")
		}
		if !valid {
			return nil, status.Errorf(codes.Unauthenticated, "invalid or expired session id")
		}

		newCtx := context.WithValue(ctx, "user-id", userID)

		return handler(newCtx, req)
	}
}

func ValidateSession(ctx context.Context, redisClient *redis.Client, sessionID string) (bool, error) {
	_, err := redisClient.Get(ctx, sessionID).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}

	expTime, err := redisClient.TTL(ctx, sessionID).Result()
	if err != nil {
		return false, err
	}

	if expTime <= 0 {
		return false, nil
	}

	return true, nil
}
