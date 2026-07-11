package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/runstore"
)

// Filter controls when to send (F4.5).
type Filter string

const (
	FilterAlways   Filter = "always"
	FilterFail     Filter = "fail"
	FilterWarnFail Filter = "warn+fail"
)

// Config from env.
type Config struct {
	SlackWebhook   string
	TelegramToken  string
	TelegramChatID string
	Filter         Filter
	DashboardBase  string // optional link prefix
}

// FromEnv reads notify settings.
func FromEnv() Config {
	f := Filter(strings.TrimSpace(os.Getenv("LAB_NOTIFY_ON")))
	if f == "" {
		f = FilterWarnFail
	}
	return Config{
		SlackWebhook:   strings.TrimSpace(os.Getenv("SLACK_WEBHOOK_URL")),
		TelegramToken:  strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		TelegramChatID: strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID")),
		Filter:         f,
		DashboardBase:  strings.TrimSpace(os.Getenv("LAB_DASHBOARD_URL")),
	}
}

func (c Config) Enabled() bool {
	return c.SlackWebhook != "" || (c.TelegramToken != "" && c.TelegramChatID != "")
}

// ShouldSend applies Filter to run status.
func (c Config) ShouldSend(status runstore.RunStatus) bool {
	switch c.Filter {
	case FilterAlways:
		return true
	case FilterFail:
		return status == runstore.StatusFail || status == runstore.StatusError
	default: // warn+fail
		return status == runstore.StatusFail || status == runstore.StatusError || status == runstore.StatusWarn
	}
}

// Message formats a short notify body.
func Message(run *runstore.Run, dashboardBase string) string {
	sum := domain.Summary{}
	if run.Report != nil {
		sum = run.Report.Summary
	}
	line := fmt.Sprintf("[%s] %s — high:%d medium:%d (%d findings)",
		run.Lab, strings.ToUpper(string(run.Status)), sum.High, sum.Medium, sum.Total)
	if run.Error != "" {
		line += "\nerror: " + run.Error
	}
	if dashboardBase != "" {
		base := strings.TrimRight(dashboardBase, "/")
		line += "\n→ " + base + "/runs/" + run.ID
	}
	return line
}

// NotifyRunFinished sends Slack and/or Telegram if configured.
func NotifyRunFinished(ctx context.Context, cfg Config, run *runstore.Run) error {
	if run == nil || !cfg.Enabled() || !cfg.ShouldSend(run.Status) {
		return nil
	}
	text := Message(run, cfg.DashboardBase)
	var errs []string
	if cfg.SlackWebhook != "" {
		if err := sendSlack(ctx, cfg.SlackWebhook, text); err != nil {
			errs = append(errs, "slack: "+err.Error())
		}
	}
	if cfg.TelegramToken != "" && cfg.TelegramChatID != "" {
		if err := sendTelegram(ctx, cfg.TelegramToken, cfg.TelegramChatID, text); err != nil {
			errs = append(errs, "telegram: "+err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

func sendSlack(ctx context.Context, webhook, text string) error {
	body, _ := json.Marshal(map[string]string{"text": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %s", resp.Status)
	}
	return nil
}

func sendTelegram(ctx context.Context, token, chatID, text string) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	body, _ := json.Marshal(map[string]string{
		"chat_id": chatID,
		"text":    text,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %s", resp.Status)
	}
	return nil
}
