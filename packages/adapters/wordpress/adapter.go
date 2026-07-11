package wordpress

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fastygo/lab/packages/domain"
)

// Adapter is a stub WordPress target: config supplies baseUrl and optional themeZip.
type Adapter struct {
	repoRoot string
	baseURL  string
	themeZip string
	seedPath string
	seed     orgSeed
}

type orgSeed struct {
	AttachmentID string `json:"attachmentId"`
	PostID       string `json:"postId"`
	PageID       string `json:"pageId"`
	CatID        string `json:"catId"`
	TagSlug      string `json:"tagSlug"`
	Imported     bool   `json:"imported"`
}

func New(repoRoot string) *Adapter {
	return &Adapter{repoRoot: repoRoot}
}

func (a *Adapter) ID() string { return "wordpress" }

func (a *Adapter) Capabilities() []string {
	return []string{"http", "html", "a11y", "seo", "wp-theme"}
}

func (a *Adapter) Prepare(_ context.Context, config map[string]string) error {
	a.baseURL = config["baseUrl"]
	if a.baseURL == "" {
		a.baseURL = "http://127.0.0.1:8080"
	}
	zip := config["themeZip"]
	if zip != "" {
		resolved, err := resolveThemeZip(zip, a.repoRoot)
		if err != nil {
			return fmt.Errorf("themeZip: %w", err)
		}
		a.themeZip = resolved
	}
	a.seedPath = config["seedFile"]
	if a.seedPath == "" {
		a.seedPath = filepath.Join(a.repoRoot, "testdata", "fixtures", "org-seed.json")
	} else if !filepath.IsAbs(a.seedPath) && a.repoRoot != "" {
		a.seedPath = filepath.Join(a.repoRoot, a.seedPath)
	}
	a.seed = loadSeed(a.seedPath)
	// Config overrides for matrix IDs (useful in tests).
	if v := config["attachmentId"]; v != "" {
		a.seed.AttachmentID = v
	}
	if v := config["postId"]; v != "" {
		a.seed.PostID = v
	}
	if v := config["pageId"]; v != "" {
		a.seed.PageID = v
	}
	if v := config["catId"]; v != "" {
		a.seed.CatID = v
	}
	if v := config["tagSlug"]; v != "" {
		a.seed.TagSlug = v
	}
	return nil
}

func loadSeed(path string) orgSeed {
	s := orgSeed{
		PostID:  "1",
		PageID:  "2",
		CatID:   "1",
		TagSlug: "test",
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	_ = json.Unmarshal(b, &s)
	if s.PostID == "" {
		s.PostID = "1"
	}
	if s.PageID == "" {
		s.PageID = "2"
	}
	if s.CatID == "" {
		s.CatID = "1"
	}
	if s.TagSlug == "" {
		s.TagSlug = "test"
	}
	return s
}

func resolveThemeZip(zip, repoRoot string) (string, error) {
	candidates := []string{zip}
	if !filepath.IsAbs(zip) {
		if wd, err := os.Getwd(); err == nil {
			candidates = append(candidates, filepath.Join(wd, zip))
		}
		if repoRoot != "" {
			candidates = append(candidates, filepath.Join(repoRoot, zip))
		}
	}
	var last error
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			last = err
			continue
		}
		if st, err := os.Stat(abs); err == nil && !st.IsDir() {
			return abs, nil
		} else if err != nil {
			last = err
		}
	}
	if last == nil {
		last = os.ErrNotExist
	}
	return "", last
}

func (a *Adapter) Serve(_ context.Context) (domain.Target, error) {
	meta := map[string]string{"adapter": "wordpress"}
	if a.themeZip != "" {
		meta["themeZip"] = a.themeZip
	}
	if a.seedPath != "" {
		meta["seedFile"] = a.seedPath
	}
	return domain.Target{BaseURL: a.baseURL, Metadata: meta}, nil
}

func (a *Adapter) Matrix(_ context.Context) ([]string, error) {
	// Reload seed each call so Gate 2 seed can expand URLs before Gate 3.
	if a.seedPath != "" {
		a.seed = loadSeed(a.seedPath)
	}
	// Query-string URLs work without Apache rewrite/.htaccess (compose org default).
	b := a.baseURL
	urls := []string{
		b + "/",
		b + "/?p=" + a.seed.PostID,
		b + "/?page_id=" + a.seed.PageID,
		b + "/?cat=" + a.seed.CatID,
		b + "/?tag=" + a.seed.TagSlug,
		b + "/?author=1",
		b + "/?s=hello",
		b + "/?p=999999&lab-404=1",
	}
	if a.seed.AttachmentID != "" {
		urls = append(urls, b+"/?attachment_id="+a.seed.AttachmentID)
	}
	return urls, nil
}

func (a *Adapter) Teardown(_ context.Context) error { return nil }
