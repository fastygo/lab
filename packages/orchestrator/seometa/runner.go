package seometa

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// Runner asserts basic SEO / document meta on matrix URLs (Q5).
type Runner struct {
	client *http.Client
}

func New() *Runner {
	return &Runner{client: &http.Client{Timeout: 15 * time.Second}}
}

func (r *Runner) ID() string { return "seo-meta" }

func (r *Runner) Run(ctx context.Context, req ports.RunnerRequest) ([]domain.Finding, error) {
	urls := req.URLs
	if len(urls) == 0 {
		urls = []string{strings.TrimRight(req.Target.BaseURL, "/") + "/"}
	}
	seoSocial := req.Check.Config["seoSocial"] == "true"

	var findings []domain.Finding
	issues := 0
	for _, u := range urls {
		fs, n := r.checkURL(ctx, req, u, seoSocial)
		findings = append(findings, fs...)
		issues += n
	}

	if !seoSocial {
		findings = append(findings, domain.Finding{
			Code:     "quality.seo.social_skipped",
			Gate:     req.Gate,
			Check:    req.Check.ID,
			Severity: domain.SeverityInfo,
			Message:  "OG/Twitter/JSON-LD checks skipped (seoSocial!=true)",
			Target:   req.Target.BaseURL,
		})
	}

	if issues == 0 {
		findings = append(findings, domain.Finding{
			Code:     "quality.seo.ok",
			Gate:     req.Gate,
			Check:    req.Check.ID,
			Severity: domain.SeverityInfo,
			Message:  fmt.Sprintf("SEO meta checks passed (%d URL(s))", len(urls)),
			Target:   req.Target.BaseURL,
			Evidence: map[string]string{"urls": strconv.Itoa(len(urls))},
		})
	}
	findings = append(findings, domain.Finding{
		Code:     "quality.seo.summary",
		Gate:     req.Gate,
		Check:    req.Check.ID,
		Severity: domain.SeverityInfo,
		Message:  fmt.Sprintf("SEO meta: %d issue(s) across %d URL(s)", issues, len(urls)),
		Target:   req.Target.BaseURL,
		Evidence: map[string]string{
			"issues": strconv.Itoa(issues),
			"urls":   strconv.Itoa(len(urls)),
		},
	})
	return findings, nil
}

func (r *Runner) checkURL(ctx context.Context, req ports.RunnerRequest, url string, seoSocial bool) ([]domain.Finding, int) {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []domain.Finding{{
			Code: "quality.seo.fetch_failed", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: err.Error(), Target: url,
		}}, 1
	}
	hreq.Header.Set("User-Agent", "FastyGo-Lab-SEO-Meta/0.2")
	resp, err := r.client.Do(hreq)
	if err != nil {
		return []domain.Finding{{
			Code: "quality.seo.fetch_failed", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: err.Error(), Target: url,
		}}, 1
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return []domain.Finding{{
			Code: "quality.seo.fetch_failed", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: err.Error(), Target: url,
		}}, 1
	}

	meta, err := parseMeta(string(body))
	if err != nil {
		return []domain.Finding{{
			Code: "quality.seo.parse_failed", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: err.Error(), Target: url,
		}}, 1
	}

	var out []domain.Finding
	issues := 0
	add := func(code string, sev domain.Severity, msg string, evidence map[string]string) {
		if sev == domain.SeverityHigh || sev == domain.SeverityMedium {
			issues++
		}
		out = append(out, domain.Finding{
			Code: code, Gate: req.Gate, Check: req.Check.ID,
			Severity: sev, Message: msg, Target: url, Evidence: evidence,
		})
	}

	if strings.TrimSpace(meta.Title) == "" {
		add("quality.seo.title_missing", domain.SeverityHigh, "Missing or empty <title>", nil)
	} else {
		add("quality.seo.title_ok", domain.SeverityInfo, "title: "+trim(meta.Title, 80), map[string]string{"title": meta.Title})
	}

	if !meta.HasViewport {
		add("quality.seo.viewport_missing", domain.SeverityHigh, "Missing meta viewport", nil)
	} else {
		add("quality.seo.viewport_ok", domain.SeverityInfo, "meta viewport present", nil)
	}

	switch meta.H1Count {
	case 1:
		add("quality.seo.h1_ok", domain.SeverityInfo, "exactly one h1", map[string]string{"h1": strconv.Itoa(meta.H1Count)})
	case 0:
		add("quality.seo.h1", domain.SeverityMedium, "no h1 on page", map[string]string{"h1": "0"})
	default:
		add("quality.seo.h1", domain.SeverityMedium, fmt.Sprintf("%d h1 elements (expected 1)", meta.H1Count), map[string]string{"h1": strconv.Itoa(meta.H1Count)})
	}

	if strings.TrimSpace(meta.Description) == "" {
		add("quality.seo.description_missing", domain.SeverityInfo, "meta description missing (soft)", nil)
	} else {
		add("quality.seo.description_ok", domain.SeverityInfo, "meta description present", nil)
	}

	if seoSocial {
		if meta.OGTitle == "" {
			add("quality.seo.og_title_missing", domain.SeverityMedium, "og:title missing", nil)
		} else {
			add("quality.seo.og_ok", domain.SeverityInfo, "og:title present", map[string]string{"og:title": meta.OGTitle})
		}
		if meta.OGType == "" {
			add("quality.seo.og_type_missing", domain.SeverityMedium, "og:type missing", nil)
		} else {
			add("quality.seo.og_type_ok", domain.SeverityInfo, "og:type present", map[string]string{"og:type": meta.OGType})
		}
		if meta.OGURL == "" {
			add("quality.seo.og_url_missing", domain.SeverityInfo, "og:url missing (soft)", nil)
		} else {
			add("quality.seo.og_url_ok", domain.SeverityInfo, "og:url present", nil)
		}
		if meta.TwitterCard == "" {
			add("quality.seo.twitter_missing", domain.SeverityMedium, "twitter:card missing", nil)
		} else {
			add("quality.seo.twitter_ok", domain.SeverityInfo, "twitter:card present", map[string]string{"twitter:card": meta.TwitterCard})
		}
		if meta.JSONLDCount == 0 {
			add("quality.seo.jsonld_missing", domain.SeverityMedium, "no JSON-LD blocks", nil)
		} else if meta.JSONLDInvalid > 0 {
			add("quality.seo.jsonld_invalid", domain.SeverityMedium, fmt.Sprintf("%d invalid JSON-LD block(s)", meta.JSONLDInvalid), map[string]string{
				"invalid": strconv.Itoa(meta.JSONLDInvalid),
				"total":   strconv.Itoa(meta.JSONLDCount),
			})
		} else {
			add("quality.seo.jsonld_ok", domain.SeverityInfo, fmt.Sprintf("%d valid JSON-LD block(s)", meta.JSONLDCount), map[string]string{"count": strconv.Itoa(meta.JSONLDCount)})
		}
	}

	return out, issues
}

type pageMeta struct {
	Title          string
	Description    string
	HasViewport    bool
	H1Count        int
	OGTitle        string
	OGType         string
	OGURL          string
	TwitterCard    string
	JSONLDCount    int
	JSONLDInvalid  int
}

func parseMeta(src string) (pageMeta, error) {
	doc, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return pageMeta{}, err
	}
	var m pageMeta
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				m.Title = textContent(n)
			case "meta":
				name := attr(n, "name")
				prop := attr(n, "property")
				content := attr(n, "content")
				if strings.EqualFold(name, "viewport") {
					m.HasViewport = true
				}
				if strings.EqualFold(name, "description") {
					m.Description = content
				}
				if strings.EqualFold(name, "twitter:card") {
					m.TwitterCard = content
				}
				if strings.EqualFold(prop, "og:title") {
					m.OGTitle = content
				}
				if strings.EqualFold(prop, "og:type") {
					m.OGType = content
				}
				if strings.EqualFold(prop, "og:url") {
					m.OGURL = content
				}
			case "h1":
				m.H1Count++
			case "script":
				if strings.EqualFold(attr(n, "type"), "application/ld+json") {
					m.JSONLDCount++
					raw := strings.TrimSpace(textContent(n))
					if raw == "" {
						m.JSONLDInvalid++
					} else {
						var v any
						if err := json.Unmarshal([]byte(raw), &v); err != nil {
							m.JSONLDInvalid++
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return m, nil
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(c *html.Node) {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
		for x := c.FirstChild; x != nil; x = x.NextSibling {
			walk(x)
		}
	}
	walk(n)
	return strings.TrimSpace(b.String())
}

func trim(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
