package connection

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sgladkov/harvester/internal/metrics"
)

type RestyClient struct {
	client *resty.Client
	server string
}

func gzipEncoder(c *resty.Client, req *resty.Request) error {
	m := req.Body.(*metrics.Metrics)
	if m == nil {
		// compress Metrics updates only
		return nil
	}
	req.SetHeader("Content-Encoding", "gzip")
	originalBody, err := json.Marshal(m)
	if err != nil {
		return err
	}
	var compressedBody bytes.Buffer
	w, err := gzip.NewWriterLevel(&compressedBody, gzip.BestSpeed)
	_, err = w.Write(originalBody)
	w.Close()
	req.SetBody(compressedBody)
	return nil
}

func NewRestyClient(server string) *RestyClient {
	result := RestyClient{}
	result.server = server
	result.client = resty.New()
	result.client.SetHeader("Content-Type", "application/json")
	result.client.OnBeforeRequest(gzipEncoder)
	return &result
}

func (c *RestyClient) UpdateMetrics(m *metrics.Metrics) error {
	_, err := c.client.R().
		SetBody(m).
		Post(fmt.Sprintf("%s/update/", c.server))
	return err
}
