package themecheck

import (
	"testing"

	"github.com/fastygo/lab/packages/domain"
)

func TestParseJSONRequired(t *testing.T) {
	t.Parallel()
	raw := []byte(`[
		{"type":"REQUIRED","value":"Missing something"},
		{"type":"WARNING","value":"Be careful"},
		{"type":"RECOMMENDED","value":"Nice to have"}
	]`)
	got, err := ParseJSON(raw, "C2", "theme-check", "http://wp.test", "latte")
	if err != nil {
		t.Fatal(err)
	}
	var req, warn int
	for _, f := range got {
		switch f.Code {
		case "org.themecheck.required":
			req++
			if f.Severity != domain.SeverityHigh {
				t.Fatalf("severity=%s", f.Severity)
			}
		case "org.themecheck.warning":
			warn++
		}
	}
	if req != 1 || warn != 1 {
		t.Fatalf("req=%d warn=%d findings=%+v", req, warn, got)
	}
}

func TestParseJSONNoRequiredSummary(t *testing.T) {
	t.Parallel()
	raw := []byte(`[{"type":"WARNING","value":"only warn"}]`)
	got, err := ParseJSON(raw, "C2", "tc", "http://wp", "latte")
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Code != "org.themecheck.no_required" {
		t.Fatalf("%+v", got[0])
	}
}
