package takers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

func (c *Client) authenticate(ctx context.Context, email, password string) (token string, e error) {
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

	respBody, err := c.doRequest(ctx, http.MethodPost, pathAuth, nil, bytes.NewBuffer(body))
	if err != nil {
		return "", errors.WithMessage(err, "failed to do request")
	}
	defer respBody.Close()

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

func (c *Client) Authenticate(ctx context.Context) error {
	token, err := c.authenticate(ctx, c.email, c.password)
	if err != nil {
		return err
	}

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	c.token = token

	return nil
}
