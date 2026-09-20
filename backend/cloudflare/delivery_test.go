package cloudflare

import (
	"strings"
	"testing"
)

func TestDeliveryURLs(t *testing.T) {
	for _, tc := range []struct {
		name, base, id, origin, want string
	}{
		{"hosted", "https://imagedelivery.net/hash/", "tas-photo", "", "https://imagedelivery.net/hash/tas-photo/width=1920,fit=scale-down,quality=80,format=webp"},
		{"custom hosted delivery", "https://imagedelivery.net/hash/", "tas-photo", "https://images.example.org/", "https://images.example.org/cdn-cgi/imagedelivery/hash/tas-photo/width=1920,fit=scale-down,quality=80,format=webp"},
		{"custom delivery and escaped ID", "https://images.example.org/cdn-cgi/imagedelivery/hash", "folder/photo?x#y", "https://www.example.org", "https://www.example.org/cdn-cgi/imagedelivery/hash/folder%2Fphoto%3Fx%23y/width=1920,fit=scale-down,quality=80,format=webp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := DeliveryURL(tc.base, tc.id, tc.origin, 1920); strings.Contains(got, "/cdn-cgi/image/") || strings.Count(got, "https://") != 1 {
				t.Fatalf("nested remote transformation generated: %s", got)
			}
			if got := DeliveryURL(tc.base, tc.id, tc.origin, 1920); got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestTransformOriginValidation(t *testing.T) {
	for _, origin := range []string{"", "https://images.example.org", "https://images.example.org/"} {
		if err := ValidateTransformOrigin(origin); err != nil {
			t.Errorf("valid origin rejected: %s", origin)
		}
	}
	for _, origin := range []string{"http://images.example.org", "https://", "//images.example.org", "https://user:secret@images.example.org", "https://images.example.org/cdn-cgi/image", "https://images.example.org?x=1", "https://images.example.org?", "https://images.example.org#fragment"} {
		if ValidateTransformOrigin(origin) == nil {
			t.Errorf("invalid origin accepted: %s", origin)
		}
	}
}
