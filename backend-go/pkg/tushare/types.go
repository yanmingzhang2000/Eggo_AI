package tushare

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type IndexDaily struct {
	TsCode     string
	TradeDate  string
	Close      float64
	Open       float64
	High       float64
	Low        float64
	PreClose   float64
	Change     float64
	PctChange  float64
	Vol        float64
	Amount     float64
}

type FundDaily struct {
	TsCode      string
	TradeDate   string
	UnitNav     float64
	AccumNav    float64
	DailyReturn float64
}

type FundBasic struct {
	TsCode     string
	Name       string
	FundType   string
	Management string
}

func (c *Client) IndexDaily(tsCode string, startDate string, endDate string, limit int) ([]IndexDaily, error) {
	params := map[string]interface{}{"ts_code": tsCode}
	if startDate != "" {
		params["start_date"] = startDate
	}
	if endDate != "" {
		params["end_date"] = endDate
	}
	if limit > 0 {
		params["limit"] = limit
	}

	fields := []string{"ts_code", "trade_date", "close", "open", "high", "low", "pre_close", "change", "pct_chg", "vol", "amount"}
	resp, err := c.Query("index_daily", params, fields)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}

	result := make([]IndexDaily, 0, len(resp.Data.Items))
	for _, item := range resp.Data.Items {
		var row []interface{}
		if err := json.Unmarshal(item, &row); err != nil {
			continue
		}
		if len(row) < 9 {
			continue
		}

		d := IndexDaily{TsCode: toString(row[0]), TradeDate: toString(row[1])}
		d.Close = toFloat(row[2])
		d.Open = toFloat(row[3])
		d.High = toFloat(row[4])
		d.Low = toFloat(row[5])
		if len(row) > 6 {
			d.PreClose = toFloat(row[6])
		}
		if len(row) > 7 {
			d.Change = toFloat(row[7])
		}
		if len(row) > 8 {
			d.PctChange = toFloat(row[8])
		}
		if len(row) > 9 {
			d.Vol = toFloat(row[9])
		}
		if len(row) > 10 {
			d.Amount = toFloat(row[10])
		}
		result = append(result, d)
	}
	return result, nil
}

func (c *Client) FundDaily(tsCode string, startDate string, endDate string, limit int) ([]FundDaily, error) {
	params := map[string]interface{}{"ts_code": tsCode}
	if startDate != "" {
		params["start_date"] = startDate
	}
	if endDate != "" {
		params["end_date"] = endDate
	}
	if limit > 0 {
		params["limit"] = limit
	}

	fields := []string{"ts_code", "trade_date", "unit_nav", "accum_nav", "daily_return"}
	resp, err := c.Query("fund_daily", params, fields)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}

	result := make([]FundDaily, 0, len(resp.Data.Items))
	for _, item := range resp.Data.Items {
		var row []interface{}
		if err := json.Unmarshal(item, &row); err != nil {
			continue
		}
		if len(row) < 3 {
			continue
		}

		d := FundDaily{TsCode: toString(row[0]), TradeDate: toString(row[1])}
		d.UnitNav = toFloat(row[2])
		if len(row) > 3 {
			d.AccumNav = toFloat(row[3])
		}
		if len(row) > 4 {
			d.DailyReturn = toFloat(row[4])
		}
		result = append(result, d)
	}
	return result, nil
}

func (c *Client) FundBasic(tsCode string) (*FundBasic, error) {
	params := map[string]interface{}{"ts_code": tsCode}
	fields := []string{"ts_code", "name", "fund_type", "management"}
	resp, err := c.Query("fund_basic", params, fields)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil || len(resp.Data.Items) == 0 {
		return nil, fmt.Errorf("fund %s not found", tsCode)
	}

	var row []interface{}
	if err := json.Unmarshal(resp.Data.Items[0], &row); err != nil {
		return nil, err
	}
	if len(row) < 2 {
		return nil, fmt.Errorf("unexpected fund_basic response for %s", tsCode)
	}

	return &FundBasic{
		TsCode:     toString(row[0]),
		Name:       toString(row[1]),
		FundType:   toStrPtr(row, 2),
		Management: toStrPtr(row, 3),
	}, nil
}

// IndexBasic returns index info including name
func (c *Client) IndexBasic(tsCode string) (name string, err error) {
	params := map[string]interface{}{"ts_code": tsCode}
	fields := []string{"ts_code", "name"}
	resp, err := c.Query("index_basic", params, fields)
	if err != nil {
		return "", err
	}
	if resp.Data == nil || len(resp.Data.Items) == 0 {
		return "", fmt.Errorf("index %s not found", tsCode)
	}
	var row []interface{}
	if err := json.Unmarshal(resp.Data.Items[0], &row); err != nil {
		return "", err
	}
	if len(row) < 2 {
		return "", fmt.Errorf("unexpected index_basic response")
	}
	return toString(row[1]), nil
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func toStrPtr(row []interface{}, idx int) string {
	if idx >= len(row) {
		return ""
	}
	return toString(row[idx])
}

func toFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(val), 64)
		return f
	case int:
		return float64(val)
	}
	return 0
}

func ParseDate(s string) (time.Time, error) {
	if len(s) == 8 {
		return time.Parse("20060102", s)
	}
	return time.Parse("2006-01-02", s)
}
