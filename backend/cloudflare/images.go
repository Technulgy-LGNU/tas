package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Images struct {
	AccountID, Token string
	BaseURL          string
	Client           *http.Client
}

func NewImages(accountID, token string) *Images {
	return &Images{AccountID: accountID, Token: token, BaseURL: "https://api.cloudflare.com/client/v4",
		Client: &http.Client{Timeout: 60 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (s *Images) endpoint(id string) string {
	endpoint := strings.TrimRight(s.BaseURL, "/") + "/accounts/" + url.PathEscape(s.AccountID) + "/images/v1"
	if id != "" {
		endpoint += "/" + url.PathEscape(id)
	}
	return endpoint
}

func (s *Images) Upload(ctx context.Context, id, filename, creator string, data []byte) error {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	// Supplying our own ID makes even a timed-out upload identifiable and deletable.
	for key, value := range map[string]string{"id": id, "creator": creator, "requireSignedURLs": "false"} {
		if err := form.WriteField(key, value); err != nil {
			return err
		}
	}
	file, err := form.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		return err
	}
	if err = form.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint(""), &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", form.FormDataContentType())
	result, err := s.do(req)
	if err != nil {
		return err
	}
	if result.ID != id {
		return errors.New("Cloudflare returned an unexpected image ID")
	}
	return nil
}

func (s *Images) Delete(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.endpoint(id), nil)
	if err != nil {
		return err
	}
	_, err = s.do(req)
	return err
}

type imageResult struct {
	ID string `json:"id"`
}

func (s *Images) do(req *http.Request) (imageResult, error) {
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Header.Set("Accept", "application/json")
	res, err := s.Client.Do(req)
	if err != nil {
		return imageResult{}, errors.New("Cloudflare request failed")
	}
	defer res.Body.Close()
	// Allows retry after the remote deletion succeeded but the database deletion failed.
	if req.Method == http.MethodDelete && res.StatusCode == http.StatusNotFound {
		return imageResult{}, nil
	}
	var envelope struct {
		Success bool        `json:"success"`
		Result  imageResult `json:"result"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	decodeErr := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&envelope)
	if res.StatusCode < 200 || res.StatusCode >= 300 || (decodeErr == nil && !envelope.Success) {
		message := fmt.Sprintf("Cloudflare returned HTTP %d", res.StatusCode)
		if decodeErr == nil {
			for i, detail := range envelope.Errors {
				if i == 3 {
					break
				}
				// Never log raw response bodies or credentials echoed by a provider.
				clean := detail.Message
				if s.Token != "" {
					clean = strings.ReplaceAll(clean, s.Token, "[redacted]")
				}
				clean = strings.Join(strings.Fields(clean), " ")
				if len(clean) > 500 {
					clean = clean[:500] + "…"
				}
				message += fmt.Sprintf("; code %d: %q", detail.Code, clean)
			}
		}
		return imageResult{}, errors.New(message)
	}
	if decodeErr != nil {
		return imageResult{}, errors.New("invalid Cloudflare response")
	}
	return envelope.Result, nil
}
