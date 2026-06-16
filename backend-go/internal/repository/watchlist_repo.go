package repository

import (
	"gorm.io/gorm"

	"github.com/jishengdan/backend-go/internal/model"
)

// WatchlistRepository 自选基金数据访问
type WatchlistRepository struct {
	db *gorm.DB
}

// NewWatchlistRepository 创建 WatchlistRepository 实例
func NewWatchlistRepository(db *gorm.DB) *WatchlistRepository {
	return &WatchlistRepository{db: db}
}

// FindByUserID 查询用户所有自选基金（含基金信息）
func (r *WatchlistRepository) FindByUserID(userID string) ([]model.Watchlist, error) {
	var items []model.Watchlist
	err := r.db.
		Preload("Fund").
		Where("user_id = ?", userID).
		Order("sort_order ASC, created_at ASC").
		Find(&items).Error
	return items, err
}

// Exists 检查用户是否已收藏某基金
func (r *WatchlistRepository) Exists(userID, fundID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Watchlist{}).
		Where("user_id = ? AND fund_id = ?", userID, fundID).
		Count(&count).Error
	return count > 0, err
}

// Create 添加自选
func (r *WatchlistRepository) Create(w *model.Watchlist) error {
	return r.db.Create(w).Error
}

// Delete 删除自选（按 user_id + fund_id）
func (r *WatchlistRepository) Delete(userID, fundID string) error {
	return r.db.
		Where("user_id = ? AND fund_id = ?", userID, fundID).
		Delete(&model.Watchlist{}).Error
}
