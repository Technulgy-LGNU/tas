package cloudflare

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateTransformOrigin accepts a Cloudflare-proxied HTTPS origin, not a path
// or a full transformation URL. Empty preserves hosted Images delivery.
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

// DeliveryURL requests real resizing/compression rather than a named "webp"
// variant. Hosted Images still negotiates the format using the Accept header.
// An explicit transformation origin instead requests WebP from /cdn-cgi/image.
func DeliveryURL(base, imageID, variant, transformOrigin string, width int) string {
	source := strings.TrimRight(base, "/") + "/" + url.PathEscape(imageID)
	options := fmt.Sprintf("width=%d,fit=scale-down,quality=80,format=webp", width)
	if transformOrigin != "" {
		return strings.TrimRight(transformOrigin, "/") + "/cdn-cgi/image/" + options + "/" + source + "/" + url.PathEscape(variant)
	}
	return source + "/" + options
}
