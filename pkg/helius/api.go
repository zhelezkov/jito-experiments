package helius

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/valyala/fastjson"
)

const MaxBatchSize = 1000

type HeliusClient struct {
	Url    string
	parser fastjson.Parser
}

func (c *HeliusClient) GetAssetsBatch(ids []string) (*fastjson.Value, error) {
	if len(ids) > MaxBatchSize {
		return nil, errors.New("batch size is too big")
	}

	body, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "GetAssetsBatch_" + ids[0],
		"method":  "getAssetBatch",
		"params": map[string]interface{}{
			"ids": ids,
		},
	})

	resp, err := http.Post(c.Url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return c.parser.ParseBytes(bodyResp)
}
