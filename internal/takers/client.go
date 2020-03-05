package takers

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/pkg/errors"
)

const (
	pathAuth   = "/auth/authenticate"
	pathTakers = "/test-takers"
)

type Client struct {
	baseClient *http.Client
	baseURL    *url.URL

	token   string
	tokenMu *sync.RWMutex

	email, password string
}

type CodeError struct {
	StatusCode int
	Body       []byte
}

func (c CodeError) Error() string {
	return fmt.Sprintf("API returned code %v", c.StatusCode)
}

func NewClient(baseClient *http.Client, baseURLString string, email, password string) (*Client, error) {
	if baseClient == nil {
		baseClient = &http.Client{}
	}

	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse base URL")
	}

	client := &Client{
		baseClient: baseClient,
		baseURL:    baseURL,
		tokenMu:    &sync.RWMutex{},
		email:      email,
		password:   password,
	}

	return client, nil
}

func (c *Client) signRequest(req *http.Request) {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()

	if c.token == "" {
		return
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
}

func (c *Client) doRequest(ctx context.Context, method string, pth string, query url.Values, body io.Reader) (io.ReadCloser, error) {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(endpoint.Path, pth)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	c.signRequest(req)

	log.Printf("%s %s %s", method, pth, query.Encode())

	resp, err := c.baseClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	if resp.StatusCode != http.StatusOK {
		bts, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		return nil, CodeError{
			StatusCode: resp.StatusCode,
			Body:       bts,
		}
	}

	return resp.Body, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values) (io.ReadCloser, error) {
	return c.doRequest(ctx, http.MethodGet, path, query, nil)
}
