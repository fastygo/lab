# Theme zip fixtures

Place sample theme zips here for local org lab runs, or set:

```yaml
spec:
  adapter:
    config:
      themeZip: /absolute/path/to/latte.zip
```

Compose mounts this directory at `/themes` inside the WordPress container when `LAB_THEME_ZIP_DIR` is unset (defaults to this folder).
