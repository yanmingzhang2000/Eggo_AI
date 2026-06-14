package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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
		client: &http.Client{Timeout: 10 * time.Second},
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

// GetCache 缓存辅助
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

// FetchSinaIndex 从新浪财经获取指数行情
func (s *MarketService) FetchSinaIndex(codes []string) ([]IndexData, error) {
	cacheKey := "sina_indices"
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]IndexData), nil
	}

	codeStr := strings.Join(codes, ",")
	url := fmt.Sprintf("https://hq.sinajs.cn/list=%s", codeStr)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Referer", "https://finance.sina.com.cn")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch sina index failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	reader := transform.NewReader(strings.NewReader(string(body)), simplifiedchinese.GBK.NewDecoder())
	decoded, _ := io.ReadAll(reader)
	content := string(decoded)

	var results []IndexData
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析 var hq_str_sh000001="上证指数,3387.52,...";
		re := regexp.MustCompile(`var hq_str_(\w+)="(.*)";`)
		matches := re.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		code := matches[1]
		data := matches[2]
		fields := strings.Split(data, ",")
		if len(fields) < 32 {
			continue
		}

		idx := IndexData{
			Name:     fields[0],
			Code:     code,
			Open:     parseSinaFloat(fields[1]),
			PreClose: parseSinaFloat(fields[2]),
			Price:    parseSinaFloat(fields[3]),
			High:     parseSinaFloat(fields[4]),
			Low:      parseSinaFloat(fields[5]),
			Amount:   formatAmount(parseSinaFloat(fields[9])),
		}
		idx.Change = idx.Price - idx.PreClose
		if idx.PreClose > 0 {
			idx.ChangePercent = (idx.Change / idx.PreClose) * 100
		}
		idx.UpdateAt = time.Now().Format("2006-01-02 15:04:05")
		idx.Volume = formatVolume(parseSinaFloat(fields[8]))

		results = append(results, idx)
	}

	if len(results) > 0 {
		s.setCache(cacheKey, results, 30*time.Second)
	}

	return results, nil
}

// FetchEastmoneySectors 从东方财富获取板块行情
func (s *MarketService) FetchEastmoneySectors() ([]SectorData, error) {
	cacheKey := "eastmoney_sectors"
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]SectorData), nil
	}

	// 东方财富行业板块接口
	url := "https://push2.eastmoney.com/api/qt/clist/get?cb=&fid=f3&po=1&pz=10&pn=1&np=1&fltt=2&invt=2&fs=m:90+t:2&fields=f2,f3,f4,f12,f14"

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch eastmoney sectors failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			Diff []struct {
				F2  float64 `json:"f2"`
				F3  float64 `json:"f3"`
				F12 string  `json:"f12"`
				F14 string  `json:"f14"`
			} `json:"diff"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse eastmoney sectors failed: %w", err)
	}

	var sectors []SectorData
	for _, item := range result.Data.Diff {
		sectors = append(sectors, SectorData{
			Name:          item.F14,
			Code:          item.F12,
			ChangePercent: item.F3,
			Price:         item.F2,
		})
	}

	if len(sectors) > 0 {
		s.setCache(cacheKey, sectors, 60*time.Second)
	}

	return sectors, nil
}

// FetchEastmoneyConcepts 从东方财富获取概念板块
func (s *MarketService) FetchEastmoneyConcepts() ([]SectorData, error) {
	cacheKey := "eastmoney_concepts"
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]SectorData), nil
	}

	url := "https://push2.eastmoney.com/api/qt/clist/get?cb=&fid=f3&po=1&pz=10&pn=1&np=1&fltt=2&invt=2&fs=m:90+t:3&fields=f2,f3,f4,f12,f14"

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch eastmoney concepts failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			Diff []struct {
				F2  float64 `json:"f2"`
				F3  float64 `json:"f3"`
				F12 string  `json:"f12"`
				F14 string  `json:"f14"`
			} `json:"diff"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse eastmoney concepts failed: %w", err)
	}

	var sectors []SectorData
	for _, item := range result.Data.Diff {
		sectors = append(sectors, SectorData{
			Name:          item.F14,
			Code:          item.F12,
			ChangePercent: item.F3,
			Price:         item.F2,
		})
	}

	if len(sectors) > 0 {
		s.setCache(cacheKey, sectors, 60*time.Second)
	}

	return sectors, nil
}

// FetchMarketOverview 获取市场概览（涨跌家数、涨停跌停等）
func (s *MarketService) FetchMarketOverview() (*MarketStats, error) {
	cacheKey := "market_overview"
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.(*MarketStats), nil
	}

	// 获取全市场A股数据来统计涨跌
	url := "https://push2.eastmoney.com/api/qt/clist/get?cb=&fid=f3&po=1&pz=5000&pn=1&np=1&fltt=2&invt=2&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048&fields=f2,f3,f12,f14"

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch market overview failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			Total int `json:"total"`
			Diff  []struct {
				F2  float64 `json:"f2"`
				F3  float64 `json:"f3"`
				F12 string  `json:"f12"`
				F14 string  `json:"f14"`
			} `json:"diff"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse market overview failed: %w", err)
	}

	stats := &MarketStats{
		UpdateAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	for _, item := range result.Data.Diff {
		if item.F3 > 0 {
			stats.UpCount++
		} else if item.F3 < 0 {
			stats.DownCount++
		} else {
			stats.FlatCount++
		}

		// 涨停：涨幅 >= 9.9%
		if item.F3 >= 9.9 {
			stats.LimitUp++
		}
		// 跌停：跌幅 <= -9.9%
		if item.F3 <= -9.9 {
			stats.LimitDown++
		}
	}

	total := stats.UpCount + stats.DownCount + stats.FlatCount
	if stats.UpCount > total*60/100 {
		stats.Sentiment = "强势"
	} else if stats.UpCount > total*50/100 {
		stats.Sentiment = "偏多"
	} else if stats.DownCount > total*60/100 {
		stats.Sentiment = "弱势"
	} else if stats.DownCount > total*50/100 {
		stats.Sentiment = "偏空"
	} else {
		stats.Sentiment = "震荡"
	}

	s.setCache(cacheKey, stats, 30*time.Second)
	return stats, nil
}

// FetchNorthFlow 获取北向资金数据
func (s *MarketService) FetchNorthFlow() string {
	url := "https://push2.eastmoney.com/api/qt/kamtbs.wss?fields1=f1,f2,f3,f4&fields2=f51,f52,f53,f54,f55,f56"

	resp, err := s.client.Get(url)
	if err != nil {
		return "--"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			S2N struct {
				F52 float64 `json:"f52"` // 沪股通净流入
				F54 float64 `json:"f54"` // 深股通净流入
			} `json:"s2n"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "--"
	}

	total := result.Data.S2N.F52 + result.Data.S2N.F54
	if total >= 0 {
		return fmt.Sprintf("+%.1f亿", total)
	}
	return fmt.Sprintf("%.1f亿", total)
}

func parseSinaFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
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
