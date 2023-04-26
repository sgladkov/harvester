package connection

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sgladkov/harvester/internal/interfaces"
	"github.com/sgladkov/harvester/internal/logger"
	"go.uber.org/zap"
)

type RestyClient struct {
	client *resty.Client
	server string
}

func gzipEncoder(_ *resty.Client, req *resty.Request) error {
	logger.Log.Info("gzipEncoder")
	m := req.Body.(*interfaces.Metrics)
	if m == nil {
		// compress Metrics updates only
		logger.Log.Warn("Failed to get metrics from request body")
		return nil
	}
	logger.Log.Info("Got Metrics from request body")
	req.SetHeader("Content-Encoding", "gzip")
	req.SetHeader("Accept-Encoding", "gzip")
	req.SetHeader("Content-Type", "application/json")
	originalBody, err := json.Marshal(m)
	if err != nil {
		logger.Log.Warn("Failed to marshall Metrics")
		return err
	}
	logger.Log.Info("Marshalled", zap.String("body", string(originalBody)))
	var compressedBody bytes.Buffer
	w, err := gzip.NewWriterLevel(&compressedBody, gzip.BestCompression)
	if err != nil {
		logger.Log.Warn("Failed to compress JSON")
		return err
	}
	_, err = w.Write(originalBody)
	if err != nil {
		logger.Log.Warn("Failed to write to writer")
		return err
	}
	err = w.Close()
	if err != nil {
		logger.Log.Warn("Failed to close writer")
		return err
	}
	logger.Log.Info("Compressed", zap.Int("length", compressedBody.Len()),
		zap.String("body", compressedBody.String()))
	req.SetBody(compressedBody.String())
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
	reply, err := c.client.R().
		SetBody(m).
		Post(fmt.Sprintf("%s/update/", c.server))
	if err != nil {
		return err
	}
	if reply.IsError() {
		return errors.New(fmt.Sprintf("Failed to report metrics, status code is %d,  reply is [%s]",
			reply.StatusCode(), string(reply.Body())))
	}
	logger.Log.Info("Reply",
		zap.String("body", string(reply.Body())),
		zap.Int("status_code", reply.StatusCode()))
	return nil
}
