package document

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateDocument(ctx context.Context, displayName string) (Document, error) {
	name := displayName
	if name == "" {
		name = DefaultDisplayName
	}
	doc := Document{
		DocumentID:  uuid.New(),
		DisplayName: name,
		Content:     []byte{},
	}
	if err := s.db.WithContext(ctx).Create(&doc).Error; err != nil {
		return Document{}, err
	}
	return doc, nil
}

func (s *Store) GetDocument(ctx context.Context, id uuid.UUID) (Document, error) {
	var doc Document
	if err := s.db.WithContext(ctx).First(&doc, "document_id = ?", id).Error; err != nil {
		return Document{}, err
	}
	return doc, nil
}

func (s *Store) ListDocuments(ctx context.Context, query string, limit, offset int) ([]Document, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	qb := s.db.WithContext(ctx).Model(&Document{})
	if query != "" {
		qb = qb.Where("display_name ILIKE ?", "%"+query+"%")
	}
	var docs []Document
	if err := qb.Order("updated_at desc").Limit(limit).Offset(offset).Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func (s *Store) UpdateContent(ctx context.Context, id uuid.UUID, content []byte) error {
	result := s.db.WithContext(ctx).Model(&Document{}).
		Where("document_id = ?", id).
		Updates(map[string]interface{}{
			"content":    content,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Store) UpdateTitle(ctx context.Context, id uuid.UUID, title string) error {
	result := s.db.WithContext(ctx).Model(&Document{}).
		Where("document_id = ?", id).
		Updates(map[string]interface{}{
			"display_name": title,
			"updated_at":   time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Store) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&Document{}, "document_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
