package repository

import (
	"gorm.io/gorm"

	"github.com/jishengdan/backend-go/internal/model"
)

// FundRepository 基金数据访问
type FundRepository struct {
	db *gorm.DB
}

// NewFundRepository 创建 FundRepository 实例
func NewFundRepository(db *gorm.DB) *FundRepository {
	return &FundRepository{db: db}
}

// FindByCode 根据基金代码查找基金
func (r *FundRepository) FindByCode(code string) (*model.Fund, error) {
	var fund model.Fund
	err := r.db.Where("fund_code = ?", code).First(&fund).Error
	if err != nil {
		return nil, err
	}
	return &fund, nil
}

// Upsert 插入或更新基金基础信息（按 fund_code 唯一键）
func (r *FundRepository) Upsert(fund *model.Fund) error {
	return r.db.
		Where(model.Fund{FundCode: fund.FundCode}).
		Assign(model.Fund{
			FundName:  fund.FundName,
			FundType:  fund.FundType,
			Manager:   fund.Manager,
			Custodian: fund.Custodian,
			Status:    1,
		}).
		FirstOrCreate(fund).Error
}

// List 基金列表查询
func (r *FundRepository) List(query interface{}) ([]model.Fund, error) {
	var funds []model.Fund
	err := r.db.Where("status = 1").Find(&funds).Error
	return funds, err
}
