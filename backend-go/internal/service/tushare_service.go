package service

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/jishengdan/backend-go/pkg/tushare"
)

type TushareService struct {
	client *tushare.Client
	cache  sync.Map
}

func NewTushareService(client *tushare.Client) *TushareService {
	return &TushareService{client: client}
}

type IndexDataTS struct {
	Name          string
	Code          string
	Price         float64
	Change        float64
	ChangePercent float64
	Volume        string
	High          float64
	Low           float64
	Open          float64
	PreClose      float64
	Amount        string
	UpdateAt      string
}

type KLineDataTS struct {
	Date  string
	Open  float64
	Close float64
	High  float64
	Low   float64
}

type FundNavPointTS struct {
	Date        string
	UnitNav     float64
	AccNav      float64
	DailyReturn float64
}

type FundBasicInfoTS struct {
	Code     string
	Name     string
	FundType string
}

// tsCodeMap maps EastMoney secid (1.000001) to Tushare ts_code (000001.SH)
func emToTsCode(code string) string {
	if len(code) < 2 || code[1] != '.' {
		return code
	}
	market := code[0]
	symbol := code[2:]
	switch market {
	case '1':
		return symbol + ".SH"
	case '0':
		return symbol + ".SZ"
	default:
		return code
	}
}

// tsToEmCode maps ts_code back to EastMoney format
func tsToEmCode(tsCode string) string {
	if len(tsCode) < 8 {
		return tsCode
	}
	symbol := tsCode[:6]
	suffix := tsCode[7:]
	switch suffix {
	case "SH":
		return "1." + symbol
	case "SZ":
		if symbol[:1] == "3" || symbol[:1] == "0" {
			return "0." + symbol
		}
		return "1." + symbol
	default:
		return tsCode
	}
}

// GetIndex returns latest index quotes for given codes (in EastMoney secid format)
func (s *TushareService) GetIndex(codes []string) ([]IndexDataTS, error) {
	cacheKey := "ts_idx_" + fmt.Sprintf("%v", codes)
	if v, ok := s.cache.Load(cacheKey); ok {
		return v.([]IndexDataTS), nil
	}

	var results []IndexDataTS
	for _, emCode := range codes {
		tsCode := emToTsCode(emCode)
		data, err := s.client.IndexDaily(tsCode, "", "", 1)
		if err != nil || len(data) == 0 {
			continue
		}
		d := data[0]
		change := d.Close - d.PreClose
		changePct := 0.0
		if d.PreClose > 0 {
			changePct = (change / d.PreClose) * 100
		}
		results = append(results, IndexDataTS{
			Name:          indexDisplayName(emCode),
			Code:          emCode,
			Price:         math.Round(d.Close*100) / 100,
			Change:        math.Round(change*100) / 100,
			ChangePercent: math.Round(changePct*100) / 100,
			High:          d.High,
			Low:           d.Low,
			Open:          d.Open,
			PreClose:      d.PreClose,
			Volume:        fmt.Sprintf("%.0f亿手", d.Vol/1e8),
			Amount:        fmt.Sprintf("%.0f亿", d.Amount/1e8),
			UpdateAt:      formatTradeDate(d.TradeDate),
		})
	}

	if len(results) > 0 {
		s.cache.Store(cacheKey, results)
		time.AfterFunc(30*time.Second, func() { s.cache.Delete(cacheKey) })
	}

	return results, nil
}

// GetIndexHistory returns index history K-line
func (s *TushareService) GetIndexHistory(code string, days int) ([]KLineDataTS, error) {
	cacheKey := fmt.Sprintf("ts_idxh_%s_%d", code, days)
	if v, ok := s.cache.Load(cacheKey); ok {
		return v.([]KLineDataTS), nil
	}

	tsCode := emToTsCode(code)
	endDate := time.Now().Format("20060102")
	startDate := time.Now().AddDate(0, 0, -days*2).Format("20060102")

	data, err := s.client.IndexDaily(tsCode, startDate, endDate, days*2)
	if err != nil {
		return nil, fmt.Errorf("fetch index history: %w", err)
	}

	// data is already in ascending order from Tushare (oldest first)
	// Sort ascending by trade_date just in case
	sort.Slice(data, func(i, j int) bool {
		return data[i].TradeDate < data[j].TradeDate
	})

	klines := make([]KLineDataTS, 0, len(data))
	for _, d := range data {
		klines = append(klines, KLineDataTS{
			Date:  formatTradeDate(d.TradeDate),
			Open:  math.Round(d.Open*100) / 100,
			Close: math.Round(d.Close*100) / 100,
			High:  math.Round(d.High*100) / 100,
			Low:   math.Round(d.Low*100) / 100,
		})
	}

	if len(klines) > days {
		klines = klines[len(klines)-days:]
	}

	if len(klines) > 0 {
		s.cache.Store(cacheKey, klines)
		time.AfterFunc(5*time.Minute, func() { s.cache.Delete(cacheKey) })
	}

	return klines, nil
}

// GetFundNav returns latest NAV for a fund code
func (s *TushareService) GetFundNav(code string) (*FundNavPointTS, error) {
	cacheKey := "ts_fnav_" + code
	if v, ok := s.cache.Load(cacheKey); ok {
		return v.(*FundNavPointTS), nil
	}

	tsCode := code + ".OF"
	data, err := s.client.FundDaily(tsCode, "", "", 1)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("fund %s nav not found", code)
	}

	d := data[0]
	pt := &FundNavPointTS{
		Date:        formatTradeDate(d.TradeDate),
		UnitNav:     d.UnitNav,
		AccNav:      d.AccumNav,
		DailyReturn: d.DailyReturn,
	}

	s.cache.Store(cacheKey, pt)
	time.AfterFunc(30*time.Minute, func() { s.cache.Delete(cacheKey) })
	return pt, nil
}

// GetFundNavHistory returns fund NAV history for given days
func (s *TushareService) GetFundNavHistory(code string, days int) ([]FundNavPointTS, error) {
	cacheKey := fmt.Sprintf("ts_fnavh_%s_%d", code, days)
	if v, ok := s.cache.Load(cacheKey); ok {
		return v.([]FundNavPointTS), nil
	}

	tsCode := code + ".OF"
	endDate := time.Now().Format("20060102")
	startDate := time.Now().AddDate(0, 0, -days*2).Format("20060102")

	data, err := s.client.FundDaily(tsCode, startDate, endDate, days*2)
	if err != nil {
		return nil, fmt.Errorf("fetch fund nav history: %w", err)
	}

	// Sort ascending by trade_date
	sort.Slice(data, func(i, j int) bool {
		return data[i].TradeDate < data[j].TradeDate
	})

	points := make([]FundNavPointTS, 0, len(data))
	for _, d := range data {
		points = append(points, FundNavPointTS{
			Date:        formatTradeDate(d.TradeDate),
			UnitNav:     d.UnitNav,
			AccNav:      d.AccumNav,
			DailyReturn: d.DailyReturn,
		})
	}

	if len(points) > days {
		points = points[len(points)-days:]
	}

	if len(points) > 0 {
		s.cache.Store(cacheKey, points)
		time.AfterFunc(30*time.Minute, func() { s.cache.Delete(cacheKey) })
	}

	return points, nil
}

// GetFundInfo returns fund basic info
func (s *TushareService) GetFundInfo(code string) (*FundBasicInfoTS, error) {
	cacheKey := "ts_finfo_" + code
	if v, ok := s.cache.Load(cacheKey); ok {
		return v.(*FundBasicInfoTS), nil
	}

	tsCode := code + ".OF"
	info, err := s.client.FundBasic(tsCode)
	if err != nil {
		return nil, err
	}

	fi := &FundBasicInfoTS{
		Code:     code,
		Name:     info.Name,
		FundType: info.FundType,
	}

	s.cache.Store(cacheKey, fi)
	time.AfterFunc(24*time.Hour, func() { s.cache.Delete(cacheKey) })
	return fi, nil
}

func formatTradeDate(s string) string {
	if len(s) == 8 {
		return s[:4] + "-" + s[4:6] + "-" + s[6:8]
	}
	return s
}
