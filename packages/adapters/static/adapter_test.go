package static_test

import (
	"context"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fastygo/lab/packages/adapters/static"
)

func TestStaticServe(t *testing.T) {
	t.Parallel()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "testdata", "fixtures", "quality-site"))
	a := static.New(root)
	if err := a.Prepare(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	target, err := a.Serve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Teardown(context.Background()) })
	resp, err := http.Get(target.BaseURL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	urls, err := a.Matrix(context.Background())
	if err != nil || len(urls) != 2 {
		t.Fatalf("matrix=%v err=%v", urls, err)
	}
}
