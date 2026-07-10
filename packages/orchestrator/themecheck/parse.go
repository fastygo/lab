package themecheck

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fastygo/lab/packages/domain"
)

// Message is one row from `wp theme-check run --format=json`.
type Message struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ParseJSON converts theme-check CLI JSON into Lab findings.
func ParseJSON(raw []byte, gate, check, target, slug string) ([]domain.Finding, error) {
	var rows []Message
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("theme-check json: %w", err)
	}
	var findings []domain.Finding
	required := 0
	warning := 0
	other := 0
	for _, row := range rows {
		if strings.TrimSpace(row.Value) == "" {
			continue
		}
		typeU := strings.ToUpper(strings.TrimSpace(row.Type))
		sev := domain.SeverityInfo
		code := "org.themecheck.info"
		switch {
		case strings.Contains(typeU, "REQUIRED") || typeU == "ERROR":
			sev = domain.SeverityHigh
			code = "org.themecheck.required"
			required++
		case strings.Contains(typeU, "WARNING"):
			sev = domain.SeverityMedium
			code = "org.themecheck.warning"
			warning++
		case strings.Contains(typeU, "RECOMMENDED") || strings.Contains(typeU, "INFO"):
			sev = domain.SeverityLow
			code = "org.themecheck.recommended"
			other++
		default:
			other++
		}
		msg := row.Value
		if typeU != "" {
			msg = "[" + typeU + "] " + row.Value
		}
		findings = append(findings, domain.Finding{
			Code:     code,
			Gate:     gate,
			Check:    check,
			Severity: sev,
			Message:  msg,
			Target:   target,
			Evidence: map[string]string{"type": typeU, "theme": slug},
		})
	}
	if len(findings) == 0 {
		return []domain.Finding{{
			Code:     "org.themecheck.ok",
			Gate:     gate,
			Check:    check,
			Severity: domain.SeverityInfo,
			Message:  "Theme Check returned no messages",
			Target:   target,
			Evidence: map[string]string{"theme": slug},
		}}, nil
	}
	if required == 0 {
		summary := domain.Finding{
			Code:     "org.themecheck.no_required",
			Gate:     gate,
			Check:    check,
			Severity: domain.SeverityInfo,
			Message:  fmt.Sprintf("Theme Check: 0 required, %d warning, %d other", warning, other),
			Target:   target,
			Evidence: map[string]string{
				"theme":    slug,
				"required": "0",
				"warning":  fmt.Sprintf("%d", warning),
			},
		}
		findings = append([]domain.Finding{summary}, findings...)
	}
	return findings, nil
}
