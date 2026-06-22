package tushare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.tushare.pro"

type Client struct {
	token   string
	baseURL string
	hc      *http.Client
}

type request struct {
	APIName string                 `json:"api_name"`
	Token   string                 `json:"token"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Fields  []string               `json:"fields,omitempty"`
}

type Response struct {
	Code    int        `json:"code"`
	Message string     `json:"msg"`
	Data    *DataBlock `json:"data"`
}

type DataBlock struct {
	Fields []string     `json:"fields"`
	Items  []json.RawMessage `json:"items"`
	HasMore bool        `json:"has_more"`
}

func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: defaultBaseURL,
		hc:      &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Query(apiName string, params map[string]interface{}, fields []string) (*Response, error) {
	body := request{
		APIName: apiName,
		Token:   c.token,
		Params:  params,
		Fields:  fields,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("tushare marshal request: %w", err)
	}

	resp, err := c.hc.Post(c.baseURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("tushare post %s: %w", apiName, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tushare read body: %w", err)
	}

	var result Response
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("tushare decode response: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("tushare api error (code=%d): %s", result.Code, result.Message)
	}
	if result.Data == nil || len(result.Data.Items) == 0 {
		return &result, nil
	}

	return &result, nil
}

// UnmarshalItem unmarshals a single item row into a struct using field order.
// The target struct must have tushare tags matching the field names.
func UnmarshalItem(item json.RawMessage, fields []string, target interface{}) error {
	var row []interface{}
	if err := json.Unmarshal(item, &row); err != nil {
		return err
	}
	if len(row) != len(fields) {
		return fmt.Errorf("tushare field count mismatch: %d items vs %d fields", len(row), len(fields))
	}
	return nil
}
