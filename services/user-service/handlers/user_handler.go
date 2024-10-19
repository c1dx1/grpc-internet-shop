package handlers

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"internet-shop/repository"
	"internet-shop/services/user-service/sessions"
	"internet-shop/shared/proto"
	"log"
)

type UserHandler struct {
	repo        repository.UserRepository
	redisClient *redis.Client
	proto.UnimplementedUserServiceServer
}

func NewUserHandler(repo repository.UserRepository, redisClient *redis.Client) *UserHandler {
	return &UserHandler{repo: repo, redisClient: redisClient}
}

func (h *UserHandler) SignInUser(ctx context.Context, req *proto.SignInRequest) (*proto.SignInResponse, error) {
	userID, err := h.repo.SignInUser(ctx, req.Email, req.Password)
	if err != nil {
		log.Printf("user_handler: sign in user: error repo with data:%v, %v: err:%v", req.Email, req.Password, err)
		return nil, err
	}

	sessionID, err := sessions.CreateSession(ctx, h.redisClient, userID)
	if err != nil {
		log.Printf("user_handler: sign in user: error create session with data:%v: err:%v", userID, err)
		return nil, err
	}

	return &proto.SignInResponse{SessionId: sessionID}, nil
}

func (h *UserHandler) SignUpUser(ctx context.Context, req *proto.SignUpRequest) (*proto.SignInResponse, error) {
	if req.Password != req.RepeatPassword {
		return nil, errors.New("password does not match")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error bcrypt pass %s", err)
		return nil, err
	}

	userID, err := h.repo.SignUpUser(ctx, req.Email, string(passwordHash))
	if err != nil {
		log.Printf("error repo user %s", err)
		return nil, err
	}

	sessionID, err := sessions.CreateSession(ctx, h.redisClient, userID)
	if err != nil {
		log.Printf("error create session %s", err)
		return nil, err
	}

	return &proto.SignInResponse{SessionId: sessionID}, nil
}

func (h *UserHandler) SignOutUser(ctx context.Context, req *proto.SignOutRequest) (*proto.SignOutResponse, error) {
	err := sessions.DeleteSession(ctx, h.redisClient, req.SessionId)
	if err != nil {
		log.Printf("user_handler: sign out user: error delete session %s", err)
		return nil, err
	}

	return &proto.SignOutResponse{Success: true}, nil
}
