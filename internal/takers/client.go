package takers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

const (
	pathAuth = "/auth/authenticate"
)

type Client struct {
	baseClient *http.Client
	baseURL    *url.URL
}

func NewClient(baseClient *http.Client, baseURLString string) (*Client, error) {
	if baseClient == nil {
		baseClient = &http.Client{}
	}

	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse base URL")
	}

	return &Client{
		baseClient: baseClient,
		baseURL:    baseURL,
	}, nil
}

func (c *Client) doRequest(ctx context.Context, method string, pth string, body io.Reader) (io.ReadCloser, error) {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(endpoint.Path, pth)

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	resp, err := c.baseClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()

		return nil, errors.Errorf("API responded with code %v", resp.StatusCode)
	}

	return resp.Body, nil
}

func (c *Client) Authenticate(ctx context.Context, email, password string) (token string, e error) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	body, err := json.Marshal(request{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal body")
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, pathAuth, bytes.NewBuffer(body))
	if err != nil {
		return "", errors.WithMessage(err, "failed to do request")
	}
	defer func() {
		io.Copy(ioutil.Discard, respBody)
		respBody.Close()
	}()

	type response struct {
		Token string `json:"access_token"`
	}

	var resp response

	err = json.NewDecoder(respBody).Decode(&resp)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode response")
	}

	return resp.Token, nil
}
