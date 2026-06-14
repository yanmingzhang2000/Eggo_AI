package service

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MarketService 大盘行情服务
type MarketService struct {
	client  *http.Client
	cache   map[string]cacheEntry
	cacheMu sync.RWMutex
}

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

func NewMarketService() *MarketService {
	return &MarketService{
		client: &http.Client{Timeout: 20 * time.Second},
		cache:  make(map[string]cacheEntry),
	}
}

// IndexData 指数行情数据
type IndexData struct {
	Name          string  `json:"name"`
	Code          string  `json:"code"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
	Volume        string  `json:"volume"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Open          float64 `json:"open"`
	PreClose      float64 `json:"preClose"`
	Amount        string  `json:"amount"`
	UpdateAt      string  `json:"updateAt"`
}

// SectorData 板块数据
type SectorData struct {
	Name          string  `json:"name"`
	Code          string  `json:"code,omitempty"`
	ChangePercent float64 `json:"changePercent"`
	Price         float64 `json:"price"`
	LeadStock     string  `json:"leadStock,omitempty"`
	Reason        string  `json:"reason,omitempty"`
	Amount        string  `json:"amount,omitempty"`
}

// MarketStats 市场统计
type MarketStats struct {
	UpCount      int    `json:"upCount"`
	DownCount    int    `json:"downCount"`
	FlatCount    int    `json:"flatCount"`
	LimitUp      int    `json:"limitUp"`
	LimitDown    int    `json:"limitDown"`
	TotalVolume  string `json:"totalVolume"`
	NorthFlow    string `json:"northFlow"`
	Sentiment    string `json:"sentiment"`
	UpdateAt     string `json:"updateAt"`
}

// getCache 缓存辅助
func (s *MarketService) getCache(key string) interface{} {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	if e, ok := s.cache[key]; ok && time.Now().Before(e.expiresAt) {
		return e.data
	}
	return nil
}

func (s *MarketService) setCache(key string, data interface{}, ttl time.Duration) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cache[key] = cacheEntry{data: data, expiresAt: time.Now().Add(ttl)}
}

// eastmoneyToYahooCode 将东方财富 secid 格式转换为 Yahoo Finance symbol
// 东方财富格式：1.000001（沪市）、0.399001（深市）
// Yahoo 格式：000001.SS（沪）、399001.SZ（深）
func eastmoneyToYahooCode(code string) string {
	parts := strings.SplitN(code, ".", 2)
	if len(parts) != 2 {
		return code
	}
	market, symbol := parts[0], parts[1]
	switch market {
	case "1":
		return symbol + ".SS"
	case "0":
		return symbol + ".SZ"
	default:
		return code
	}
}

// yahooIndexName 返回指数的中文名称
func yahooIndexName(code string) string {
	names := map[string]string{
		"1.000001": "上证指数",
		"0.399001": "深证成指",
		"0.399006": "创业板指",
		"1.000300": "沪深300",
		"1.000905": "中证500",
		"1.000852": "中证1000",
	}
	if n, ok := names[code]; ok {
		return n
	}
	return code
}

// fetchYahooQuote 从 Yahoo Finance 获取单个指数的实时行情
func (s *MarketService) fetchYahooQuote(emCode string) (*IndexData, error) {
	yahooSymbol := eastmoneyToYahooCode(emCode)
	apiURL := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=5d",
		yahooSymbol,
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch yahoo quote %s failed: %w", yahooSymbol, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Symbol           string  `json:"symbol"`
					RegularMarketPrice    float64 `json:"regularMarketPrice"`
					PreviousClose         float64 `json:"chartPreviousClose"`
					RegularMarketVolume   int64   `json:"regularMarketVolume"`
					RegularMarketDayHigh  float64 `json:"regularMarketDayHigh"`
					RegularMarketDayLow   float64 `json:"regularMarketDayLow"`
					RegularMarketOpen     float64 `json:"regularMarketOpen"`
				} `json:"meta"`
			} `json:"result"`
			Error *struct {
				Code string `json:"code"`
			} `json:"error"`
		} `json:"chart"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse yahoo quote failed: %w", err)
	}
	if result.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo error: %s", result.Chart.Error.Code)
	}
	if len(result.Chart.Result) == 0 {
		return nil, fmt.Errorf("yahoo: no result for %s", yahooSymbol)
	}

	meta := result.Chart.Result[0].Meta
	price := meta.RegularMarketPrice
	preClose := meta.PreviousClose
	change := price - preClose
	changePct := 0.0
	if preClose > 0 {
		changePct = (change / preClose) * 100
	}

	idx := &IndexData{
		Name:          yahooIndexName(emCode),
		Code:          emCode,
		Price:         math.Round(price*100) / 100,
		Change:        math.Round(change*100) / 100,
		ChangePercent: math.Round(changePct*100) / 100,
		High:          meta.RegularMarketDayHigh,
		Low:           meta.RegularMarketDayLow,
		Open:          meta.RegularMarketOpen,
		PreClose:      preClose,
		Volume:        formatVolume(float64(meta.RegularMarketVolume)),
		Amount:        "--",
		UpdateAt:      time.Now().Format("2006-01-02 15:04:05"),
	}
	return idx, nil
}

// FetchSinaIndex 获取多个指数的实时行情（已切换至 Yahoo Finance）
// codes 参数保持东方财富 secid 格式，兼容原有调用方
func (s *MarketService) FetchSinaIndex(codes []string) ([]IndexData, error) {
	cacheKey := "yahoo_indices_" + strings.Join(codes, "_")
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]IndexData), nil
	}

	var results []IndexData
	for _, code := range codes {
		idx, err := s.fetchYahooQuote(code)
		if err != nil {
			// 单个失败不中断，跳过
			continue
		}
		results = append(results, *idx)
	}

	if len(results) > 0 {
		s.setCache(cacheKey, results, 30*time.Second)
	}

	return results, nil
}

// FetchEastmoneySectors 获取行业板块行情
// Yahoo Finance 无 A 股板块数据，返回空列表，前端正常渲染空状态
func (s *MarketService) FetchEastmoneySectors() ([]SectorData, error) {
	return []SectorData{}, nil
}

// FetchEastmoneyConcepts 获取概念板块行情
// Yahoo Finance 无 A 股概念数据，返回空列表，前端正常渲染空状态
func (s *MarketService) FetchEastmoneyConcepts() ([]SectorData, error) {
	return []SectorData{}, nil
}

// FetchMarketOverview 获取市场概览
// 从 Yahoo Finance 上证/深证/创业板 ETF 推算市场情绪；涨跌家数在境外无可靠接口，返回 0
func (s *MarketService) FetchMarketOverview() (*MarketStats, error) {
	cacheKey := "market_overview"
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.(*MarketStats), nil
	}

	// 用沪深300涨跌幅判断市场情绪
	idx, err := s.fetchYahooQuote("1.000300")
	stats := &MarketStats{
		NorthFlow: "--",
		UpdateAt:  time.Now().Format("2006-01-02 15:04:05"),
	}

	if err == nil {
		cp := idx.ChangePercent
		switch {
		case cp > 1.5:
			stats.Sentiment = "强势"
		case cp > 0.3:
			stats.Sentiment = "偏多"
		case cp < -1.5:
			stats.Sentiment = "弱势"
		case cp < -0.3:
			stats.Sentiment = "偏空"
		default:
			stats.Sentiment = "震荡"
		}
	} else {
		stats.Sentiment = "数据获取中"
	}

	s.setCache(cacheKey, stats, 60*time.Second)
	return stats, nil
}

// FetchNorthFlow 北向资金（境外无可靠接口，返回 "--"）
func (s *MarketService) FetchNorthFlow() string {
	return "--"
}

func parseSinaFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func formatAmount(val float64) string {
	if val >= 1e12 {
		return fmt.Sprintf("%.0f万亿", val/1e12)
	}
	return fmt.Sprintf("%.0f亿", val/1e8)
}

func formatVolume(val float64) string {
	if val >= 1e8 {
		return fmt.Sprintf("%.0f亿手", val/1e8)
	}
	return fmt.Sprintf("%.0f万手", val/1e4)
}

// KLineData K线数据
type KLineData struct {
	Date  string  `json:"date"`
	Open  float64 `json:"open"`
	Close float64 `json:"close"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
}

// IndexOption 指数选项
type IndexOption struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// GetIndexOptions 返回可选指数列表
func GetIndexOptions() []IndexOption {
	return []IndexOption{
		{Name: "上证指数", Code: "1.000001"},
		{Name: "深证成指", Code: "0.399001"},
		{Name: "创业板指", Code: "0.399006"},
		{Name: "沪深300", Code: "1.000300"},
		{Name: "中证500", Code: "1.000905"},
		{Name: "中证1000", Code: "1.000852"},
	}
}

// FetchIndexHistory 从 Yahoo Finance 获取指数历史K线
// Yahoo Finance 境外可正常访问
func (s *MarketService) FetchIndexHistory(code string, days int) ([]KLineData, error) {
	cacheKey := fmt.Sprintf("index_history_%s_%d", code, days)
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]KLineData), nil
	}

	yahooSymbol := eastmoneyToYahooCode(code)

	// 将交易天数换算为日历天数范围（交易日约为日历日的 70%，多加缓冲）
	calDays := int(float64(days)*1.5) + 30
	rangeStr := fmt.Sprintf("%dd", calDays)

	apiURL := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=%s",
		yahooSymbol, rangeStr,
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch yahoo history failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	// Yahoo Finance v8 chart 响应结构
	var result struct {
		Chart struct {
			Result []struct {
				Timestamps []int64 `json:"timestamp"`
				Indicators struct {
					Quote []struct {
						Open  []float64 `json:"open"`
						Close []float64 `json:"close"`
						High  []float64 `json:"high"`
						Low   []float64 `json:"low"`
					} `json:"quote"`
				} `json:"indicators"`
			} `json:"result"`
			Error *struct {
				Code string `json:"code"`
			} `json:"error"`
		} `json:"chart"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse yahoo history failed: %w", err)
	}
	if result.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo error: %s", result.Chart.Error.Code)
	}
	if len(result.Chart.Result) == 0 {
		return []KLineData{}, nil
	}

	res := result.Chart.Result[0]
	if len(res.Indicators.Quote) == 0 {
		return []KLineData{}, nil
	}

	quotes := res.Indicators.Quote[0]
	timestamps := res.Timestamps

	var klines []KLineData
	for i, ts := range timestamps {
		if i >= len(quotes.Close) || i >= len(quotes.Open) || i >= len(quotes.High) || i >= len(quotes.Low) {
			break
		}
		// Yahoo 有时会在数组中插入 NaN（用 0 表示），跳过无效数据
		if quotes.Close[i] == 0 || math.IsNaN(quotes.Close[i]) {
			continue
		}
		date := time.Unix(ts, 0).UTC().Format("2006-01-02")
		klines = append(klines, KLineData{
			Date:  date,
			Open:  math.Round(quotes.Open[i]*100) / 100,
			Close: math.Round(quotes.Close[i]*100) / 100,
			High:  math.Round(quotes.High[i]*100) / 100,
			Low:   math.Round(quotes.Low[i]*100) / 100,
		})
	}

	// 只保留最后 days 条交易日数据
	if len(klines) > days {
		klines = klines[len(klines)-days:]
	}

	if len(klines) > 0 {
		s.setCache(cacheKey, klines, 5*time.Minute)
	}

	return klines, nil
}
