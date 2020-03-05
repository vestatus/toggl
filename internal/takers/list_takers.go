package takers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"toggl/internal/service"

	"github.com/pkg/errors"
)

// God bless https://mholt.github.io/json-to-go/
type taker struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	URL             string  `json:"url"`
	HireState       string  `json:"hire_state"`
	SubmittedInTime bool    `json:"submitted_in_time"`
	IsDemo          bool    `json:"is_demo"`
	Percent         float64 `json:"percent"`
	Points          float64 `json:"points"`
	StartedAt       int     `json:"started_at"`
	FinishedAt      int     `json:"finished_at"`
	ContactInfo     struct {
		Phone        string `json:"phone"`
		FullName     string `json:"full_name"`
		Street       string `json:"street"`
		City         string `json:"city"`
		ZipCode      string `json:"zip_code"`
		State        string `json:"state"`
		Country      string `json:"country"`
		Website      string `json:"website"`
		Linkedin     string `json:"linkedin"`
		ContactEmail string `json:"contact_email"`
	} `json:"contact_info"`
	TestDurationInSeconds int `json:"test_duration_in_seconds"`
}

func (c *Client) nextTakers(ctx context.Context, offset, limit int) (takers []taker, total int, e error) {
	query := url.Values{}
	query.Set("offset", strconv.Itoa(offset))
	query.Set("limit", strconv.Itoa(limit))

	body, err := c.get(ctx, pathTakers, query)
	// one retry in case the token has expired
	if errCode, ok := err.(CodeError); ok && errCode.StatusCode == http.StatusUnauthorized {
		err = c.Authenticate(ctx)
		if err != nil {
			return nil, 0, errors.WithMessage(err, "failed to re-authenticate after a 401")
		}
		body, err = c.get(ctx, pathTakers, query)
	}
	if err != nil {
		return nil, 0, errors.WithMessage(err, "request failed")
	}
	defer body.Close()

	type response struct {
		Takers []taker `json:"test_takers"`
		Total  int     `json:"total"`
	}

	var resp response

	err = json.NewDecoder(body).Decode(&resp)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to unmarshal takers")
	}

	return resp.Takers, resp.Total, nil
}

func (t *taker) toDomain() service.Taker {
	email := t.ContactInfo.ContactEmail
	if email == "" {
		email = t.Email
	}

	name := t.ContactInfo.FullName
	if name == "" {
		name = t.Name
	}

	return service.Taker{
		ID:      t.ID,
		Name:    name,
		Email:   email,
		Points:  t.Points,
		Percent: t.Percent,
		Demo:    t.IsDemo,
	}
}

func (c *Client) ListTakers(ctx context.Context) ([]service.Taker, error) {
	const pageSize = 10

	total := pageSize

	allTakers := make([]service.Taker, 0, pageSize)

	// in case new takers are added
	idSet := map[int]struct{}{}

	for offset := 0; offset < total; offset += pageSize {
		takers, newTotal, err := c.nextTakers(ctx, offset, pageSize)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to load more takers")
		}

		for i := range takers {
			id := takers[i].ID

			if _, found := idSet[id]; found {
				continue
			}

			idSet[id] = struct{}{}
			allTakers = append(allTakers, takers[i].toDomain())
		}

		total = newTotal
	}

	return allTakers, nil
}
