package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"golang.org/x/crypto/bcrypt"
	"internet-shop/repository"
	"internet-shop/services/user-service/sessions"
	"internet-shop/shared/proto"
	"log"
)

type UserHandler struct {
	repo        repository.UserRepository
	rabbitCh    *amqp.Channel
	redisClient *redis.Client
	proto.UnimplementedUserServiceServer
}

func NewUserHandler(repo repository.UserRepository, rabbitCh *amqp.Channel, redisClient *redis.Client) *UserHandler {
	return &UserHandler{repo: repo, rabbitCh: rabbitCh, redisClient: redisClient}
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

	message := map[string]interface{}{
		"user_id": userID,
	}

	messageBody, _ := json.Marshal(message)

	err = h.rabbitCh.Publish(
		"",
		"user_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})

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

func (h *UserHandler) GetEmailById(ctx context.Context, req *proto.IdRequest) (*proto.EmailResponse, error) {
	email, err := h.repo.GetEmailById(ctx, req.Id)
	if err != nil {
		log.Printf("user_handler: get email by id %s error %v", req.Id, err)
		return nil, err
	}
	return &proto.EmailResponse{Email: email}, nil
}
