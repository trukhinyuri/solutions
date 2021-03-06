package clients

import (
	"context"

	"errors"

	"time"

	"github.com/go-resty/resty"
	"github.com/sirupsen/logrus"
)

// DownloadClient is an interface to resource-service.
type DownloadClient interface {
	DownloadFile(ctx context.Context, url string) ([]byte, error)
}

type httpDownloadClient struct {
	rest *resty.Client
	log  *logrus.Entry
}

// NewHTTPDownloadClient returns client for resource-service working via restful api
func NewHTTPDownloadClient(debug bool) DownloadClient {
	log := logrus.WithField("component", "download_client")
	client := resty.New().
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(debug).
		SetTimeout(3 * time.Second)
	return &httpDownloadClient{
		rest: client,
		log:  log,
	}
}

func (c *httpDownloadClient) DownloadFile(ctx context.Context, url string) ([]byte, error) {
	c.log.WithField("URL", url).Infoln("Downloading file")

	resp, err := c.rest.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() > 399 {
		return nil, errors.New("unable to download file")
	}
	return resp.Body(), nil
}
