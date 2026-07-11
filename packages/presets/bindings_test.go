package presets_test

import (
	"testing"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/presets"
)

func TestApplyBindings(t *testing.T) {
	t.Parallel()
	m := &domain.Manifest{
		APIVersion: "lab.fastygo.dev/v1",
		Metadata:   domain.ManifestMetadata{Name: "t"},
		Spec: domain.ManifestSpec{
			Lab: "org",
			Adapter: domain.AdapterRef{
				ID:     "wordpress",
				Config: map[string]string{"baseUrl": "http://old", "themeZip": "old.zip"},
			},
			Gates: []domain.Gate{{
				ID: "g1",
				Checks: []domain.Check{
					{ID: "zip-lint", Runner: "zip-lint"},
					{ID: "css", Runner: "css-lint", Config: map[string]string{"cssDir": "old"}},
				},
			}},
		},
	}
	presets.ApplyBindings(m, presets.Bindings{
		ThemeZip:    "testdata/dist/latte.zip",
		BaseURL:     "http://127.0.0.1:8080",
		CheckConfig: map[string]string{"cssDir": "testdata/fixtures/quality-site"},
	})
	if m.Spec.Adapter.Config["themeZip"] != "testdata/dist/latte.zip" {
		t.Fatalf("adapter themeZip=%q", m.Spec.Adapter.Config["themeZip"])
	}
	if m.Spec.Adapter.Config["baseUrl"] != "http://127.0.0.1:8080" {
		t.Fatalf("baseUrl=%q", m.Spec.Adapter.Config["baseUrl"])
	}
	if m.Spec.Gates[0].Checks[0].Config["themeZip"] != "testdata/dist/latte.zip" {
		t.Fatalf("zip-lint themeZip missing")
	}
	if m.Spec.Gates[0].Checks[1].Config["cssDir"] != "testdata/fixtures/quality-site" {
		t.Fatalf("cssDir=%q", m.Spec.Gates[0].Checks[1].Config["cssDir"])
	}
}
