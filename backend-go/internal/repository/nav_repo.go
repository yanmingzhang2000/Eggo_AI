package repository

import (
	"gorm.io/gorm"

	"github.com/jishengdan/backend-go/internal/model"
)

// NavRepository 净值数据访问
type NavRepository struct {
	db *gorm.DB
}

// NewNavRepository 创建 NavRepository 实例
func NewNavRepository(db *gorm.DB) *NavRepository {
	return &NavRepository{db: db}
}

// FindByFundCodeOrderDate 根据基金代码查净值按日期升序
func (r *NavRepository) FindByFundCodeOrderDate(fundCode string, limit int) ([]model.FundNavDaily, error) {
	var navs []model.FundNavDaily
	err := r.db.
		Joins("JOIN funds ON funds.id = fund_nav_daily.fund_id").
		Where("funds.fund_code = ?", fundCode).
		Order("nav_date ASC").
		Limit(limit).
		Find(&navs).Error
	return navs, err
}

// FindByFundIDOrderDate 根据基金ID查净值按日期升序
func (r *NavRepository) FindByFundIDOrderDate(fundID string, limit int) ([]model.FundNavDaily, error) {
	var navs []model.FundNavDaily
	err := r.db.
		Where("fund_id = ?", fundID).
		Order("nav_date ASC").
		Limit(limit).
		Find(&navs).Error
	return navs, err
}
