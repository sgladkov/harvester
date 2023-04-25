package connection

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sgladkov/harvester/internal/interfaces"
)

type RestyClient struct {
	client *resty.Client
	server string
}

func gzipEncoder(_ *resty.Client, req *resty.Request) error {
	m := req.Body.(*interfaces.Metrics)
	if m == nil {
		// compress Metrics updates only
		return nil
	}
	req.SetHeader("Content-Encoding", "gzip")
	req.SetHeader("Accept-Encoding", "gzip")
	originalBody, err := json.Marshal(m)
	if err != nil {
		return err
	}
	var compressedBody bytes.Buffer
	w, err := gzip.NewWriterLevel(&compressedBody, gzip.BestSpeed)
	if err != nil {
		return err
	}
	_, err = w.Write(originalBody)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
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

func (c *RestyClient) UpdateMetrics(m *interfaces.Metrics) error {
	_, err := c.client.R().
		SetBody(m).
		Post(fmt.Sprintf("%s/update/", c.server))
	return err
}
