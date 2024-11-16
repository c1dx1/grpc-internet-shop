package models

import "time"

type Notification struct {
	ID        int64
	UserID    int64
	Author    string
	Subject   string
	Content   string
	CreatedAt time.Time
}
