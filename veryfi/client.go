package veryfi

import (
	"fmt"
	"github.com/pkg/errors"

	"github.com/go-resty/resty/v2"
	"github.com/hoanhan101/veryfi-go/veryfi/scheme"
)

// httpClient implements a Veryfi API Client.
type httpClient struct {
	// options is the global config options of the client.
	options *Options

	// client holds the resty.Client.
	client *resty.Client

	// apiVersion is the current API version of Veryfi that we are
	// communicating with.
	apiVersion string
}

// NewClientV7 returns a new instance of a client for v7 API.
func NewClientV7(opts *Options) (Client, error) {
	c, err := createClient(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a client")
	}

	return &httpClient{
		options:    opts,
		client:     c,
		apiVersion: "v7",
	}, nil
}

// createClient setups a resty client with configured options.
func createClient(opts *Options) (*resty.Client, error) {
	err := setDefaults(opts)
	if err != nil {
		return nil, err
	}

	// Create a resty client with configured options.
	client := resty.New()
	client = client.
		SetTimeout(opts.HTTP.Timeout).
		SetRetryCount(int(opts.HTTP.Retry.Count)).
		SetRetryWaitTime(opts.HTTP.Retry.WaitTime).
		SetRetryMaxWaitTime(opts.HTTP.Retry.MaxWaitTime)

	return client, nil
}

// Config returns the client configuration options.
func (c *httpClient) Config() *Options {
	return c.options
}

// ProcessDocumentURL returns the processed document using URL.
func (c *httpClient) ProcessDocumentURL(opts scheme.DocumentURLOptions) (*scheme.Document, error) {
	out := new(*scheme.Document)
	if err := c.post(documentURI, opts, out); err != nil {
		return nil, err
	}

	return *out, nil
}

// get performs a GET request against Veryfi API.
// func (c *httpClient) get(uri string, params map[string]string, okScheme interface{}) error {
// 	errScheme := new(scheme.Error)
// 	_, err := c.setBaseURL().R().SetQueryParams(params).SetResult(okScheme).SetError(errScheme).Get(uri)
// 	return check(err, errScheme)

// }

// post performs a POST request against Veryfi API.
func (c *httpClient) post(uri string, body interface{}, okScheme interface{}) error {
	errScheme := new(scheme.Error)
	_, err := c.setBaseURL().R().
		SetBody(body).
		SetHeaders(map[string]string{
			"Content-Type":  "application/json",
			"Accept":        "application/json",
			"CLIENT-ID":     c.options.ClientID,
			"AUTHORIZATION": fmt.Sprintf("apikey %s:%s", c.options.Username, c.options.APIKey),
		}).
		SetResult(okScheme).
		SetError(errScheme).
		Post(uri)

	return check(err, errScheme)
}

// setBaseURL returns a client that uses Veryfi's base URL.
func (c *httpClient) setBaseURL() *resty.Client {
	return c.client.SetHostURL(buildURL(c.options.EnvironmentURL, "api", c.apiVersion))
}

// check validates returned response from Veryfi.
func check(err error, errResp *scheme.Error) error {
	if err != nil {
		return errors.Wrap(err, "failed to make a request to Veryfi")
	}

	if *errResp != (scheme.Error{}) {
		return errors.Errorf(
			"got a %v error response from Veryfi at %v, saying %v, with context: %v",
			errResp.HTTPCode, errResp.Timestamp, errResp.Description, errResp.Context,
		)
	}

	return nil
}