package service

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MarketService 大盘行情服务
type MarketService struct {
	client    *http.Client
	cache     map[string]cacheEntry
	cacheMu   sync.RWMutex
	proxyBase string // Cloudflare Worker 代理地址，为空则直连
	proxyKey  string // 代理 API Key
}

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

func NewMarketService() *MarketService {
	// 从环境变量读取代理配置
	proxyBase := os.Getenv("CF_PROXY_URL")  // e.g. https://eggo-fund-proxy.xxx.workers.dev
	proxyKey := os.Getenv("CF_PROXY_KEY")    // e.g. eggo-proxy-2026

	return &MarketService{
		client:    &http.Client{Timeout: 20 * time.Second},
		cache:     make(map[string]cacheEntry),
		proxyBase: proxyBase,
		proxyKey:  proxyKey,
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

// proxyFetch 通过 Cloudflare Worker 代理发起请求
// 如果 proxyBase 为空则直接请求（用于境内直连）
func (s *MarketService) proxyFetch(rawURL string) (*http.Response, error) {
	if s.proxyBase == "" {
		// 无代理，直连
		req, err := http.NewRequest("GET", rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Referer", "https://fund.eastmoney.com/")
		return s.client.Do(req)
	}

	// 通过代理转发
	proxyURL := fmt.Sprintf("%s/proxy?url=%s&key=%s",
		s.proxyBase,
		url.QueryEscape(rawURL),
		url.QueryEscape(s.proxyKey),
	)

	req, err := http.NewRequest("GET", proxyURL, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(req)
}

// proxyFetchText 通过代理获取文本内容（简化调用）
func (s *MarketService) proxyFetchText(rawURL string) (string, error) {
	resp, err := s.proxyFetch(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
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

// ─── 基金涨跌分布 ──────────────────────────────────────────────────────────────

// FundDistribution 全市场基金涨跌分布统计
type FundDistribution struct {
	RiseCount  int     `json:"riseCount"`  // 上涨只数
	FallCount  int     `json:"fallCount"`  // 下跌只数
	FlatCount  int     `json:"flatCount"`  // 持平只数
	Total      int     `json:"total"`      // 样本总数
	AvgReturn  float64 `json:"avgReturn"`  // 平均涨幅 %
	TopFunds   []FundQuote `json:"topFunds"`   // 涨幅 Top5
	FlopFunds  []FundQuote `json:"flopFunds"`  // 跌幅 Top5
	UpdateAt   string  `json:"updateAt"`
}

// FundQuote 基金实时估值
type FundQuote struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	EstReturn  float64 `json:"estReturn"`  // 今日估算涨跌幅 %
	EstNav     float64 `json:"estNav"`     // 估算净值
	LastNav    float64 `json:"lastNav"`    // 昨日净值
	UpdateTime string  `json:"updateTime"`
}

// FetchFundDistribution 从天天基金获取全市场基金当日涨跌分布
// 拉取当日涨幅榜和跌幅榜各100只，统计分布
func (s *MarketService) FetchFundDistribution() (*FundDistribution, error) {
	cacheKey := "fund_distribution"
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.(*FundDistribution), nil
	}

	today := time.Now().Format("2006-01-02")

	// 拉取涨幅榜 Top100
	riseURL := fmt.Sprintf(
		"https://fund.eastmoney.com/data/rankhandler.aspx?op=ph&dt=kf&ft=all&rs=&gs=0&sc=rzdf&st=desc&sd=%s&ed=%s&qdii=&tabSubtype=,,,,,&pi=1&pn=100&dx=1",
		today, today,
	)
	riseData, err := s.fetchEastmoneyRank(riseURL)
	if err != nil {
		return nil, fmt.Errorf("fetch rise rank failed: %w", err)
	}

	// 拉取跌幅榜 Top100（按涨幅升序）
	fallURL := fmt.Sprintf(
		"https://fund.eastmoney.com/data/rankhandler.aspx?op=ph&dt=kf&ft=all&rs=&gs=0&sc=rzdf&st=asc&sd=%s&ed=%s&qdii=&tabSubtype=,,,,,&pi=1&pn=100&dx=1",
		today, today,
	)
	fallData, err := s.fetchEastmoneyRank(fallURL)
	if err != nil {
		return nil, fmt.Errorf("fetch fall rank failed: %w", err)
	}

	dist := &FundDistribution{
		UpdateAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 用涨幅榜统计涨跌分布（取样200只）
	allFunds := append(riseData, fallData...)
	seen := map[string]bool{}
	var sumReturn float64
	for _, f := range allFunds {
		if seen[f.Code] {
			continue
		}
		seen[f.Code] = true
		dist.Total++
		sumReturn += f.EstReturn
		if f.EstReturn > 0.01 {
			dist.RiseCount++
		} else if f.EstReturn < -0.01 {
			dist.FallCount++
		} else {
			dist.FlatCount++
		}
	}
	if dist.Total > 0 {
		dist.AvgReturn = math.Round(sumReturn/float64(dist.Total)*100) / 100
	}

	// Top5 涨幅（riseData 已降序）
	for i, f := range riseData {
		if i >= 5 {
			break
		}
		dist.TopFunds = append(dist.TopFunds, f)
	}
	// Top5 跌幅（fallData 已升序，取前5即跌幅最大）
	for i, f := range fallData {
		if i >= 5 {
			break
		}
		dist.FlopFunds = append(dist.FlopFunds, f)
	}

	s.setCache(cacheKey, dist, 3*time.Minute)
	return dist, nil
}

// fetchEastmoneyRank 解析天天基金排行榜接口
// 字段顺序（0-indexed）：0=code, 1=name, 2=pinyin, 3=date, 4=unit_nav, 5=acc_nav, 6=daily_return, ...
func (s *MarketService) fetchEastmoneyRank(apiURL string) ([]FundQuote, error) {
	content, err := s.proxyFetchText(apiURL)
	if err != nil {
		return nil, err
	}

	// 响应格式：var rankData = {datas:["code,name,...","code,name,..."],allRecords:N,...}
	// 提取 datas 数组内容
	start := strings.Index(content, `"`)
	end := strings.LastIndex(content, `"`)
	if start < 0 || end <= start {
		return nil, fmt.Errorf("unexpected rankhandler response")
	}
	inner := content[start+1 : end]
	items := strings.Split(inner, `","`)

	var quotes []FundQuote
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		// 字段：0=代码,1=名称,2=拼音,3=日期,4=单位净值,5=累计净值,6=日涨跌%,...
		parts := strings.Split(item, ",")
		if len(parts) < 7 {
			continue
		}
		code := parts[0]
		name := parts[1]
		lastNav := parseSinaFloat(parts[4])
		dailyReturn := parseSinaFloat(parts[6])
		if code == "" || name == "" {
			continue
		}
		quotes = append(quotes, FundQuote{
			Code:       code,
			Name:       name,
			EstReturn:  dailyReturn,
			LastNav:    lastNav,
			UpdateTime: parts[3],
		})
	}
	return quotes, nil
}

// FetchFundQuotes 批量获取指定基金代码的实时估值（天天基金 fundgz 接口）
func (s *MarketService) FetchFundQuotes(codes []string) ([]FundQuote, error) {
	if len(codes) == 0 {
		return []FundQuote{}, nil
	}

	var results []FundQuote
	for _, code := range codes {
		cacheKey := "fundgz_" + code
		if cached := s.getCache(cacheKey); cached != nil {
			results = append(results, cached.(FundQuote))
			continue
		}

		apiURL := fmt.Sprintf("https://fundgz.1234567.com.cn/js/%s.js?rt=%d", code, time.Now().Unix())
		content, err := s.proxyFetchText(apiURL)
		if err != nil {
			continue
		}

		// 格式：jsonpgz({"fundcode":"003095","name":"...","jzrq":"2026-06-11","dwjz":"1.5353","gsz":"1.5661","gszzl":"2.01","gztime":"2026-06-12 15:00"});
		jsonStart := strings.Index(content, "(")
		jsonEnd := strings.LastIndex(content, ")")
		if jsonStart < 0 || jsonEnd <= jsonStart {
			continue
		}
		jsonStr := content[jsonStart+1 : jsonEnd]

		var raw struct {
			Fundcode string `json:"fundcode"`
			Name     string `json:"name"`
			Jzrq     string `json:"jzrq"`   // 净值日期
			Dwjz     string `json:"dwjz"`   // 单位净值
			Gsz      string `json:"gsz"`    // 估算净值
			Gszzl    string `json:"gszzl"`  // 估算涨跌幅 %
			Gztime   string `json:"gztime"` // 估值时间
		}
		if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
			continue
		}

		q := FundQuote{
			Code:       raw.Fundcode,
			Name:       raw.Name,
			EstReturn:  parseSinaFloat(raw.Gszzl),
			EstNav:     parseSinaFloat(raw.Gsz),
			LastNav:    parseSinaFloat(raw.Dwjz),
			UpdateTime: raw.Gztime,
		}
		s.setCache(cacheKey, q, 3*time.Minute)
		results = append(results, q)
	}

	return results, nil
}

// ─── 分时指数 ────────────────────────────────────────────────────────────

// FundNavPoint 单条历史净值
type FundNavPoint struct {
	Date        string  `json:"date"`
	UnitNav     float64 `json:"unitNav"`
	AccNav      float64 `json:"accNav"`
	DailyReturn float64 `json:"dailyReturn"`
}

// FetchFundNavHistory 从天天基金接口拉取基金历史净值
// 接口：https://fund.eastmoney.com/f10/F10DataApi.aspx?type=lsjz&code=007339&page=1&sdate=&edate=&per=120
func (s *MarketService) FetchFundNavHistory(code string, days int) ([]FundNavPoint, error) {
	cacheKey := fmt.Sprintf("nav_history_%s_%d", code, days)
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]FundNavPoint), nil
	}

	// 计算起始日期（多拿一点）
	startDate := time.Now().AddDate(0, 0, -days*2).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")
	per := days
	if per > 200 {
		per = 200
	}

	apiURL := fmt.Sprintf(
		"https://fund.eastmoney.com/f10/F10DataApi.aspx?type=lsjz&code=%s&page=1&sdate=%s&edate=%s&per=%d",
		code, startDate, endDate, per,
	)

	content, err := s.proxyFetchText(apiURL)
	if err != nil {
		return nil, fmt.Errorf("fetch nav history failed: %w", err)
	}

	// 响应格式：var apidata={content:"<table>...</table>",records:120,...}
	// 从 records 字段判断有没有数据，从 content 解析 HTML 表格太脆
	// 改用 JSON 接口：https://api.fund.eastmoney.com/f10/lsjz
	apiURL2 := fmt.Sprintf(
		"https://api.fund.eastmoney.com/f10/lsjz?fundCode=%s&pageIndex=1&pageSize=%d&startDate=%s&endDate=%s&callback=cb",
		code, per, startDate, endDate,
	)
	content, err = s.proxyFetchText(apiURL2)
	if err != nil {
		return nil, fmt.Errorf("fetch nav history json failed: %w", err)
	}

	// 去掉 callback 包裹
	jsonStart := strings.Index(content, "(")
	jsonEnd := strings.LastIndex(content, ")")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		content = content[jsonStart+1 : jsonEnd]
	}

	var result struct {
		Data struct {
			LSJZList []struct {
				FSRQ string `json:"FSRQ"` // 净值日期
				DWJZ string `json:"DWJZ"` // 单位净值
				LJJZ string `json:"LJJZ"` // 累计净值
				JZZZL string `json:"JZZZL"` // 日增长率 %
			} `json:"LSJZList"`
		} `json:"Data"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parse nav history failed: %w", err)
	}

	list := result.Data.LSJZList
	// 天天基金返回的是降序（最新在前），反转为升序
	points := make([]FundNavPoint, 0, len(list))
	for i := len(list) - 1; i >= 0; i-- {
		item := list[i]
		unitNav := parseSinaFloat(item.DWJZ)
		accNav := parseSinaFloat(item.LJJZ)
		dailyReturn := parseSinaFloat(item.JZZZL)
		if unitNav == 0 {
			continue
		}
		points = append(points, FundNavPoint{
			Date:        item.FSRQ,
			UnitNav:     unitNav,
			AccNav:      accNav,
			DailyReturn: dailyReturn,
		})
	}

	// 截取最近 days 条
	if len(points) > days {
		points = points[len(points)-days:]
	}

	if len(points) > 0 {
		s.setCache(cacheKey, points, 30*time.Minute)
	}

	return points, nil
}

// FundBasicInfo 基金基础信息（用于自动播种）
type FundBasicInfo struct {
	Code      string
	Name      string
	FundType  string
}

// FetchFundInfo 从天天基金搜索接口获取基金基础信息
// 接口：https://fund.eastmoney.com/js/fundcode_search.js
// 返回格式：var r = [["000001","HXCZHH","华夏成长混合","混合型","HUAXIACHENGZHANGHUNHE"],...];
func (s *MarketService) FetchFundInfo(code string) (*FundBasicInfo, error) {
	cacheKey := "fundinfo_" + code
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.(*FundBasicInfo), nil
	}

	// 先尝试从 fundgz 接口获取名称（已有 FetchFundQuotes 的逻辑）
	quotes, err := s.FetchFundQuotes([]string{code})
	if err == nil && len(quotes) > 0 && quotes[0].Name != "" {
		info := &FundBasicInfo{
			Code: code,
			Name: quotes[0].Name,
		}
		s.setCache(cacheKey, info, 24*time.Hour)
		return info, nil
	}

	// fallback：天天基金搜索接口
	apiURL := fmt.Sprintf(
		"https://fundsuggest.eastmoney.com/FundSearch/api/FundSearchAPI.ashx?callback=cb&m=1&key=%s",
		url.QueryEscape(code),
	)
	content, err := s.proxyFetchText(apiURL)
	if err != nil {
		return nil, fmt.Errorf("fetch fund info failed: %w", err)
	}

	// 响应格式：cb({"Datas":[{"CODE":"007339","NAME":"...","FundType":"..."}]})
	jsonStart := strings.Index(content, "(")
	jsonEnd := strings.LastIndex(content, ")")
	if jsonStart < 0 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("unexpected fund search response")
	}
	jsonStr := content[jsonStart+1 : jsonEnd]

	var result struct {
		Datas []struct {
			Code     string `json:"CODE"`
			Name     string `json:"NAME"`
			FundType string `json:"FundType"`
		} `json:"Datas"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parse fund search response failed: %w", err)
	}

	for _, d := range result.Datas {
		if d.Code == code {
			info := &FundBasicInfo{
				Code:     d.Code,
				Name:     d.Name,
				FundType: d.FundType,
			}
			s.setCache(cacheKey, info, 24*time.Hour)
			return info, nil
		}
	}

	return nil, fmt.Errorf("fund %s not found in search results", code)
}

// IntradayPoint 分时数据点
type IntradayPoint struct {
	Time  string  `json:"time"`  // "09:35"
	Price float64 `json:"price"` // 当前价
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Vol   float64 `json:"vol"` // 成交量（手）
}

// IntradayData 分时数据
type IntradayData struct {
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	PreClose  float64         `json:"preClose"` // 昨收
	Points    []IntradayPoint `json:"points"`
	UpdateAt  string          `json:"updateAt"`
}

// emToSinaCode 将东方财富 secid 转为新浪 symbol
// 1.000001 → sh000001, 0.399001 → sz399001
func emToSinaCode(emCode string) string {
	parts := strings.SplitN(emCode, ".", 2)
	if len(parts) != 2 {
		return emCode
	}
	market, symbol := parts[0], parts[1]
	if market == "1" {
		return "sh" + symbol
	}
	return "sz" + symbol
}

// FetchIntraday 通过新浪财经 API 获取指数当天分时数据（5分钟K线）
func (s *MarketService) FetchIntraday(code string) (*IntradayData, error) {
	cacheKey := "intraday_" + code
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.(*IntradayData), nil
	}

	sinaCode := emToSinaCode(code)
	// 新浪 5 分钟 K 线，datalen=48 覆盖全天 4 小时
	apiURL := fmt.Sprintf(
		"https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=%s&scale=5&ma=no&datalen=48",
		sinaCode,
	)

	content, err := s.proxyFetchText(apiURL)
	if err != nil {
		return nil, fmt.Errorf("fetch intraday %s failed: %w", code, err)
	}

	// 新浪返回 JSON 数组：[{"day":"2026-06-15 09:35:00","open":"...","high":"...","low":"...","close":"...","volume":"..."}, ...]
	var rawBars []struct {
		Day    string `json:"day"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	if err := json.Unmarshal([]byte(content), &rawBars); err != nil {
		return nil, fmt.Errorf("parse intraday json failed: %w", err)
	}

	result := &IntradayData{
		Code:     code,
		Name:     yahooIndexName(code), // 复用已有的名称映射
		UpdateAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 只保留今天的数据
	today := time.Now().Format("2006-01-02")
	var todayBars []struct {
		Day    string `json:"day"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	for _, bar := range rawBars {
		if strings.HasPrefix(bar.Day, today) {
			todayBars = append(todayBars, bar)
		}
	}

	// 如果今天没有数据（非交易日或未开盘），取最后的数据展示
	if len(todayBars) == 0 {
		todayBars = rawBars
	}

	for _, bar := range todayBars {
		// 取时间部分 "2026-06-15 09:35:00" → "09:35"
		timePart := bar.Day
		if idx := strings.Index(timePart, " "); idx >= 0 {
			timePart = timePart[idx+1:]
		}
		// 去掉秒 "09:35:00" → "09:35"
		if len(timePart) > 5 {
			timePart = timePart[:5]
		}

		result.Points = append(result.Points, IntradayPoint{
			Time:  timePart,
			Price: parseSinaFloat(bar.Close),
			Open:  parseSinaFloat(bar.Open),
			High:  parseSinaFloat(bar.High),
			Low:   parseSinaFloat(bar.Low),
			Vol:   parseSinaFloat(bar.Volume),
		})
	}

	// 计算昨收：取第一个点的开盘价作为近似昨收（新浪不直接提供）
	if len(result.Points) > 0 {
		result.PreClose = result.Points[0].Open
	}

	s.setCache(cacheKey, result, 30*time.Second)
	return result, nil
}
