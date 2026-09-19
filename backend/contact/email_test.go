package contact

import (
	"strings"
	"testing"
)

func TestEmailConfigurationAndEncoding(t *testing.T) {
	cfg := Config{Host: "mail.example.org", Port: 587, Username: "user", TLSMode: "starttls", From: "noreply@example.org", To: "team@example.org"}
	if cfg.Ready() {
		t.Fatal("authenticated SMTP accepted without password")
	}
	cfg.Password = "secret"
	if !cfg.Ready() {
		t.Fatal("valid configuration rejected")
	}
	cfg.TLSMode = "none"
	if cfg.Ready() {
		t.Fatal("unencrypted SMTP accepted")
	}
	cfg.TLSMode = "starttls"
	m := Message{Name: "Visitor", Email: "visitor@example.org", Subject: "Grüße", Body: "First line\nSecond line", Language: "de"}
	wire := string(Encode(cfg, m))
	for _, part := range []string{"From: noreply@example.org\r\n", "To: team@example.org\r\n", "Reply-To: visitor@example.org\r\n", "Content-Type: text/plain; charset=UTF-8", "First line\r\nSecond line"} {
		if !strings.Contains(wire, part) {
			t.Fatalf("missing email part: %s", part)
		}
	}
	if strings.Contains(wire, "secret") {
		t.Fatal("SMTP secret leaked")
	}
	for _, address := range []string{"bad", "a@example.org\r\nBcc: b@example.org", "Name <a@example.org>"} {
		if ValidAddress(address) {
			t.Fatalf("invalid address accepted: %q", address)
		}
	}
}
