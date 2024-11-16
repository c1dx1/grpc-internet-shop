package repository

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"internet-shop/shared/models"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) SaveNotification(ctx context.Context, userID int64, author, subject, content string) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "INSERT INTO notifications(user_id, author, subject, content) VALUES ($1, $2, $3, $4)",
		userID, author, subject, content)
	if err != nil {
		return err
	}

	return nil
}

func (r *NotificationRepository) GetNotifications(ctx context.Context, userID int64) ([]models.Notification, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var notifications []models.Notification
	rows, err := conn.Query(ctx, "SELECT * FROM notifications WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var n models.Notification

		err = rows.Scan(&n.ID, &n.UserID, &n.Author, &n.Subject, &n.Content, &n.CreatedAt)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, n)
	}

	return notifications, nil
}
