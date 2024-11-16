package handlers

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"internet-shop/repository"
	"internet-shop/shared/config"
	"internet-shop/shared/proto"
	"log"
)

type NotificationHandler struct {
	repo repository.NotificationRepository
	proto.UnimplementedNotificationServiceServer
}

func NewNotificationHandler(repo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

func NewUserConnection() (*grpc.ClientConn, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("notification_handler: newuc: Error loading config: %s", err)
		return nil, err
	}
	conn, err := grpc.NewClient(fmt.Sprintf("localhost%s", cfg.UserPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
		return nil, err
	}
	return conn, nil
}

func (h *NotificationHandler) GetNotifications(ctx context.Context, req *proto.GetNotificationsRequest) (*proto.NotificationsList, error) {
	notificationsDB, err := h.repo.GetNotifications(ctx, req.UserId)
	if err != nil {
		log.Printf("notif. handler: getnotif.: repo error: %v", err)
		return nil, err
	}

	var notifications []*proto.Notification
	for _, n := range notificationsDB {
		notifications = append(notifications, &proto.Notification{
			Id:        n.ID,
			UserId:    n.UserID,
			Author:    n.Author,
			Subject:   n.Subject,
			Content:   n.Content,
			CreatedAt: n.CreatedAt.String(),
		})
	}

	return &proto.NotificationsList{Notifications: notifications}, nil
}

func (h *NotificationHandler) SaveNotificaion(ctx context.Context, userID int64, author, subject, content string) error {
	err := h.repo.SaveNotification(ctx, userID, author, subject, content)
	if err != nil {
		log.Printf("notif. handler: save: repo error: %v", err)
		return err
	}
	return nil
}

func (h *NotificationHandler) GetEmailById(ctx context.Context, userID int64) (string, error) {
	userConn, err := NewUserConnection()
	if err != nil {
		log.Printf("notif. handler: getemail: userConn: %v", err)
		return "", err
	}
	defer userConn.Close()

	userClient := proto.NewUserServiceClient(userConn)

	resp, err := userClient.GetEmailById(ctx, &proto.IdRequest{Id: userID})
	if err != nil {
		log.Printf("notif. handler: gemail: userClient error: %v", err)
		return "", err
	}

	return resp.Email, nil
}
