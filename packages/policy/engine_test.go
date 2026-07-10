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

func TestEngineMapSecureBaseline(t *testing.T) {
	t.Parallel()
	e := NewEngine("secure-baseline")
	decisions := e.Map([]domain.Finding{
		{Code: "sec.recon.xmlrpc", Severity: domain.SeverityHigh},
	})
	if decisions[0].Basket != domain.BasketSiteDefaultOff {
		t.Fatalf("%+v", decisions[0])
	}
}
