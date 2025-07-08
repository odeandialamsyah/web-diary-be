package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User merepresentasikan satu dokumen pengguna di koleksi 'users' MongoDB
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Username     string             `bson:"username"`
	Email        string             `bson:"email"` 
	Password     string             `bson:"password"`        
	CreatedAt    time.Time          `bson:"created_at"` 
	UpdatedAt    time.Time          `bson:"updated_at,omitempty"`
}

type DiaryEntry struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Title     string             `json:"title" bson:"title,omitempty"`
	Content   string             `json:"content" bson:"content,omitempty"`
	Emotion   string             `json:"emotion,omitempty" bson:"emotion,omitempty"`     // Contoh: "Joy", "Sadness", "Anger"
	Sentiment string             `json:"sentiment,omitempty" bson:"sentiment,omitempty"` // Contoh: "Positive", "Negative", "Neutral"
	CreatedAt time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

