package document

import (
	"time"

	"github.com/google/uuid"
)

const DefaultDisplayName = "Untitled"

type Document struct {
	DocumentID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	DisplayName string    `gorm:"type:text;not null"`
	Content     []byte    `gorm:"type:bytea;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Document) TableName() string {
	return "documents"
}

func Models() []interface{} {
	return []interface{}{Document{}}
}
