package ziplint

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// Runner validates a WordPress theme zip against Gate 1 packaging rules
// (.project/check/theme-check.md).
type Runner struct{}

func New() *Runner { return &Runner{} }

func (r *Runner) ID() string { return "zip-lint" }

func (r *Runner) Run(_ context.Context, req ports.RunnerRequest) ([]domain.Finding, error) {
	zipPath := req.Check.Config["themeZip"]
	if zipPath == "" {
		zipPath = req.Target.Metadata["themeZip"]
	}
	if zipPath == "" {
		return []domain.Finding{{
			Code:     "org.zip.missing_path",
			Gate:     req.Gate,
			Check:    req.Check.ID,
			Severity: domain.SeverityHigh,
			Message:  "themeZip path not provided in check config or target metadata",
		}}, nil
	}

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return []domain.Finding{{
			Code:     "org.zip.open_failed",
			Gate:     req.Gate,
			Check:    req.Check.ID,
			Severity: domain.SeverityCritical,
			Message:  err.Error(),
		}}, nil
	}
	defer zr.Close()

	files := map[string]*zip.File{}
	var names []string
	var rootPrefix string
	for _, f := range zr.File {
		name := path.Clean(strings.ReplaceAll(f.Name, "\\", "/"))
		files[name] = f
		names = append(names, name)
		parts := strings.Split(name, "/")
		if rootPrefix == "" && len(parts) > 1 && !f.FileInfo().IsDir() {
			rootPrefix = parts[0] + "/"
		}
	}
	slug := strings.TrimSuffix(rootPrefix, "/")

	var findings []domain.Finding
	findings = append(findings, checkForbiddenPaths(req, names)...)
	findings = append(findings, checkForbiddenExtensions(req, names)...)
	findings = append(findings, checkRequiredFiles(req, files, rootPrefix)...)

	if f := lookup(files, rootPrefix, "screenshot.png"); f != nil {
		findings = append(findings, checkScreenshot(req, f)...)
	} else if f := lookup(files, rootPrefix, "screenshot.jpg"); f != nil {
		findings = append(findings, checkScreenshot(req, f)...)
	} else {
		findings = append(findings, finding(req, "org.zip.missing_screenshot", domain.SeverityHigh, "missing screenshot.png"))
	}

	var styleBody string
	if sf := lookup(files, rootPrefix, "style.css"); sf != nil {
		data, err := readAll(sf)
		if err != nil {
			findings = append(findings, finding(req, "org.zip.style_read", domain.SeverityMedium, err.Error()))
		} else {
			styleBody = string(data)
			findings = append(findings, checkStyleCSS(req, styleBody, slug)...)
		}
	}

	var readmeBody string
	if rf := lookup(files, rootPrefix, "readme.txt"); rf != nil {
		if data, err := readAll(rf); err == nil {
			readmeBody = string(data)
		}
	}
	findings = append(findings, checkResources(req, names, rootPrefix, readmeBody)...)
	findings = append(findings, checkMinifiedTwins(req, names, rootPrefix)...)
	findings = append(findings, checkPolicyScan(req, files, rootPrefix)...)

	if len(findings) == 0 {
		findings = append(findings, finding(req, "org.zip.ok", domain.SeverityInfo, "theme zip Gate 1 packaging checks passed"))
	}
	return findings, nil
}

func finding(req ports.RunnerRequest, code string, sev domain.Severity, msg string) domain.Finding {
	return domain.Finding{Code: code, Gate: req.Gate, Check: req.Check.ID, Severity: sev, Message: msg}
}

func hasFile(files map[string]*zip.File, root, rel string) bool {
	return lookup(files, root, rel) != nil
}

func lookup(files map[string]*zip.File, root, rel string) *zip.File {
	cands := []string{rel, path.Join(strings.TrimSuffix(root, "/"), rel)}
	for _, c := range cands {
		c = path.Clean(c)
		if f, ok := files[c]; ok && !f.FileInfo().IsDir() {
			return f
		}
	}
	for name, f := range files {
		if (strings.HasSuffix(name, "/"+rel) || name == rel) && !f.FileInfo().IsDir() {
			return f
		}
	}
	return nil
}

func readAll(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func relPath(name, root string) string {
	name = path.Clean(strings.ReplaceAll(name, "\\", "/"))
	if root != "" && strings.HasPrefix(name, root) {
		return strings.TrimPrefix(name, root)
	}
	return name
}

// --- Gate 1 checks ---

var forbiddenPathParts = []string{
	"/.git/", "/.cursor/", "/node_modules/",
	"/packages/", "/.workspaces/", "/.project/",
	"/vendor/bin/", // often accidental; composer vendor in theme is reviewed separately
}

func checkForbiddenPaths(req ports.RunnerRequest, names []string) []domain.Finding {
	var out []domain.Finding
	seen := map[string]bool{}
	for _, name := range names {
		lower := "/" + strings.ToLower(name) + "/"
		// normalize for contains checks
		padded := "/" + strings.ToLower(strings.ReplaceAll(name, "\\", "/"))
		for _, part := range forbiddenPathParts {
			if strings.Contains(padded+"/", part) || strings.HasPrefix(strings.ToLower(name), strings.Trim(part, "/")) {
				key := part + name
				if seen[key] {
					continue
				}
				seen[key] = true
				out = append(out, finding(req, "org.zip.forbidden_path", domain.SeverityHigh,
					fmt.Sprintf("forbidden path segment %s in zip: %s", strings.Trim(part, "/"), name)))
			}
		}
		_ = lower
		base := strings.ToLower(path.Base(name))
		if base == "thumbs.db" || base == "desktop.ini" || base == ".ds_store" {
			out = append(out, finding(req, "org.zip.forbidden_path", domain.SeverityHigh, "OS junk file in zip: "+name))
		}
	}
	return out
}

// Allowed XML filenames per WP.org Required §9 exceptions (subset).
var xmlAllow = map[string]bool{
	"wpml-config.xml": true,
	"loco.xml":        true,
	"phpcs.xml":       true,
	"phpcs.xml.dist":  true,
}

var forbiddenExt = map[string]string{
	".sh":  "org.zip.forbidden_ext_sh",
	".sql": "org.zip.forbidden_ext_sql",
	".zip": "org.zip.nested_zip",
}

func checkForbiddenExtensions(req ports.RunnerRequest, names []string) []domain.Finding {
	var out []domain.Finding
	for _, name := range names {
		if strings.HasSuffix(name, "/") {
			continue
		}
		lower := strings.ToLower(name)
		base := path.Base(lower)
		ext := path.Ext(lower)
		if ext == ".xml" {
			if !xmlAllow[base] {
				out = append(out, finding(req, "org.zip.forbidden_ext_xml", domain.SeverityHigh,
					"XML file not on allowlist: "+name))
			}
			continue
		}
		if code, ok := forbiddenExt[ext]; ok {
			out = append(out, finding(req, code, domain.SeverityHigh, "forbidden extension in zip: "+name))
		}
	}
	return out
}

func checkRequiredFiles(req ports.RunnerRequest, files map[string]*zip.File, root string) []domain.Finding {
	var out []domain.Finding
	for _, rel := range []string{"style.css", "readme.txt", "functions.php"} {
		if !hasFile(files, root, rel) {
			out = append(out, finding(req, "org.zip.missing_"+strings.ReplaceAll(rel, ".", "_"), domain.SeverityHigh, "missing required file: "+rel))
		}
	}
	if !hasFile(files, root, "LICENSE") && !hasFile(files, root, "license.txt") && !hasFile(files, root, "LICENSE.txt") {
		out = append(out, finding(req, "org.zip.missing_license", domain.SeverityHigh, "missing LICENSE or license.txt"))
	}
	return out
}

func checkScreenshot(req ports.RunnerRequest, f *zip.File) []domain.Finding {
	data, err := readAll(f)
	if err != nil {
		return []domain.Finding{finding(req, "org.zip.screenshot_read", domain.SeverityMedium, err.Error())}
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return []domain.Finding{finding(req, "org.zip.screenshot_decode", domain.SeverityHigh, "cannot decode screenshot: "+err.Error())}
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	var out []domain.Finding
	if w > 1200 || h > 900 {
		out = append(out, finding(req, "org.zip.screenshot_size", domain.SeverityHigh,
			fmt.Sprintf("screenshot %dx%d exceeds 1200x900", w, h)))
	}
	ratio := float64(w) / float64(h)
	if ratio < 1.2 || ratio > 1.45 {
		out = append(out, finding(req, "org.zip.screenshot_ratio", domain.SeverityMedium,
			fmt.Sprintf("screenshot ratio %.3f not ~4:3", ratio)))
	}
	return out
}

var (
	reVersion     = regexp.MustCompile(`(?im)^\s*\*?\s*Version:\s*(\S+)`)
	reRequires    = regexp.MustCompile(`(?im)^\s*\*?\s*Requires at least:\s*(\S+)`)
	reTestedUpTo  = regexp.MustCompile(`(?im)^\s*\*?\s*Tested up to:\s*(\S+)`)
	reTextDomain  = regexp.MustCompile(`(?im)^\s*\*?\s*Text Domain:\s*(\S+)`)
	reTagsLine    = regexp.MustCompile(`(?im)^\s*\*?\s*Tags:\s*(.+)$`)
	reRequiresPHP = regexp.MustCompile(`(?im)^\s*\*?\s*Requires PHP:\s*(\S+)`)
)

func checkStyleCSS(req ports.RunnerRequest, s, slug string) []domain.Finding {
	var out []domain.Finding
	if !reVersion.MatchString(s) {
		out = append(out, finding(req, "org.zip.style_version", domain.SeverityHigh, "style.css missing Version header"))
	}
	if !reRequires.MatchString(s) {
		out = append(out, finding(req, "org.zip.style_requires", domain.SeverityHigh, "style.css missing Requires at least header"))
	}
	if !reTestedUpTo.MatchString(s) {
		out = append(out, finding(req, "org.zip.style_tested", domain.SeverityHigh, "style.css missing Tested up to header"))
	}
	if !reRequiresPHP.MatchString(s) {
		out = append(out, finding(req, "org.zip.style_requires_php", domain.SeverityMedium, "style.css missing Requires PHP header"))
	}
	m := reTextDomain.FindStringSubmatch(s)
	if m == nil {
		out = append(out, finding(req, "org.zip.style_textdomain", domain.SeverityHigh, "style.css missing Text Domain header"))
	} else if slug != "" && m[1] != slug {
		out = append(out, finding(req, "org.zip.style_textdomain_slug", domain.SeverityHigh,
			fmt.Sprintf("Text Domain %q must match theme folder slug %q", m[1], slug)))
	}

	allowA11y := req.Check.Config["allowAccessibilityReady"] == "true"
	allowWoo := req.Check.Config["allowEcommerce"] == "true"
	if tags := reTagsLine.FindStringSubmatch(s); tags != nil {
		tagStr := strings.ToLower(tags[1])
		if !allowA11y && strings.Contains(tagStr, "accessibility-ready") {
			out = append(out, finding(req, "org.zip.tag_accessibility_ready", domain.SeverityHigh,
				"style.css claims accessibility-ready; blocked until focus-trap verified"))
		}
		if !allowWoo && (strings.Contains(tagStr, "e-commerce") || strings.Contains(tagStr, "ecommerce") || strings.Contains(tagStr, "woocommerce")) {
			out = append(out, finding(req, "org.zip.tag_ecommerce", domain.SeverityHigh,
				"style.css claims e-commerce/woocommerce without allowEcommerce=true"))
		}
	}
	return out
}

func checkResources(req ports.RunnerRequest, names []string, root, readme string) []domain.Finding {
	resourcesSection := extractResourcesSection(readme)
	var out []domain.Finding
	if resourcesSection == "" && hasAssetOrLib(names, root) {
		out = append(out, finding(req, "org.zip.resources_missing_section", domain.SeverityHigh,
			"readme.txt missing == Resources == section but assets/ or lib/ present"))
		return out
	}
	resLower := strings.ToLower(resourcesSection)
	seen := map[string]bool{}
	for _, name := range names {
		rel := relPath(name, root)
		lower := strings.ToLower(rel)
		if !(strings.HasPrefix(lower, "assets/") || strings.HasPrefix(lower, "lib/")) {
			continue
		}
		if strings.HasSuffix(lower, "/") {
			continue
		}
		base := path.Base(lower)
		// skip tiny/noise
		if base == "." || base == ".." || strings.HasPrefix(base, ".") {
			continue
		}
		// require basename (without .min) mentioned in Resources
		token := strings.TrimSuffix(base, path.Ext(base))
		token = strings.TrimSuffix(token, ".min")
		if token == "" || seen[token] {
			continue
		}
		seen[token] = true
		if resourcesSection != "" && !strings.Contains(resLower, strings.ToLower(token)) && !strings.Contains(resLower, base) {
			out = append(out, finding(req, "org.zip.resources_unattributed", domain.SeverityMedium,
				fmt.Sprintf("bundled file %s not mentioned in readme Resources", rel)))
		}
	}
	return out
}

func hasAssetOrLib(names []string, root string) bool {
	for _, name := range names {
		rel := strings.ToLower(relPath(name, root))
		if strings.HasPrefix(rel, "assets/") || strings.HasPrefix(rel, "lib/") {
			return true
		}
	}
	return false
}

func extractResourcesSection(readme string) string {
	lower := strings.ToLower(readme)
	idx := strings.Index(lower, "== resources ==")
	if idx < 0 {
		idx = strings.Index(lower, "==resources==")
	}
	if idx < 0 {
		return ""
	}
	rest := readme[idx:]
	// next == section
	lines := strings.Split(rest, "\n")
	var b strings.Builder
	b.WriteString(lines[0])
	b.WriteByte('\n')
	for i := 1; i < len(lines); i++ {
		l := strings.TrimSpace(lines[i])
		if strings.HasPrefix(l, "==") && !strings.EqualFold(strings.Trim(l, "= "), "Resources") {
			break
		}
		b.WriteString(lines[i])
		b.WriteByte('\n')
	}
	return b.String()
}

func checkMinifiedTwins(req ports.RunnerRequest, names []string, root string) []domain.Finding {
	set := map[string]bool{}
	for _, name := range names {
		if strings.HasSuffix(name, "/") {
			continue
		}
		set[strings.ToLower(relPath(name, root))] = true
	}
	var out []domain.Finding
	for rel := range set {
		if !strings.Contains(rel, ".min.") {
			continue
		}
		// foo.min.js -> foo.js ; style.min.css -> style.css
		twin := strings.Replace(rel, ".min.", ".", 1)
		if !set[twin] {
			out = append(out, finding(req, "org.zip.minified_without_source", domain.SeverityHigh,
				fmt.Sprintf("minified %s without unminified twin %s", rel, twin)))
		}
	}
	return out
}

var policyPatterns = []struct {
	code    string
	re      *regexp.Regexp
	sev     domain.Severity
	message string
}{
	{"org.zip.policy_cpt", regexp.MustCompile(`(?i)register_post_type\s*\(`), domain.SeverityHigh, "register_post_type found (plugin territory)"},
	{"org.zip.policy_shortcode", regexp.MustCompile(`(?i)add_shortcode\s*\(`), domain.SeverityHigh, "add_shortcode found (plugin territory)"},
	{"org.zip.policy_woocommerce", regexp.MustCompile(`(?i)woocommerce`), domain.SeverityMedium, "woocommerce reference found"},
	{"org.zip.policy_comments_template", regexp.MustCompile(`(?i)comments_template\s*\(`), domain.SeverityMedium, "comments_template found"},
	{"org.zip.policy_contact_form", regexp.MustCompile(`(?i)(contact[_\- ]?form|wp_mail\s*\(.*\$_(POST|REQUEST)|newsletter[_\- ]?signup)`), domain.SeverityMedium, "possible contact/newsletter form pattern"},
}

func checkPolicyScan(req ports.RunnerRequest, files map[string]*zip.File, root string) []domain.Finding {
	if req.Check.Config["skipPolicyScan"] == "true" {
		return nil
	}
	allowWoo := req.Check.Config["allowEcommerce"] == "true"
	var out []domain.Finding
	hit := map[string]bool{}
	for name, f := range files {
		if f.FileInfo().IsDir() {
			continue
		}
		rel := relPath(name, root)
		lower := strings.ToLower(rel)
		if !strings.HasSuffix(lower, ".php") && !strings.HasSuffix(lower, ".js") && !strings.HasSuffix(lower, ".latte") {
			continue
		}
		// skip vendor bulk noise except flagging once is enough from theme PHP
		if strings.Contains(lower, "/vendor/") || strings.Contains(lower, "/node_modules/") {
			continue
		}
		data, err := readAll(f)
		if err != nil {
			continue
		}
		body := string(data)
		for _, p := range policyPatterns {
			if allowWoo && p.code == "org.zip.policy_woocommerce" {
				continue
			}
			if hit[p.code] {
				continue
			}
			if p.re.MatchString(body) {
				hit[p.code] = true
				out = append(out, finding(req, p.code, p.sev, p.message+" in "+rel))
			}
		}
	}
	return out
}
