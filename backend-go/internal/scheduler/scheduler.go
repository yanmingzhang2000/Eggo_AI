package scheduler

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/jishengdan/backend-go/internal/model"
	"github.com/jishengdan/backend-go/pkg/tushare"
)

type Scheduler struct {
	db      *gorm.DB
	ts      *tushare.Client
	cron    *cron.Cron
}

func New(db *gorm.DB, ts *tushare.Client) *Scheduler {
	return &Scheduler{
		db:   db,
		ts:   ts,
		cron: cron.New(cron.WithLocation(time.FixedZone("CST", 8*3600))),
	}
}

func (s *Scheduler) Start() {
	s.cron.AddFunc("30 15 * * 1-5", s.DailyNavUpdate)
	s.cron.Start()
	log.Println("[Scheduler] 定时任务已启动（每个交易日 15:30 更新净值）")
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("[Scheduler] 定时任务已停止")
}

func (s *Scheduler) DailyNavUpdate() {
	log.Println("[Scheduler] 开始每日净值更新")

	var funds []model.Fund
	if err := s.db.Where("status = 1").Find(&funds).Error; err != nil {
		log.Printf("[Scheduler] 查询基金列表失败: %v", err)
		return
	}

	now := time.Now()
	today := now.Format("20060102")

	for _, fund := range funds {
		func(f model.Fund) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Scheduler] %s 处理异常: %v", f.FundCode, r)
				}
			}()

			data, err := s.ts.FundDaily(f.FundCode+".OF", "", today, 1)
			if err != nil {
				log.Printf("[Scheduler] %s 获取净值失败: %v", f.FundCode, err)
				return
			}
			if len(data) == 0 {
				log.Printf("[Scheduler] %s 无净值数据", f.FundCode)
				return
			}

			d := data[0]
			navDate, _ := tushare.ParseDate(d.TradeDate)
			if navDate.IsZero() {
				return
			}

			unitNav := decimal.NewFromFloat(d.UnitNav)
			var accNav *decimal.Decimal
			if d.AccumNav > 0 {
				v := decimal.NewFromFloat(d.AccumNav)
				accNav = &v
			}
			var dailyReturn *decimal.Decimal
			if d.DailyReturn != 0 {
				v := decimal.NewFromFloat(d.DailyReturn)
				dailyReturn = &v
			}

			navRecord := model.FundNavDaily{
				FundID:      f.ID,
				NavDate:     navDate,
				UnitNav:     unitNav,
				AccNav:      accNav,
				DailyReturn: dailyReturn,
			}
			if err := s.db.Where("fund_id = ? AND nav_date = ?", f.ID, navDate).
				Assign(navRecord).
				FirstOrCreate(&model.FundNavDaily{}).Error; err != nil {
				log.Printf("[Scheduler] %s 写入净值失败: %v", f.FundCode, err)
				return
			}

			s.updateDailyMetrics(f, navDate, unitNav, dailyReturn)
			log.Printf("[Scheduler] %s (%s) 净值更新完成: %v", f.FundCode, f.FundName, unitNav)
		}(fund)
	}

	log.Printf("[Scheduler] 每日净值更新完成, 共处理 %d 只基金", len(funds))
}

func (s *Scheduler) updateDailyMetrics(f model.Fund, navDate time.Time, unitNav decimal.Decimal, dailyReturn *decimal.Decimal) {
	zero := decimal.NewFromInt(0)
	dr := zero
	if dailyReturn != nil {
		dr = *dailyReturn
	}

	metrics := model.FundDailyMetrics{
		FundID:      f.ID,
		FundCode:    f.FundCode,
		FundName:    f.FundName,
		MetricsDate: navDate,
		UnitNav:     unitNav,
		DailyReturn: dr,
	}

	var prevMetrics []model.FundDailyMetrics
	s.db.Where("fund_code = ? AND metrics_date < ?", f.FundCode, navDate).
		Order("metrics_date DESC").Limit(5).Find(&prevMetrics)

	if len(prevMetrics) > 0 {
		weekStart := navDate.AddDate(0, 0, -5)
		var weekSum decimal.Decimal
		for _, pm := range prevMetrics {
			if pm.MetricsDate.After(weekStart) || pm.MetricsDate.Equal(weekStart) {
				weekSum = weekSum.Add(pm.DailyReturn)
			}
		}
		metrics.WeekReturn = weekSum.Add(dr)

		if dr.IsPositive() {
			metrics.ConsecutiveUp = 1
			if prevMetrics[0].DailyReturn.IsPositive() {
				metrics.ConsecutiveUp = prevMetrics[0].ConsecutiveUp + 1
			}
		}
	}

	if err := s.db.Where("fund_code = ? AND metrics_date = ?", f.FundCode, navDate).
		Assign(metrics).
		FirstOrCreate(&model.FundDailyMetrics{}).Error; err != nil {
		log.Printf("[Scheduler] %s 写入指标失败: %v", f.FundCode, err)
	}
}
