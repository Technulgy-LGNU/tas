package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCloudflareUploadAndDelete(t *testing.T) {
	for _, status := range []int{200, 400, 401, 404, 429, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer private-token" {
					t.Error("missing authorization")
				}
				if r.Method == "POST" {
					if r.URL.Path != "/accounts/account/images/v1" {
						t.Error("wrong upload endpoint")
					}
					if err := r.ParseMultipartForm(1024); err != nil {
						t.Error(err)
					}
					if r.FormValue("id") != "image-id" || r.FormValue("creator") != "user-id" || r.FormValue("requireSignedURLs") != "false" {
						t.Error("wrong upload fields")
					}
					file, header, err := r.FormFile("file")
					if err != nil {
						t.Error(err)
						return
					}
					defer file.Close()
					data, _ := io.ReadAll(file)
					if string(data) != "image-bytes" || header.Filename != "image.png" {
						t.Error("file was not forwarded")
					}
				} else if r.Method != "DELETE" || r.URL.Path != "/accounts/account/images/v1/image-id" {
					t.Error("wrong delete endpoint")
				}
				w.WriteHeader(status)
				_, _ = fmt.Fprint(w, `{"success":true,"result":{"id":"image-id"}}`)
			}))
			defer server.Close()
			provider := NewImages("account", "private-token")
			provider.BaseURL = server.URL
			err := provider.Upload(context.Background(), "image-id", "image.png", "user-id", []byte("image-bytes"))
			if (err == nil) != (status == 200) {
				t.Fatalf("upload returned unexpected error: %v", err)
			}
			err = provider.Delete(context.Background(), "image-id")
			if (err == nil) != (status == 200 || status == 404) {
				t.Fatalf("delete returned unexpected error: %v", err)
			}
		})
	}
}

func TestCloudflareRejectsInvalidSuccessResponses(t *testing.T) {
	for _, body := range []string{`{"success":false}`, `{"success":true,"result":{"id":"wrong"}}`, `not-json`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = fmt.Fprint(w, body) }))
		provider := NewImages("account", "token")
		provider.BaseURL = server.URL
		if err := provider.Upload(context.Background(), "image-id", "image.png", "user", []byte("image")); err == nil {
			t.Fatal("invalid response accepted")
		}
		server.Close()
	}
}

func TestCloudflareErrorDetails(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"provider detail", 400, `{"success":false,"errors":[{"code":5455,"message":"Custom ID is invalid"}]}`, `code 5455: "Custom ID is invalid"`},
		{"redacted credential", 400, `{"success":false,"errors":[{"code":1000,"message":"Invalid token private-token\nTry again"}]}`, `code 1000: "Invalid token [redacted] Try again"`},
		{"non-JSON response", 502, `<html>private-token</html>`, "Cloudflare returned HTTP 502"},
		{"failure with HTTP 200", 200, `{"success":false,"errors":[{"code":1000,"message":"Rejected"}]}`, `code 1000: "Rejected"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); _, _ = fmt.Fprint(w, tc.body) }))
			defer server.Close()
			provider := NewImages("account", "private-token")
			provider.BaseURL = server.URL
			err := provider.Upload(context.Background(), "test-id", "image.png", "user", []byte("image"))
			if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "private-token") || strings.ContainsAny(err.Error(), "\r\n") {
				t.Fatalf("unsafe or missing error details: %v", err)
			}
		})
	}
}
