package policy

import (
	"testing"

	"github.com/fastygo/lab/packages/domain"
)

func TestEngineMapDefault(t *testing.T) {
	t.Parallel()
	e := NewEngine("default")
	findings := []domain.Finding{
		{Code: "demo.stub.ok", Severity: domain.SeverityInfo},
		{Code: "demo.stub.ok", Severity: domain.SeverityInfo}, // dedupe
		{Code: "unknown.code", Severity: domain.SeverityLow},
	}
	decisions := e.Map(findings)
	if len(decisions) != 2 {
		t.Fatalf("len(decisions)=%d, want 2", len(decisions))
	}
	byCode := map[string]domain.Decision{}
	for _, d := range decisions {
		byCode[d.FindingCode] = d
	}
	if byCode["demo.stub.ok"].Basket != domain.BasketAccept {
		t.Fatalf("demo.stub.ok basket = %s", byCode["demo.stub.ok"].Basket)
	}
	if byCode["unknown.code"].Basket != domain.BasketAccept {
		t.Fatalf("unknown default basket = %s", byCode["unknown.code"].Basket)
	}
}

func TestEngineMapLightspeed(t *testing.T) {
	t.Parallel()
	e := NewEngine("lightspeed")
	decisions := e.Map([]domain.Finding{
		{Code: "quality.axe.color-contrast", Severity: domain.SeverityHigh},
		{Code: "runner.docker.unavailable", Severity: domain.SeverityHigh},
	})
	by := map[string]domain.Decision{}
	for _, d := range decisions {
		by[d.FindingCode] = d
	}
	if by["quality.axe.color-contrast"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["quality.axe.color-contrast"])
	}
	if by["runner.docker.unavailable"].Basket != domain.BasketAccept {
		t.Fatalf("%+v", by["runner.docker.unavailable"])
	}
}

func TestEngineMapLightspeedMotionLinksCWV(t *testing.T) {
	t.Parallel()
	e := NewEngine("lightspeed")
	decisions := e.Map([]domain.Finding{
		{Code: "quality.motion.ok", Severity: domain.SeverityInfo},
		{Code: "quality.motion.unreduced", Severity: domain.SeverityHigh},
		{Code: "quality.links.broken", Severity: domain.SeverityHigh},
		{Code: "quality.links.ok", Severity: domain.SeverityInfo},
		{Code: "quality.lighthouse.lcp", Severity: domain.SeverityInfo},
		{Code: "quality.lighthouse.exec_failed", Severity: domain.SeverityHigh},
	})
	by := map[string]domain.Decision{}
	for _, d := range decisions {
		by[d.FindingCode] = d
	}
	if by["quality.motion.ok"].Basket != domain.BasketAccept {
		t.Fatalf("%+v", by["quality.motion.ok"])
	}
	if by["quality.motion.unreduced"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["quality.motion.unreduced"])
	}
	if by["quality.links.broken"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["quality.links.broken"])
	}
	if by["quality.lighthouse.lcp"].Basket != domain.BasketBudget {
		t.Fatalf("%+v", by["quality.lighthouse.lcp"])
	}
	if by["quality.lighthouse.exec_failed"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["quality.lighthouse.exec_failed"])
	}
}

func TestEngineMapLightspeedQ356(t *testing.T) {
	t.Parallel()
	e := NewEngine("lightspeed")
	decisions := e.Map([]domain.Finding{
		{Code: "quality.css.ok", Severity: domain.SeverityInfo},
		{Code: "quality.css.parse_error", Severity: domain.SeverityHigh},
		{Code: "quality.seo.ok", Severity: domain.SeverityInfo},
		{Code: "quality.seo.title_missing", Severity: domain.SeverityHigh},
		{Code: "quality.seo.h1", Severity: domain.SeverityMedium},
		{Code: "quality.viewport.failed", Severity: domain.SeverityHigh},
		{Code: "quality.console.ok", Severity: domain.SeverityInfo},
	})
	by := map[string]domain.Decision{}
	for _, d := range decisions {
		by[d.FindingCode] = d
	}
	if by["quality.css.ok"].Basket != domain.BasketAccept {
		t.Fatalf("%+v", by["quality.css.ok"])
	}
	if by["quality.css.parse_error"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["quality.css.parse_error"])
	}
	if by["quality.seo.title_missing"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["quality.seo.title_missing"])
	}
	if by["quality.seo.h1"].Basket != domain.BasketBudget {
		t.Fatalf("%+v", by["quality.seo.h1"])
	}
	if by["quality.viewport.failed"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["quality.viewport.failed"])
	}
}


func TestEngineMapWordpressOrgGate23(t *testing.T) {
	t.Parallel()
	e := NewEngine("wordpress-org")
	decisions := e.Map([]domain.Finding{
		{Code: "org.themecheck.required", Severity: domain.SeverityHigh},
		{Code: "org.matrix.status_5xx", Severity: domain.SeverityHigh},
		{Code: "org.matrix.ok", Severity: domain.SeverityInfo},
		{Code: "org.themecheck.no_required", Severity: domain.SeverityInfo},
	})
	by := map[string]domain.Decision{}
	for _, d := range decisions {
		by[d.FindingCode] = d
	}
	if by["org.themecheck.required"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["org.themecheck.required"])
	}
	if by["org.matrix.status_5xx"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["org.matrix.status_5xx"])
	}
	if by["org.matrix.ok"].Basket != domain.BasketAccept {
		t.Fatalf("%+v", by["org.matrix.ok"])
	}
	if by["org.themecheck.no_required"].Basket != domain.BasketAccept {
		t.Fatalf("%+v", by["org.themecheck.no_required"])
	}
}

func TestEngineMapWordpressOrgGate4Keyboard(t *testing.T) {
	t.Parallel()
	e := NewEngine("wordpress-org")
	decisions := e.Map([]domain.Finding{
		{Code: "org.keyboard.ok", Severity: domain.SeverityInfo},
		{Code: "org.keyboard.skip_ok", Severity: domain.SeverityInfo},
		{Code: "org.keyboard.skip_missing", Severity: domain.SeverityHigh},
		{Code: "org.keyboard.nav_unreachable", Severity: domain.SeverityHigh},
		{Code: "org.keyboard.exec_failed", Severity: domain.SeverityHigh},
	})
	by := map[string]domain.Decision{}
	for _, d := range decisions {
		by[d.FindingCode] = d
	}
	if by["org.keyboard.ok"].Basket != domain.BasketAccept {
		t.Fatalf("%+v", by["org.keyboard.ok"])
	}
	if by["org.keyboard.skip_ok"].Basket != domain.BasketAccept {
		t.Fatalf("%+v", by["org.keyboard.skip_ok"])
	}
	if by["org.keyboard.skip_missing"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["org.keyboard.skip_missing"])
	}
	if by["org.keyboard.nav_unreachable"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["org.keyboard.nav_unreachable"])
	}
	if by["org.keyboard.exec_failed"].Basket != domain.BasketFixTheme {
		t.Fatalf("%+v", by["org.keyboard.exec_failed"])
	}
}
