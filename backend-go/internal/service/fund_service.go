package service

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/shopspring/decimal"

	"github.com/jishengdan/backend-go/internal/dto"
	"github.com/jishengdan/backend-go/internal/model"
	"github.com/jishengdan/backend-go/internal/repository"
)

// FundService 基金业务逻辑
type FundService struct {
	fundRepo    *repository.FundRepository
	navRepo     *repository.NavRepository
	eggRepo     *repository.EggRepository
	market      *MarketService
}

// NewFundService 创建 FundService 实例
func NewFundService(fundRepo *repository.FundRepository, navRepo *repository.NavRepository, eggRepo *repository.EggRepository, market *MarketService) *FundService {
	return &FundService{
		fundRepo: fundRepo,
		navRepo:  navRepo,
		eggRepo:  eggRepo,
		market:   market,
	}
}

// EnsureFundPublic 对外暴露的自动播种入口（供其他 controller 调用）
func (s *FundService) EnsureFundPublic(fundCode string) (*model.Fund, error) {
	return s.ensureFund(fundCode)
}

// ensureFund 确保基金在 funds 表里存在，不存在则从天天基金 API 自动播种
func (s *FundService) ensureFund(fundCode string) (*model.Fund, error) {
	fund, err := s.fundRepo.FindByCode(fundCode)
	if err == nil {
		return fund, nil
	}

	// 不存在，尝试从外部 API 获取基本信息
	log.Printf("[FundService] 基金 %s 不在数据库，尝试自动播种", fundCode)
	info, fetchErr := s.market.FetchFundInfo(fundCode)
	if fetchErr != nil {
		return nil, fmt.Errorf("基金 %s 不存在且无法获取信息: %w", fundCode, fetchErr)
	}

	fundType := info.FundType
	newFund := &model.Fund{
		FundCode: info.Code,
		FundName: info.Name,
		FundType: &fundType,
		Status:   1,
	}
	if upsertErr := s.fundRepo.Upsert(newFund); upsertErr != nil {
		log.Printf("[FundService] 播种基金 %s 失败: %v", fundCode, upsertErr)
		// 即使写库失败，也返回内存里的数据，不阻塞查询
		return newFund, nil
	}

	log.Printf("[FundService] 基金 %s (%s) 播种成功", fundCode, info.Name)
	return newFund, nil
}

// GetFundDetail 获取基金详情
func (s *FundService) GetFundDetail(fundCode string) (*dto.FundDetailResponse, error) {
	// 获取基金基础信息（不存在则自动播种）
	fund, err := s.ensureFund(fundCode)
	if err != nil {
		return nil, err
	}

	// 获取今日指标
	today := time.Now()
	metric, err := s.eggRepo.GetDailyMetricsByCodeAndDate(fundCode, today)
	if err != nil {
		// 如果今天没数据，用最近一条
		metrics, _ := s.eggRepo.GetRecentMetrics(fundCode, 1)
		if len(metrics) > 0 {
			metric = &metrics[0]
		}
	}

	resp := &dto.FundDetailResponse{
		FundCode:    fund.FundCode,
		FundName:    fund.FundName,
		UnitNav:     0,
		DailyReturn: 0,
		NavDate:     today.Format("2006-01-02"),
		FundType:    fund.FundType,
		Manager:     fund.Manager,
		Custodian:   fund.Custodian,
		Benchmark:   fund.Benchmark,
	}

	if fund.InceptionDate != nil {
		dateStr := fund.InceptionDate.Format("2006-01-02")
		resp.InceptionDate = &dateStr
	}

	if metric != nil {
		resp.UnitNav = metric.UnitNav.InexactFloat64()
		resp.DailyReturn = metric.DailyReturn.InexactFloat64()
		resp.ConsecutiveUp = metric.ConsecutiveUp
		resp.HasNegNews = metric.HasNegNews
		resp.HasPosPolicy = metric.HasPosPolicy
		resp.SentimentCool = metric.SentimentCool
		if metric.WeekReturn.IsPositive() || metric.WeekReturn.IsNegative() {
			resp.WeekReturn = metric.WeekReturn.InexactFloat64()
		}
		if metric.AccNav != nil {
			v := metric.AccNav.InexactFloat64()
			resp.AccNav = &v
		}

		// 计算多周期收益率
		resp.MonthReturn = s.calcPeriodReturn(fund.ID, 22)
		resp.QuarterReturn = s.calcPeriodReturn(fund.ID, 66)
		resp.YearReturn = s.calcPeriodReturn(fund.ID, 252)

		// 计算最大回撤
		resp.MaxDrawdown = s.calcMaxDrawdown(fund.ID, 252)

		// 计算连续下跌天数
		resp.ConsecutiveDown = s.calcConsecutiveDown(fundCode)
	} else {
		// metric 为空时，尝试从实时估值接口获取
		quotes, _ := s.market.FetchFundQuotes([]string{fundCode})
		if len(quotes) > 0 {
			resp.UnitNav = quotes[0].LastNav
			resp.DailyReturn = quotes[0].EstReturn
		}
	}

	return resp, nil
}

// GetNavHistory 获取基金净值历史
func (s *FundService) GetNavHistory(fundCode string, days int) (*dto.NavHistoryResponse, error) {
	fund, err := s.ensureFund(fundCode)
	if err != nil {
		return nil, err
	}

	navs, _ := s.navRepo.FindByFundCodeOrderDate(fundCode, days)

	points := make([]dto.NavPoint, 0, len(navs))
	for _, n := range navs {
		p := dto.NavPoint{
			Date:    n.NavDate.Format("2006-01-02"),
			UnitNav: n.UnitNav.InexactFloat64(),
		}
		if n.AccNav != nil {
			p.AccNav = n.AccNav.InexactFloat64()
		}
		if n.DailyReturn != nil {
			p.DailyReturn = n.DailyReturn.InexactFloat64()
		}
		points = append(points, p)
	}

	// 数据库没有历史数据时，实时从天天基金拉取
	if len(points) < 2 {
		log.Printf("[FundService] %s 数据库无净值历史，实时拉取", fundCode)
		livePoints, fetchErr := s.market.FetchFundNavHistory(fundCode, days)
		if fetchErr != nil {
			log.Printf("[FundService] 实时拉取净值历史失败: %v", fetchErr)
		} else {
			points = make([]dto.NavPoint, 0, len(livePoints))
			for _, p := range livePoints {
				points = append(points, dto.NavPoint{
					Date:        p.Date,
					UnitNav:     p.UnitNav,
					AccNav:      p.AccNav,
					DailyReturn: p.DailyReturn,
				})
			}
		}
	}

	return &dto.NavHistoryResponse{
		FundCode: fund.FundCode,
		FundName: fund.FundName,
		Points:   points,
	}, nil
}

// AnalyzeFund 对单只基金执行三铁律分析
func (s *FundService) AnalyzeFund(fundCode string) (*dto.FundAnalysisResponse, error) {
	fund, err := s.ensureFund(fundCode)
	if err != nil {
		return nil, err
	}

	today := time.Now()

	// 获取今日指标（先查数据库）
	metric, err := s.eggRepo.GetDailyMetricsByCodeAndDate(fundCode, today)
	if err != nil {
		metrics, _ := s.eggRepo.GetRecentMetrics(fundCode, 1)
		if len(metrics) > 0 {
			metric = &metrics[0]
		}
	}

	// 没有数据库指标，用实时估值 + 历史净值构造一个临时 metric
	if metric == nil {
		return s.buildLiveAnalysis(fund, fundCode), nil
	}

	// 获取相关新闻
	newsList, _ := s.eggRepo.GetFilteredNewsByFunds([]string{fundCode}, today)

	// 执行三铁律
	analysis := s.evaluateFundTrend(*metric, newsList)

	return &dto.FundAnalysisResponse{
		FundCode:    fund.FundCode,
		FundName:    fund.FundName,
		Analysis:    analysis,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// buildLiveAnalysis 无数据库指标时，用实时估值做简单分析
func (s *FundService) buildLiveAnalysis(fund *model.Fund, fundCode string) *dto.FundAnalysisResponse {
	// 拉实时估值
	quotes, _ := s.market.FetchFundQuotes([]string{fundCode})
	dailyReturn := 0.0
	if len(quotes) > 0 {
		dailyReturn = quotes[0].EstReturn
	}

	// 拉近20天历史净值判断趋势
	history, _ := s.market.FetchFundNavHistory(fundCode, 20)
	trendLabel, trendColor := s.judgeTrendFromPoints(history)

	// 用实时涨跌幅触发铁律一/三的简化判断
	var signalType, signalText, signalEmoji, reason string
	var confidence float64
	var suggestions []string

	switch {
	case dailyReturn <= -3.0:
		signalType = "buy"
		signalText = "建议关注"
		signalEmoji = "🌧️🐔"
		confidence = 0.65
		reason = fmt.Sprintf("今日估算跌幅%.2f%%，跌幅较大，可关注低吸机会（基于实时估值，非正式数据）", dailyReturn)
		suggestions = []string{
			"今日跌幅较大，可小额试探性买入",
			"建议分批操作，不要一次性重仓",
			"注意设置止损线控制风险",
		}
	case dailyReturn >= 3.0:
		signalType = "hold"
		signalText = "继续持有"
		signalEmoji = "🐔🎵"
		confidence = 0.60
		reason = fmt.Sprintf("今日估算涨幅%.2f%%，行情较好，持有为主（基于实时估值，非正式数据）", dailyReturn)
		suggestions = []string{
			"今日表现强势，持有观望",
			"可设置回撤止盈线",
			"避免追高买入",
		}
	default:
		signalType = "observe"
		signalText = "继续观察"
		signalEmoji = "👁️"
		confidence = 0.45
		reason = fmt.Sprintf("今日估算涨跌幅%.2f%%，波动平稳（基于实时估值，仅供参考）", dailyReturn)
		suggestions = []string{
			"当前无明确买卖信号",
			"建议维持现有仓位，继续观察",
			"关注后续净值变化",
		}
	}

	return &dto.FundAnalysisResponse{
		FundCode: fund.FundCode,
		FundName: fund.FundName,
		Analysis: dto.FundAnalysis{
			FundCode:    fund.FundCode,
			FundName:    fund.FundName,
			SignalType:  signalType,
			SignalText:  signalText,
			SignalEmoji: signalEmoji,
			Confidence:  confidence,
			Reason:      reason,
			TrendLabel:  trendLabel,
			TrendColor:  trendColor,
			Suggestions: suggestions,
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
}

// evaluateFundTrend 评估基金趋势并生成买卖建议
func (s *FundService) evaluateFundTrend(metric model.FundDailyMetrics, newsList []model.AiNews) dto.FundAnalysis {
	dailyReturn := metric.DailyReturn.InexactFloat64()
	weekReturn := metric.WeekReturn.InexactFloat64()
	consecutiveUp := metric.ConsecutiveUp

	// 获取更多最近指标用于趋势判断
	recentMetrics, _ := s.eggRepo.GetRecentMetrics(metric.FundCode, 30)
	trendLabel, trendColor := s.judgeTrend(recentMetrics)

	suggestions := make([]string, 0)

	// 铁律三优先检查：连续上涨+舆情降温 → 止盈
	if consecutiveUp >= 5 && weekReturn >= 8.0 && metric.SentimentCool {
		suggestions = append(suggestions,
			"基金连续上涨且舆情降温，建议部分止盈锁定收益",
			"可保留3-5成底仓，其余分批止盈",
			"关注后续是否出现连续下跌信号",
		)
		return dto.FundAnalysis{
			FundCode:    metric.FundCode,
			FundName:    metric.FundName,
			SignalType:  "sell",
			SignalText:  "建议卖出",
			SignalEmoji: "🥚💰",
			Confidence:  0.90,
			Reason:      fmt.Sprintf("%s连涨%d天累计%.2f%%且舆情降温，触发止盈信号", metric.FundName, consecutiveUp, weekReturn),
			TrendLabel:  trendLabel,
			TrendColor:  trendColor,
			Suggestions: suggestions,
		}
	}

	// 铁律一检查：大跌+无负面 → 建议买入
	if dailyReturn <= -3.0 && !metric.HasNegNews {
		suggestions = append(suggestions,
			"基金单日跌幅较大但无负面舆情，可能是错杀机会",
			"建议分批小额定投，降低持仓成本",
			"设置3-5%的止损线，控制下行风险",
		)
		return dto.FundAnalysis{
			FundCode:    metric.FundCode,
			FundName:    metric.FundName,
			SignalType:  "buy",
			SignalText:  "建议买入",
			SignalEmoji: "🌧️🐔",
			Confidence:  0.85,
			Reason:      fmt.Sprintf("%s日跌%.2f%%且无负面舆情，是低成本建仓良机", metric.FundName, dailyReturn),
			TrendLabel:  trendLabel,
			TrendColor:  trendColor,
			Suggestions: suggestions,
		}
	}

	// 铁律二检查：政策利好+大涨 → 建议持有
	if dailyReturn >= 4.0 && metric.HasPosPolicy {
		suggestions = append(suggestions,
			"基金受政策利好推动大涨，行情向好",
			"建议继续持有，让利润奔跑",
			"可设置回撤止盈线（如回撤5%即止盈部分仓位）",
		)
		return dto.FundAnalysis{
			FundCode:    metric.FundCode,
			FundName:    metric.FundName,
			SignalType:  "hold",
			SignalText:  "建议持有",
			SignalEmoji: "🐔🎵",
			Confidence:  0.80,
			Reason:      fmt.Sprintf("%s日涨%.2f%%且关联政策利好，行情向好宜持有", metric.FundName, dailyReturn),
			TrendLabel:  trendLabel,
			TrendColor:  trendColor,
			Suggestions: suggestions,
		}
	}

	// 无铁律触发，基于趋势给出观察建议
	switch {
	case weekReturn > 5.0:
		suggestions = append(suggestions,
			"近一周表现强势，但未触发铁律信号",
			"建议继续持有观望，不宜追高",
			"注意设置止盈点，防范回调风险",
		)
	case weekReturn < -5.0:
		suggestions = append(suggestions,
			"近一周跌幅较大，但未满足定投触发条件",
			"建议观望，等待企稳信号",
			"如有持仓可考虑适度减仓控制风险",
		)
	default:
		suggestions = append(suggestions,
			"基金波动在正常范围内，无明确买卖信号",
			"建议维持现有仓位，继续观察",
			"关注后续市场走势和基金表现",
		)
	}

	confidence := 0.5
	if len(recentMetrics) > 20 {
		confidence = 0.6
	}

	return dto.FundAnalysis{
		FundCode:    metric.FundCode,
		FundName:    metric.FundName,
		SignalType:  "observe",
		SignalText:  "继续观察",
		SignalEmoji: "👁️",
		Confidence:  confidence,
		Reason:      fmt.Sprintf("当前未触发三铁律，%s趋势：%s，建议继续观察", metric.FundName, trendLabel),
		TrendLabel:  trendLabel,
		TrendColor:  trendColor,
		Suggestions: suggestions,
	}
}

// judgeTrendFromPoints 从实时拉取的历史净值判断趋势
func (s *FundService) judgeTrendFromPoints(points []FundNavPoint) (string, string) {
	if len(points) < 5 {
		return "数据不足", "#787878"
	}
	first := points[0].UnitNav
	last := points[len(points)-1].UnitNav
	mid := points[len(points)/2].UnitNav
	if first == 0 {
		return "数据异常", "#787878"
	}
	shortPct := (last - mid) / mid * 100
	totalPct := (last - first) / first * 100
	if shortPct > 3.0 && totalPct > 5.0 {
		return "上升期", "#ff4d4f"
	}
	if shortPct < -3.0 && totalPct < -5.0 {
		return "下跌期", "#00d68f"
	}
	if shortPct > 1.0 || totalPct > 2.0 {
		return "震荡偏涨", "#f7ba1e"
	}
	if shortPct < -1.0 || totalPct < -2.0 {
		return "震荡偏跌", "#f7ba1e"
	}
	return "平稳期", "#787878"
}

// judgeTrend 根据近期净值数据判断趋势
func (s *FundService) judgeTrend(metrics []model.FundDailyMetrics) (string, string) {	if len(metrics) < 5 {
		return "数据不足", "#787878"
	}

	// 计算短期（5日）和中期（20日）涨幅
	shortTerm := s.calcMetricsReturn(metrics, 5)
	midTerm := s.calcMetricsReturn(metrics, 20)

	if shortTerm > 3.0 && midTerm > 5.0 {
		return "上升期", "#ff4d4f"
	}
	if shortTerm < -3.0 && midTerm < -5.0 {
		return "下跌期", "#00d68f"
	}
	if shortTerm > 1.0 || midTerm > 2.0 {
		return "震荡偏涨", "#f7ba1e"
	}
	if shortTerm < -1.0 || midTerm < -2.0 {
		return "震荡偏跌", "#f7ba1e"
	}
	return "平稳期", "#787878"
}

// calcMetricsReturn 计算近期涨幅
func (s *FundService) calcMetricsReturn(metrics []model.FundDailyMetrics, days int) float64 {
	if len(metrics) < days {
		return 0
	}
	start := metrics[len(metrics)-days]
	end := metrics[len(metrics)-1]
	if start.UnitNav.IsZero() {
		return 0
	}
	return (end.UnitNav.Sub(start.UnitNav).Div(start.UnitNav)).Mul(decimal.NewFromInt(100)).InexactFloat64()
}

// calcPeriodReturn 计算指定周期收益率（从 fund_nav_daily 表）
func (s *FundService) calcPeriodReturn(fundID string, days int) *float64 {
	navs, err := s.navRepo.FindByFundIDOrderDate(fundID, days+5)
	if err != nil || len(navs) < 2 {
		return nil
	}

	// 取最新的净值和 N 天前的净值
	latest := navs[len(navs)-1].UnitNav
	// 找到 days 天前的数据点
	targetIdx := len(navs) - 1 - days
	if targetIdx < 0 {
		targetIdx = 0
	}
	oldest := navs[targetIdx].UnitNav
	if oldest.IsZero() {
		return nil
	}
	ret := latest.Sub(oldest).Div(oldest).Mul(decimal.NewFromInt(100)).InexactFloat64()
	return &ret
}

// calcMaxDrawdown 计算最大回撤
func (s *FundService) calcMaxDrawdown(fundID string, days int) *float64 {
	navs, err := s.navRepo.FindByFundIDOrderDate(fundID, days)
	if err != nil || len(navs) < 2 {
		return nil
	}

	var maxDrawdown float64 = 0
	var peak float64 = navs[0].UnitNav.InexactFloat64()

	for _, n := range navs {
		nav := n.UnitNav.InexactFloat64()
		if nav > peak {
			peak = nav
		}
		drawdown := (peak - nav) / peak * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	maxDrawdown = math.Round(maxDrawdown*100) / 100
	return &maxDrawdown
}

// calcConsecutiveDown 计算连续下跌天数
func (s *FundService) calcConsecutiveDown(fundCode string) int {
	metrics, err := s.eggRepo.GetRecentMetrics(fundCode, 30)
	if err != nil || len(metrics) == 0 {
		return 0
	}

	count := 0
	// 从最新开始往前数
	for i := len(metrics) - 1; i >= 0; i-- {
		if metrics[i].DailyReturn.InexactFloat64() < 0 {
			count++
		} else {
			break
		}
	}
	return count
}

// buildObserveResponse 无数据时的分析
func (s *FundService) buildObserveResponse(fund *model.Fund) *dto.FundAnalysisResponse {
	return &dto.FundAnalysisResponse{
		FundCode: fund.FundCode,
		FundName: fund.FundName,
		Analysis: dto.FundAnalysis{
			FundCode:    fund.FundCode,
			FundName:    fund.FundName,
			SignalType:  "observe",
			SignalText:  "继续观察",
			SignalEmoji: "👁️",
			Confidence:  0,
			Reason:      "暂无该基金数据，无法分析",
			TrendLabel:  "数据不足",
			TrendColor:  "#787878",
			Suggestions: []string{"暂无数据，请稍后再试"},
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
}

// 为排序实现
type byNavDate []model.FundNavDaily

func (a byNavDate) Len() int           { return len(a) }
func (a byNavDate) Less(i, j int) bool { return a[i].NavDate.Before(a[j].NavDate) }
func (a byNavDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Ensure sort.Interface is implemented
var _ sort.Interface = (*byNavDate)(nil)

// logHelper
func (s *FundService) logInfo(format string, args ...interface{}) {
	log.Printf("[FundService] "+format, args...)
}
