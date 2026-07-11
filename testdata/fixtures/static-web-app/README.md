# Static-web fixture (Cycle E)

Prebuilt Vite-style `dist/` for the `static-web` adapter — **no Node required** for `mode: dist`.

```bash
go run ./apps/cli run -f testdata/manifests/quality-staticweb.lab.yaml
```

Optional preview (Node):

```yaml
adapter:
  id: static-web
  config:
    root: testdata/fixtures/static-web-app
    mode: preview
    previewCommand: "npx --yes serve dist -l {port}"
```
