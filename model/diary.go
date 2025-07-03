package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DiaryEntry struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title,omitempty"`
	Content   string             `json:"content" bson:"content,omitempty"`
	Emotion   string             `json:"emotion,omitempty" bson:"emotion,omitempty"`     // Contoh: "Joy", "Sadness", "Anger"
	Sentiment string             `json:"sentiment,omitempty" bson:"sentiment,omitempty"` // Contoh: "Positive", "Negative", "Neutral"
	CreatedAt time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
}