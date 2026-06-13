package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/jishengdan/backend-go/internal/model"
)

// EggRepository 母鸡状态数据访问层
type EggRepository struct {
	db *gorm.DB
}

// NewEggRepository 创建 EggRepository 实例
func NewEggRepository(db *gorm.DB) *EggRepository {
	return &EggRepository{db: db}
}

// GetDailyMetricsByDate 获取指定日期的所有基金每日指标
func (r *EggRepository) GetDailyMetricsByDate(date time.Time) ([]model.FundDailyMetrics, error) {
	var metrics []model.FundDailyMetrics
	dateStr := date.Format("2006-01-02")

	err := r.db.
		Preload("Fund").
		Where("metrics_date = ?", dateStr).
		Order("fund_code ASC").
		Find(&metrics).Error

	return metrics, err
}

// GetDailyMetricsByCodeAndDate 获取指定基金指定日期的指标
func (r *EggRepository) GetDailyMetricsByCodeAndDate(fundCode string, date time.Time) (*model.FundDailyMetrics, error) {
	var metric model.FundDailyMetrics
	dateStr := date.Format("2006-01-02")

	err := r.db.
		Preload("Fund").
		Where("fund_code = ? AND metrics_date = ?", fundCode, dateStr).
		First(&metric).Error

	if err != nil {
		return nil, err
	}
	return &metric, nil
}

// GetRecentMetrics 获取指定基金近 N 天的指标（按日期升序）
func (r *EggRepository) GetRecentMetrics(fundCode string, days int) ([]model.FundDailyMetrics, error) {
	var metrics []model.FundDailyMetrics

	err := r.db.
		Where("fund_code = ?", fundCode).
		Order("metrics_date ASC").
		Limit(days).
		Find(&metrics).Error

	return metrics, err
}

// GetFilteredNewsByDate 获取指定日期的 AI 过滤新闻（按重要性降序）
func (r *EggRepository) GetFilteredNewsByDate(date time.Time) ([]model.AiNews, error) {
	var news []model.AiNews
	dateStr := date.Format("2006-01-02")

	err := r.db.
		Where("DATE(processed_at) = ?", dateStr).
		Order("importance DESC, published_at DESC").
		Find(&news).Error

	return news, err
}

// GetFilteredNewsByFunds 获取与指定基金相关的新闻
func (r *EggRepository) GetFilteredNewsByFunds(fundCodes []string, date time.Time) ([]model.AiNews, error) {
	var news []model.AiNews
	dateStr := date.Format("2006-01-02")

	// MySQL JSON 数组包含查询
	query := r.db.Where("DATE(processed_at) = ?", dateStr)

	// 构建 JSON_CONTAINS 条件
	conditions := make([]string, 0, len(fundCodes))
	args := make([]interface{}, 0, len(fundCodes)*2)
	for _, code := range fundCodes {
		conditions = append(conditions, "JSON_CONTAINS(related_funds, ?)")
		args = append(args, `"`+code+`"`)
	}

	if len(conditions) > 0 {
		query = query.Where(conditions[0], args[0])
		for i := 1; i < len(conditions); i++ {
			query = query.Or(conditions[i], args[i])
		}
	}

	err := query.Order("importance DESC, published_at DESC").Find(&news).Error
	return news, err
}

// GetNegativeNewsCount 获取指定基金指定日期的负面新闻数量
func (r *EggRepository) GetNegativeNewsCount(fundCode string, date time.Time) (int64, error) {
	var count int64
	dateStr := date.Format("2006-01-02")

	err := r.db.Model(&model.AiNews{}).
		Where("DATE(processed_at) = ? AND sentiment = -1 AND JSON_CONTAINS(related_funds, ?)",
			dateStr, `"`+fundCode+`"`).
		Count(&count).Error

	return count, err
}

// GetPositivePolicyNews 获取政策利好新闻
func (r *EggRepository) GetPositivePolicyNews(date time.Time) ([]model.AiNews, error) {
	var news []model.AiNews
	dateStr := date.Format("2006-01-02")

	// MySQL JSON 数组查询
	err := r.db.
		Where("DATE(processed_at) = ? AND sentiment = 1 AND (JSON_CONTAINS(tags, '\"政策\"') OR JSON_CONTAINS(tags, '\"宏观\"') OR JSON_CONTAINS(tags, '\"行业\"'))", dateStr).
		Order("importance DESC").
		Find(&news).Error

	return news, err
}

// HasSentimentCoolingSignal 检查是否有舆情降温信号
// 判断逻辑：近3天新闻重要性均值 < 近7天均值的 0.7 倍
func (r *EggRepository) HasSentimentCoolingSignal(fundCode string, date time.Time) (bool, error) {
	dateStr := date.Format("2006-01-02")

	var avgRecent3, avgRecent7 float64

	// 近3天平均重要性
	err := r.db.Model(&model.AiNews{}).
		Where("DATE(processed_at) BETWEEN ? AND ? AND JSON_CONTAINS(related_funds, ?)",
			date.AddDate(0, 0, -3).Format("2006-01-02"), dateStr, `"`+fundCode+`"`).
		Select("COALESCE(AVG(importance), 0)").
		Scan(&avgRecent3).Error
	if err != nil {
		return false, err
	}

	// 近7天平均重要性
	err = r.db.Model(&model.AiNews{}).
		Where("DATE(processed_at) BETWEEN ? AND ? AND JSON_CONTAINS(related_funds, ?)",
			date.AddDate(0, 0, -7).Format("2006-01-02"), dateStr, `"`+fundCode+`"`).
		Select("COALESCE(AVG(importance), 0)").
		Scan(&avgRecent7).Error
	if err != nil {
		return false, err
	}

	// 降温信号：近3天均值 < 近7天均值 * 0.7
	if avgRecent7 > 0 && avgRecent3 < avgRecent7*0.7 {
		return true, nil
	}

	return false, nil
}
