package ziplint

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/png"
	_ "image/jpeg"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// Runner validates a WordPress theme zip for .org packaging basics.
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

	var findings []domain.Finding
	files := map[string]*zip.File{}
	var rootPrefix string
	for _, f := range zr.File {
		name := path.Clean(strings.ReplaceAll(f.Name, "\\", "/"))
		files[name] = f
		parts := strings.Split(name, "/")
		if rootPrefix == "" && len(parts) > 1 {
			rootPrefix = parts[0] + "/"
		}
		lower := strings.ToLower(name)
		if strings.Contains(lower, "/.git/") || strings.HasPrefix(lower, ".git/") || strings.Contains(lower, "/.cursor/") {
			findings = append(findings, finding(req, "org.zip.forbidden_path", domain.SeverityHigh, "forbidden path in zip: "+name))
		}
		if strings.Contains(lower, "/node_modules/") {
			findings = append(findings, finding(req, "org.zip.forbidden_path", domain.SeverityHigh, "node_modules in zip: "+name))
		}
	}

	required := []string{"style.css", "readme.txt", "functions.php"}
	for _, rel := range required {
		if !hasFile(files, rootPrefix, rel) {
			findings = append(findings, finding(req, "org.zip.missing_"+strings.ReplaceAll(rel, ".", "_"), domain.SeverityHigh, "missing required file: "+rel))
		}
	}
	if !hasFile(files, rootPrefix, "LICENSE") && !hasFile(files, rootPrefix, "license.txt") && !hasFile(files, rootPrefix, "LICENSE.txt") {
		findings = append(findings, finding(req, "org.zip.missing_license", domain.SeverityHigh, "missing LICENSE or license.txt"))
	}
	shotName := ""
	for _, cand := range []string{"screenshot.png", "screenshot.jpg"} {
		if hasFile(files, rootPrefix, cand) {
			shotName = rootPrefix + cand
			if rootPrefix == "" {
				shotName = cand
			}
			// try with and without prefix
			if f := lookup(files, rootPrefix, cand); f != nil {
				shotName = f.Name
			}
			break
		}
	}
	if shotName == "" {
		findings = append(findings, finding(req, "org.zip.missing_screenshot", domain.SeverityHigh, "missing screenshot.png"))
	} else if f := files[path.Clean(strings.ReplaceAll(shotName, "\\", "/"))]; f != nil {
		findings = append(findings, checkScreenshot(req, f)...)
	} else if f := lookup(files, rootPrefix, "screenshot.png"); f != nil {
		findings = append(findings, checkScreenshot(req, f)...)
	}

	if sf := lookup(files, rootPrefix, "style.css"); sf != nil {
		findings = append(findings, checkStyleCSS(req, sf)...)
	}

	if len(findings) == 0 {
		findings = append(findings, finding(req, "org.zip.ok", domain.SeverityInfo, "theme zip packaging checks passed"))
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
	// suffix match
	for name, f := range files {
		if strings.HasSuffix(name, "/"+rel) || name == rel {
			if !f.FileInfo().IsDir() {
				return f
			}
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
	if ratio < 1.2 || ratio > 1.45 { // ~4:3 = 1.333
		out = append(out, finding(req, "org.zip.screenshot_ratio", domain.SeverityMedium,
			fmt.Sprintf("screenshot ratio %.3f not ~4:3", ratio)))
	}
	return out
}

var (
	reVersion    = regexp.MustCompile(`(?i)Version:\s*\S+`)
	reTextDomain = regexp.MustCompile(`(?i)Text Domain:\s*\S+`)
	reA11yTag    = regexp.MustCompile(`(?i)Tags:.*accessibility-ready`)
)

func checkStyleCSS(req ports.RunnerRequest, f *zip.File) []domain.Finding {
	data, err := readAll(f)
	if err != nil {
		return []domain.Finding{finding(req, "org.zip.style_read", domain.SeverityMedium, err.Error())}
	}
	s := string(data)
	var out []domain.Finding
	if !reVersion.MatchString(s) {
		out = append(out, finding(req, "org.zip.style_version", domain.SeverityHigh, "style.css missing Version header"))
	}
	if !reTextDomain.MatchString(s) {
		out = append(out, finding(req, "org.zip.style_textdomain", domain.SeverityHigh, "style.css missing Text Domain header"))
	}
	allowA11y := req.Check.Config["allowAccessibilityReady"] == "true"
	if !allowA11y && reA11yTag.MatchString(s) {
		out = append(out, finding(req, "org.zip.tag_accessibility_ready", domain.SeverityHigh,
			"style.css claims accessibility-ready; blocked until focus-trap verified"))
	}
	return out
}
