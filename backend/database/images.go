package database

import (
	"context"
	"gorm.io/gorm"
)

// ImageStore keeps handlers testable without connecting to the configured database.
type ImageStore interface {
	List(context.Context, int, int, string) ([]Image, int64, error)
	Create(context.Context, *Image) error
	Get(context.Context, string) (Image, error)
	MarkReady(context.Context, string) error
	Update(context.Context, string, string, string) (Image, error)
	Delete(context.Context, string) error
}

type PostgresImages struct{ DB *gorm.DB }

func (s PostgresImages) List(ctx context.Context, limit, offset int, search string) ([]Image, int64, error) {
	images := []Image{}
	query := s.DB.WithContext(ctx).Model(&Image{})
	if search != "" {
		query = query.Where("strpos(lower(name), lower(?)) > 0", search)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&images).Error
	return images, total, err
}
func (s PostgresImages) Create(ctx context.Context, image *Image) error {
	return s.DB.WithContext(ctx).Create(image).Error
}
func (s PostgresImages) Get(ctx context.Context, id string) (Image, error) {
	var image Image
	err := s.DB.WithContext(ctx).First(&image, "id = ?", id).Error
	return image, err
}
func (s PostgresImages) MarkReady(ctx context.Context, id string) error {
	result := s.DB.WithContext(ctx).Model(&Image{}).Where("id = ?", id).Update("status", "ready")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (s PostgresImages) Update(ctx context.Context, id, name, alt string) (Image, error) {
	result := s.DB.WithContext(ctx).Model(&Image{}).Where("id = ?", id).Updates(map[string]any{"name": name, "alt_text": alt})
	if result.Error != nil {
		return Image{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Image{}, gorm.ErrRecordNotFound
	}
	return s.Get(ctx, id)
}
func (s PostgresImages) Delete(ctx context.Context, id string) error {
	return s.DB.WithContext(ctx).Delete(&Image{}, "id = ?", id).Error
}
