# Theme zip fixtures

Preferred artifact for org lab:

```text
testdata/dist/latte.zip
```

Wired in `testdata/manifests/org.lab.yaml`:

```yaml
adapter:
  id: wordpress
  config:
    baseUrl: http://127.0.0.1:8080
    themeZip: testdata/dist/latte.zip
```

Compose mounts `testdata/dist` at `/themes` for manual wp-cli installs. Gate 2 Theme Check mounts the zip via the Docker runner (`/lab/theme.zip`).
