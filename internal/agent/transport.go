package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/Happy-skills/metrics/internal/compress"
	"github.com/Happy-skills/metrics/internal/logger"
	"gopkg.in/h2non/gentleman.v2"
)

func sendWithRetry(url, headerValue string, body []byte) error {
	const maxRetries = 3
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = sendUpdateRequestWithCompress(url, headerValue, body)
		if err == nil {
			return nil
		}

		if classify(err) == NonRetriable {
			logger.Sugar.Infof("error sending metrics: %s", err.Error())
			return err
		}

		time.Sleep(time.Duration(attempt+(attempt-1)) * time.Second)
	}

	return fmt.Errorf("error sending metrics after %d attempts: %w", maxRetries, err)
}

func sendUpdateRequestWithCompress(url string, headerValue string, body []byte) error {
	client := gentleman.New()
	req := client.Request()
	req.Method(http.MethodPost)
	req.URL(url)
	req.SetHeader("Content-Type", headerValue)

	contentType := slices.Contains(compress.TypesForGzip, headerValue)
	if contentType {
		var buf bytes.Buffer
		defer buf.Reset()

		g := gzip.NewWriter(&buf)
		if _, err := g.Write(body); err != nil {
			logger.Sugar.Errorf("Agent error gzip write: %s", err.Error())
			return err
		}
		if err := g.Close(); err != nil {
			logger.Sugar.Errorf("Agent error gzip close: %s", err.Error())
			return err
		}
		body = buf.Bytes()
		req.SetHeader("Content-Encoding", "gzip")
		req.SetHeader("Accept-Encoding", "gzip")
	}

	req.Body(bytes.NewReader(body))

	response, err := req.Send()
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	defer func(response *gentleman.Response) {
		err := response.Close()
		if err != nil {
			logger.Log.Error(err.Error())
		}
	}(response)

	if !response.Ok {
		return fmt.Errorf("agent error Send: %d %s", response.StatusCode, response.String())
	}

	return nil
}
