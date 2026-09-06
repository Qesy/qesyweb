package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/Qesy/QesyWeb/config"
)

type Mailer struct {
	cfg      config.ResendConfig
	client   *http.Client
	endpoint string
}
type emailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text,omitempty"`
}
type emailResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func New(cfg config.ResendConfig) *Mailer {
	return &Mailer{cfg: cfg, client: &http.Client{Timeout: 15 * time.Second}, endpoint: "https://api.resend.com/emails"}
}

func (m *Mailer) withEndpoint(endpoint string) *Mailer {
	clone := *m
	clone.endpoint = endpoint
	return &clone
}
func (m *Mailer) Send(ctx context.Context, to, subject, body string) error {
	if !m.cfg.Enabled {
		return fmt.Errorf("Resend 邮件发送未启用，请在 .env 中设置 RESEND_ENABLED=true")
	}
	if m.cfg.APIKey == "" || m.cfg.From == "" {
		return fmt.Errorf("RESEND_API_KEY 和 RESEND_FROM 不能为空")
	}
	payload, err := json.Marshal(emailRequest{From: m.cfg.From, To: []string{to}, Subject: subject, HTML: "<p>" + strings.ReplaceAll(html.EscapeString(body), "\n", "<br>") + "</p>", Text: body})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "qesyweb/1.0")
	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var detail emailResponse
		_ = json.NewDecoder(res.Body).Decode(&detail)
		return fmt.Errorf("Resend 返回 HTTP %d: %s", res.StatusCode, detail.Message)
	}
	return nil
}
