package cloudflare

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateTransformOrigin accepts a Cloudflare-proxied HTTPS origin in the same
// account as Images. Empty preserves delivery through the configured base URL.
func ValidateTransformOrigin(origin string) error {
	if origin == "" {
		return nil
	}
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return fmt.Errorf("cloudflare.images_transform_origin must be an HTTPS origin without credentials, path, query or fragment")
	}
	return nil
}

// DeliveryURL applies flexible variants directly to hosted Images. Do not wrap
// imagedelivery.net in /cdn-cgi/image: that chains separate processing services
// and can fail with error 9524. The custom-domain hosted Images endpoint uses
// the same account hash and optimizations as imagedelivery.net.
// Hosted Images negotiates the output format using the request's Accept header.
func DeliveryURL(base, imageID, transformOrigin string, width int) string {
	base = strings.TrimRight(base, "/")
	options := fmt.Sprintf("width=%d,fit=scale-down,quality=80,format=webp", width)
	if transformOrigin != "" {
		// Both supported bases end in the account hash:
		// imagedelivery.net/<hash> or <domain>/cdn-cgi/imagedelivery/<hash>.
		hash := base[strings.LastIndex(base, "/")+1:]
		base = strings.TrimRight(transformOrigin, "/") + "/cdn-cgi/imagedelivery/" + hash
	}
	return base + "/" + url.PathEscape(imageID) + "/" + options
}
