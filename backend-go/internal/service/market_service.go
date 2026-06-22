package service

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type MarketService struct {
	client  *http.Client
	cache   map[string]cacheEntry
	cacheMu sync.RWMutex
	tushare *TushareService
}

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

func NewMarketService(ts *TushareService) *MarketService {
	return &MarketService{
		client:  &http.Client{Timeout: 20 * time.Second},
		cache:   make(map[string]cacheEntry),
		tushare: ts,
	}
}

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

type SectorData struct {
	Name          string  `json:"name"`
	Code          string  `json:"code,omitempty"`
	ChangePercent float64 `json:"changePercent"`
	Price         float64 `json:"price"`
	LeadStock     string  `json:"leadStock,omitempty"`
	Reason        string  `json:"reason,omitempty"`
	Amount        string  `json:"amount,omitempty"`
}

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

func (s *MarketService) FetchSinaIndex(codes []string) ([]IndexData, error) {
	cacheKey := "ts_indices_" + strings.Join(codes, "_")
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]IndexData), nil
	}

	tsData, err := s.tushare.GetIndex(codes)
	if err != nil {
		return nil, err
	}

	results := make([]IndexData, 0, len(tsData))
	for _, d := range tsData {
		results = append(results, IndexData{
			Name:          d.Name,
			Code:          d.Code,
			Price:         d.Price,
			Change:        d.Change,
			ChangePercent: d.ChangePercent,
			Volume:        d.Volume,
			High:          d.High,
			Low:           d.Low,
			Open:          d.Open,
			PreClose:      d.PreClose,
			Amount:        d.Amount,
			UpdateAt:      d.UpdateAt,
		})
	}

	if len(results) > 0 {
		s.setCache(cacheKey, results, 30*time.Second)
	}
	return results, nil
}

func (s *MarketService) FetchEastmoneySectors() ([]SectorData, error) {
	return []SectorData{}, nil
}

func (s *MarketService) FetchEastmoneyConcepts() ([]SectorData, error) {
	return []SectorData{}, nil
}

func (s *MarketService) FetchMarketOverview() (*MarketStats, error) {
	cacheKey := "market_overview"
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.(*MarketStats), nil
	}

	stats := &MarketStats{
		NorthFlow: "--",
		UpdateAt:  time.Now().Format("2006-01-02 15:04:05"),
	}

	indices, err := s.tushare.GetIndex([]string{"1.000300"})
	if err == nil && len(indices) > 0 {
		cp := indices[0].ChangePercent
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

func (s *MarketService) FetchNorthFlow() string {
	return "--"
}

func parseSinaFloat(s string) float64 {
	var f float64
	fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
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

type KLineData struct {
	Date  string  `json:"date"`
	Open  float64 `json:"open"`
	Close float64 `json:"close"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
}

type IndexOption struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

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

func (s *MarketService) FetchIndexHistory(code string, days int) ([]KLineData, error) {
	cacheKey := fmt.Sprintf("index_history_%s_%d", code, days)
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]KLineData), nil
	}

	tsData, err := s.tushare.GetIndexHistory(code, days)
	if err != nil {
		return nil, err
	}

	klines := make([]KLineData, 0, len(tsData))
	for _, d := range tsData {
		klines = append(klines, KLineData{
			Date:  d.Date,
			Open:  d.Open,
			Close: d.Close,
			High:  d.High,
			Low:   d.Low,
		})
	}

	if len(klines) > 0 {
		s.setCache(cacheKey, klines, 5*time.Minute)
	}
	return klines, nil
}

type FundDistribution struct {
	RiseCount  int         `json:"riseCount"`
	FallCount  int         `json:"fallCount"`
	FlatCount  int         `json:"flatCount"`
	Total      int         `json:"total"`
	AvgReturn  float64     `json:"avgReturn"`
	TopFunds   []FundQuote `json:"topFunds"`
	FlopFunds  []FundQuote `json:"flopFunds"`
	UpdateAt   string      `json:"updateAt"`
}

type FundQuote struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	EstReturn  float64 `json:"estReturn"`
	EstNav     float64 `json:"estNav"`
	LastNav    float64 `json:"lastNav"`
	UpdateTime string  `json:"updateTime"`
}

func (s *MarketService) FetchFundDistribution() (*FundDistribution, error) {
	cacheKey := "fund_distribution"
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.(*FundDistribution), nil
	}

	sampleCodes := []string{
		"110011", "000001", "000002", "001810", "161725", "519674", "002963", "006479", "009547", "001718",
		"008763", "008764", "519983", "000961", "001975", "000248", "270002", "003095", "006228", "001856",
		"000592", "008294", "001974", "110022", "001632", "000369", "005827", "519915", "001714", "000478",
		"000191", "007119", "003984", "001643", "002001", "007339", "004746", "163417", "260108", "161017",
		"161005", "000300", "510050", "510300", "510500", "512100", "159915", "159928", "515050", "560003",
	}

	quotes, err := s.FetchFundQuotes(sampleCodes)
	if err != nil {
		return nil, fmt.Errorf("fetch sample fund quotes failed: %w", err)
	}

	dist := &FundDistribution{
		UpdateAt:  time.Now().Format("2006-01-02 15:04:05"),
		TopFunds:  []FundQuote{},
		FlopFunds: []FundQuote{},
	}

	var sumReturn float64
	for _, q := range quotes {
		dist.Total++
		sumReturn += q.EstReturn
		if q.EstReturn > 0.01 {
			dist.RiseCount++
		} else if q.EstReturn < -0.01 {
			dist.FallCount++
		} else {
			dist.FlatCount++
		}
	}
	if dist.Total > 0 {
		dist.AvgReturn = math.Round(sumReturn/float64(dist.Total)*100) / 100
	}

	sortedQuotes := make([]FundQuote, len(quotes))
	copy(sortedQuotes, quotes)
	sort.Slice(sortedQuotes, func(i, j int) bool {
		return sortedQuotes[i].EstReturn > sortedQuotes[j].EstReturn
	})

	for i := 0; i < 5 && i < len(sortedQuotes); i++ {
		dist.TopFunds = append(dist.TopFunds, sortedQuotes[i])
	}
	for i := len(sortedQuotes) - 1; i >= 0 && len(dist.FlopFunds) < 5; i-- {
		dist.FlopFunds = append(dist.FlopFunds, sortedQuotes[i])
	}

	s.setCache(cacheKey, dist, 3*time.Minute)
	return dist, nil
}

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
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Referer", "https://fund.eastmoney.com/")

		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		content := string(body)

		jsonStart := strings.Index(content, "(")
		jsonEnd := strings.LastIndex(content, ")")
		if jsonStart < 0 || jsonEnd <= jsonStart {
			continue
		}
		jsonStr := content[jsonStart+1 : jsonEnd]

		var raw struct {
			Fundcode string `json:"fundcode"`
			Name     string `json:"name"`
			Jzrq     string `json:"jzrq"`
			Dwjz     string `json:"dwjz"`
			Gsz      string `json:"gsz"`
			Gszzl    string `json:"gszzl"`
			Gztime   string `json:"gztime"`
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

type FundNavPoint struct {
	Date        string  `json:"date"`
	UnitNav     float64 `json:"unitNav"`
	AccNav      float64 `json:"accNav"`
	DailyReturn float64 `json:"dailyReturn"`
}

func (s *MarketService) FetchFundNavHistory(code string, days int) ([]FundNavPoint, error) {
	cacheKey := fmt.Sprintf("nav_history_%s_%d", code, days)
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]FundNavPoint), nil
	}

	tsData, err := s.tushare.GetFundNavHistory(code, days)
	if err != nil {
		return nil, err
	}

	points := make([]FundNavPoint, 0, len(tsData))
	for _, d := range tsData {
		points = append(points, FundNavPoint{
			Date:        d.Date,
			UnitNav:     d.UnitNav,
			AccNav:      d.AccNav,
			DailyReturn: d.DailyReturn,
		})
	}

	if len(points) > 0 {
		s.setCache(cacheKey, points, 30*time.Minute)
	}
	return points, nil
}

type FundBasicInfo struct {
	Code     string
	Name     string
	FundType string
}

func (s *MarketService) FetchFundInfo(code string) (*FundBasicInfo, error) {
	cacheKey := "fundinfo_" + code
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.(*FundBasicInfo), nil
	}

	info, err := s.tushare.GetFundInfo(code)
	if err != nil {
		return nil, err
	}

	fi := &FundBasicInfo{
		Code:     info.Code,
		Name:     info.Name,
		FundType: info.FundType,
	}

	s.setCache(cacheKey, fi, 24*time.Hour)
	return fi, nil
}

type IntradayPoint struct {
	Time  string  `json:"time"`
	Price float64 `json:"price"`
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Vol   float64 `json:"vol"`
}

type IntradayData struct {
	Code     string          `json:"code"`
	Name     string          `json:"name"`
	PreClose float64         `json:"preClose"`
	Points   []IntradayPoint `json:"points"`
	UpdateAt string          `json:"updateAt"`
}

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

func indexDisplayName(code string) string {
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

func (s *MarketService) FetchIntraday(code string) (*IntradayData, error) {
	cacheKey := "intraday_" + code
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.(*IntradayData), nil
	}

	sinaCode := emToSinaCode(code)
	apiURL := fmt.Sprintf(
		"https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=%s&scale=5&ma=no&datalen=48",
		sinaCode,
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch intraday %s failed: %w", code, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rawBars []struct {
		Day    string `json:"day"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	if err := json.Unmarshal(body, &rawBars); err != nil {
		return nil, fmt.Errorf("parse intraday json failed: %w", err)
	}

	result := &IntradayData{
		Code:     code,
		Name:     indexDisplayName(code),
		UpdateAt: time.Now().Format("2006-01-02 15:04:05"),
	}

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
	if len(todayBars) == 0 {
		todayBars = rawBars
	}

	for _, bar := range todayBars {
		timePart := bar.Day
		if idx := strings.Index(timePart, " "); idx >= 0 {
			timePart = timePart[idx+1:]
		}
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

	if len(result.Points) > 0 {
		result.PreClose = result.Points[0].Open
	}

	s.setCache(cacheKey, result, 30*time.Second)
	return result, nil
}
