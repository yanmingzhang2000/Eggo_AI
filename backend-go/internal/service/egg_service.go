package service

import (
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"

	"github.com/jishengdan/backend-go/internal/dto"
	"github.com/jishengdan/backend-go/internal/model"
	"github.com/jishengdan/backend-go/internal/repository"
)

// EggService 母鸡状态服务（三铁律决策引擎）
type EggService struct {
	repo *repository.EggRepository
}

// NewEggService 创建 EggService 实例
func NewEggService(repo *repository.EggRepository) *EggService {
	return &EggService{repo: repo}
}

// ────────────────────────────────────────────────────────────────
// 三铁律常量
// ────────────────────────────────────────────────────────────────
const (
	// Rule1 铁律一：落汤鸡 — 日跌超 3% 且无负面新闻 → 建议小额定投
	Rule1DropThreshold = -3.0

	// Rule2 铁律二：母鸡高歌 — 政策利好且日涨超 4% → 老实持有
	Rule2RiseThreshold = 4.0

	// Rule3 铁律三：正在孵蛋 — 连续5日上涨超8%且舆情降温 → 建议部分止盈
	Rule3ConsecutiveDays = 5
	Rule3WeekReturnMin   = 8.0
)

// ────────────────────────────────────────────────────────────────
// 核心决策方法
// ────────────────────────────────────────────────────────────────

// GetEggStatus 获取母鸡状态总接口
func (s *EggService) GetEggStatus(userID string) (*dto.EggStatusResponse, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	log.Printf("[EggService] 开始计算母鸡状态, userID=%s, date=%s", userID, today.Format("2006-01-02"))

	// 1. 获取今日所有基金指标
	metrics, err := s.repo.GetDailyMetricsByDate(today)
	if err != nil {
		return nil, fmt.Errorf("获取基金指标失败: %w", err)
	}
	if len(metrics) == 0 {
		return s.buildEmptyResponse(today), nil
	}

	// 2. 获取今日 AI 过滤新闻
	newsList, err := s.repo.GetFilteredNewsByDate(today)
	if err != nil {
		return nil, fmt.Errorf("获取过滤新闻失败: %w", err)
	}

	// 3. 构建基金指标 DTO
	fundMetrics := make([]dto.FundMetric, 0, len(metrics))
	for _, m := range metrics {
		fundMetrics = append(fundMetrics, dto.FundMetric{
			FundCode:      m.FundCode,
			FundName:      m.FundName,
			UnitNav:       m.UnitNav.InexactFloat64(),
			DailyReturn:   m.DailyReturn.InexactFloat64(),
			ConsecutiveUp: m.ConsecutiveUp,
			WeekReturn:    m.WeekReturn.InexactFloat64(),
			HasNegNews:    m.HasNegNews,
			HasPosPolicy:  m.HasPosPolicy,
			SentimentCool: m.SentimentCool,
		})
	}

	// 4. 构建新闻线索 DTO
	newsClues := s.buildNewsClues(newsList)

	// 5. 执行三铁律决策引擎
	decision := s.evaluateThreeRules(metrics, newsList)

	// 6. 构建母鸡状态
	chickenStatus := s.buildChickenStatus(decision)

	response := &dto.EggStatusResponse{
		ChickenStatus: chickenStatus,
		TodayMetrics:  fundMetrics,
		NewsClues:     newsClues,
		Decision:      decision,
		GeneratedAt:   now.Format(time.RFC3339),
	}

	log.Printf("[EggService] 母鸡状态计算完成, 规则=%s, 操作=%s", decision.TriggeredRule, decision.Action)
	return response, nil
}

// ────────────────────────────────────────────────────────────────
// 三铁律决策引擎
// ────────────────────────────────────────────────────────────────

// evaluateThreeRules 执行三铁律决策
// 优先级：铁律三 > 铁律一 > 铁律二
func (s *EggService) evaluateThreeRules(
	metrics []model.FundDailyMetrics,
	newsList []model.AiNews,
) dto.Decision {
	// 铁律三优先检查（止盈信号最紧急）
	if decision, ok := s.checkRule3(metrics); ok {
		return decision
	}

	// 铁律一（定投机会）
	if decision, ok := s.checkRule1(metrics); ok {
		return decision
	}

	// 铁律二（持有观望）
	if decision, ok := s.checkRule2(metrics, newsList); ok {
		return decision
	}

	// 无触发任何规则
	return dto.Decision{
		TriggeredRule: "none",
		RuleName:      "无触发",
		Action:        "observe",
		ActionEmoji:   "👁️",
		Suggestion:    "今日无铁律触发，继续观察市场动态",
		Confidence:    0.0,
		Reason:        "当前基金指标未满足任何铁律触发条件",
		TargetFunds:   []string{},
	}
}

// checkRule1 铁律一：落汤鸡，速喂饲料
// 条件：日跌超 3% 且对应基金没有负面新闻
func (s *EggService) checkRule1(metrics []model.FundDailyMetrics) (dto.Decision, bool) {
	var targetFunds []string
	var reasons []string

	for _, m := range metrics {
		dailyReturn := m.DailyReturn.InexactFloat64()

		// 条件1：日跌超 3%
		if dailyReturn <= Rule1DropThreshold {
			// 条件2：无负面新闻
			if !m.HasNegNews {
				targetFunds = append(targetFunds, m.FundCode)
				reasons = append(reasons,
					fmt.Sprintf("%s(%s)日跌%.2f%%且无负面舆情",
						m.FundName, m.FundCode, dailyReturn))
			}
		}
	}

	if len(targetFunds) == 0 {
		return dto.Decision{}, false
	}

	return dto.Decision{
		TriggeredRule: "rule1_dca",
		RuleName:      "铁律一：落汤鸡",
		Action:        "dca",
		ActionEmoji:   "🌧️🐔",
		Suggestion:    "触发铁律一：落汤鸡，速喂饲料，建议小额定投",
		Confidence:    0.85,
		Reason:        fmt.Sprintf("检测到 %d 只基金日跌幅超3%%且无负面舆情，是低成本建仓良机: %s", len(targetFunds), joinReasons(reasons)),
		TargetFunds:   targetFunds,
	}, true
}

// checkRule2 铁律二：母鸡高歌，老实持有
// 条件：持仓基金关联行业政策利好且日涨超 4%
func (s *EggService) checkRule2(
	metrics []model.FundDailyMetrics,
	newsList []model.AiNews,
) (dto.Decision, bool) {
	var targetFunds []string
	var reasons []string

	// 检查是否有政策利好新闻
	hasPolicyPositive := false
	for _, news := range newsList {
		if news.Sentiment != nil && *news.Sentiment == 1 {
			for _, tag := range news.Tags {
				if tag == "政策" || tag == "宏观" || tag == "行业" {
					hasPolicyPositive = true
					break
				}
			}
		}
	}

	if !hasPolicyPositive {
		return dto.Decision{}, false
	}

	for _, m := range metrics {
		dailyReturn := m.DailyReturn.InexactFloat64()

		// 条件：日涨超 4% 且有政策利好
		if dailyReturn >= Rule2RiseThreshold && m.HasPosPolicy {
			targetFunds = append(targetFunds, m.FundCode)
			reasons = append(reasons,
				fmt.Sprintf("%s(%s)日涨%.2f%%且关联政策利好",
					m.FundName, m.FundCode, dailyReturn))
		}
	}

	if len(targetFunds) == 0 {
		return dto.Decision{}, false
	}

	return dto.Decision{
		TriggeredRule: "rule2_hold",
		RuleName:      "铁律二：母鸡高歌",
		Action:        "hold",
		ActionEmoji:   "🐔🎵",
		Suggestion:    "触发铁律二：母鸡高歌，老实持有",
		Confidence:    0.80,
		Reason:        fmt.Sprintf("检测到 %d 只基金日涨幅超4%%且关联政策利好，行情向好宜持有: %s", len(targetFunds), joinReasons(reasons)),
		TargetFunds:   targetFunds,
	}, true
}

// checkRule3 铁律三：正在孵蛋，落袋为安
// 条件：连续 5 日上涨超 8% 且出现舆情降温信号
func (s *EggService) checkRule3(metrics []model.FundDailyMetrics) (dto.Decision, bool) {
	var targetFunds []string
	var reasons []string

	for _, m := range metrics {
		weekReturn := m.WeekReturn.InexactFloat64()

		// 条件1：连续5日上涨
		if m.ConsecutiveUp >= Rule3ConsecutiveDays {
			// 条件2：近5日累计涨幅超 8%
			if weekReturn >= Rule3WeekReturnMin {
				// 条件3：舆情降温信号
				if m.SentimentCool {
					targetFunds = append(targetFunds, m.FundCode)
					reasons = append(reasons,
						fmt.Sprintf("%s(%s)连涨%d天累计%.2f%%且舆情降温",
							m.FundName, m.FundCode, m.ConsecutiveUp, weekReturn))
				}
			}
		}
	}

	if len(targetFunds) == 0 {
		return dto.Decision{}, false
	}

	return dto.Decision{
		TriggeredRule: "rule3_take_profit",
		RuleName:      "铁律三：正在孵蛋",
		Action:        "take_profit",
		ActionEmoji:   "🥚💰",
		Suggestion:    "触发铁律三：正在孵蛋，落袋为安，建议部分止盈",
		Confidence:    0.90,
		Reason:        fmt.Sprintf("检测到 %d 只基金连续5日上涨超8%%且舆情出现降温信号，获利了结时机: %s", len(targetFunds), joinReasons(reasons)),
		TargetFunds:   targetFunds,
	}, true
}

// ────────────────────────────────────────────────────────────────
// 辅助方法
// ────────────────────────────────────────────────────────────────

// buildChickenStatus 根据决策构建母鸡状态
func (s *EggService) buildChickenStatus(decision dto.Decision) dto.ChickenStatus {
	switch decision.Action {
	case "dca":
		return dto.ChickenStatus{
			OverallState: "soaking",
			StateEmoji:   "🌧️🐔",
			StateDesc:    "落汤鸡状态 — 市场下跌，正是低成本喂饲料的好时机",
			AlertLevel:   "info",
			AlertMessage: decision.Suggestion,
		}
	case "hold":
		return dto.ChickenStatus{
			OverallState: "laying",
			StateEmoji:   "🐔🎵",
			StateDesc:    "母鸡高歌状态 — 行情向好，安心持有等下蛋",
			AlertLevel:   "none",
			AlertMessage: decision.Suggestion,
		}
	case "take_profit":
		return dto.ChickenStatus{
			OverallState: "laying",
			StateEmoji:   "🥚💰",
			StateDesc:    "正在孵蛋状态 — 连续上涨收益可观，考虑部分止盈",
			AlertLevel:   "warning",
			AlertMessage: decision.Suggestion,
		}
	default:
		return dto.ChickenStatus{
			OverallState: "resting",
			StateEmoji:   "🐔💤",
			StateDesc:    "母鸡休息中 — 今日无铁律触发，继续观察",
			AlertLevel:   "none",
			AlertMessage: "市场平稳，母鸡在休息",
		}
	}
}

// buildNewsClues 构建新闻线索 DTO
func (s *EggService) buildNewsClues(newsList []model.AiNews) []dto.NewsClue {
	clues := make([]dto.NewsClue, 0, len(newsList))

	for _, news := range newsList {
		sentiment := 0
		if news.Sentiment != nil {
			sentiment = int(*news.Sentiment)
		}

		importance := 3
		if news.Importance != nil {
			importance = int(*news.Importance)
		}

		relevanceReason := ""
		if news.Summary != nil {
			relevanceReason = *news.Summary
		}

		publishedAt := ""
		if news.PublishedAt != nil {
			publishedAt = news.PublishedAt.Format(time.RFC3339)
		}

		clues = append(clues, dto.NewsClue{
			NewsID:          news.ID,
			Title:           news.Title,
			Source:          news.Source,
			Sentiment:       sentiment,
			Importance:      importance,
			RelevanceReason: relevanceReason,
			RelatedFunds:    []string(news.RelatedFunds),
			Tags:            []string(news.Tags),
			PublishedAt:     publishedAt,
		})
	}

	return clues
}

// buildEmptyResponse 构建空数据响应
func (s *EggService) buildEmptyResponse(date time.Time) *dto.EggStatusResponse {
	return &dto.EggStatusResponse{
		ChickenStatus: dto.ChickenStatus{
			OverallState: "resting",
			StateEmoji:   "🐔💤",
			StateDesc:    "母鸡休息中 — 暂无数据",
			AlertLevel:   "none",
			AlertMessage: "今日暂无基金数据",
		},
		TodayMetrics: []dto.FundMetric{},
		NewsClues:    []dto.NewsClue{},
		Decision: dto.Decision{
			TriggeredRule: "none",
			RuleName:      "无数据",
			Action:        "observe",
			ActionEmoji:   "👁️",
			Suggestion:    "暂无数据，请稍后再试",
			Confidence:    0,
			Reason:        "今日基金数据尚未更新",
			TargetFunds:   []string{},
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
}

// joinReasons 合并原因列表
func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}
	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		result += "; " + reasons[i]
	}
	return result
}

// ────────────────────────────────────────────────────────────────
// 以下为扩展辅助方法（可选）
// ────────────────────────────────────────────────────────────────

// evaluateSingleFund 对单只基金执行三铁律（用于细粒度分析）
func (s *EggService) evaluateSingleFund(
	metric model.FundDailyMetrics,
	newsList []model.AiNews,
) dto.Decision {
	// 铁律一检查
	dailyReturn := metric.DailyReturn.InexactFloat64()
	if dailyReturn <= Rule1DropThreshold && !metric.HasNegNews {
		return dto.Decision{
			TriggeredRule: "rule1_dca",
			RuleName:      "铁律一：落汤鸡",
			Action:        "dca",
			ActionEmoji:   "🌧️🐔",
			Suggestion:    "触发铁律一：落汤鸡，速喂饲料，建议小额定投",
			Confidence:    0.85,
			Reason:        fmt.Sprintf("%s日跌%.2f%%且无负面舆情", metric.FundName, dailyReturn),
			TargetFunds:   []string{metric.FundCode},
		}
	}

	// 铁律三检查
	if metric.ConsecutiveUp >= Rule3ConsecutiveDays &&
		metric.WeekReturn.InexactFloat64() >= Rule3WeekReturnMin &&
		metric.SentimentCool {
		return dto.Decision{
			TriggeredRule: "rule3_take_profit",
			RuleName:      "铁律三：正在孵蛋",
			Action:        "take_profit",
			ActionEmoji:   "🥚💰",
			Suggestion:    "触发铁律三：正在孵蛋，落袋为安，建议部分止盈",
			Confidence:    0.90,
			Reason: fmt.Sprintf("%s连涨%d天累计%.2f%%且舆情降温",
				metric.FundName, metric.ConsecutiveUp, metric.WeekReturn.InexactFloat64()),
			TargetFunds: []string{metric.FundCode},
		}
	}

	// 铁律二检查
	if dailyReturn >= Rule2RiseThreshold && metric.HasPosPolicy {
		return dto.Decision{
			TriggeredRule: "rule2_hold",
			RuleName:      "铁律二：母鸡高歌",
			Action:        "hold",
			ActionEmoji:   "🐔🎵",
			Suggestion:    "触发铁律二：母鸡高歌，老实持有",
			Confidence:    0.80,
			Reason:        fmt.Sprintf("%s日涨%.2f%%且关联政策利好", metric.FundName, dailyReturn),
			TargetFunds:   []string{metric.FundCode},
		}
	}

	return dto.Decision{
		TriggeredRule: "none",
		RuleName:      "无触发",
		Action:        "observe",
		ActionEmoji:   "👁️",
		Suggestion:    "今日无铁律触发",
		Confidence:    0,
		Reason:        "当前指标未满足触发条件",
		TargetFunds:   []string{},
	}
}

// decimalToFloat 安全转换 decimal 到 float64
func decimalToFloat(d *decimal.Decimal) float64 {
	if d == nil {
		return 0
	}
	return d.InexactFloat64()
}
