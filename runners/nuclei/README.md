# nuclei

Thin wrapper around `projectdiscovery/nuclei` for **owned lab targets** only.

Default: `-tags wordpress` with severity `critical,high,medium`, bounded concurrency.

## Image

```bash
docker build -t lab/nuclei:local runners/nuclei
```

First run may pull templates (needs network). Override tags via check config `env.LAB_NUCLEI_TAGS`.

## Findings

- `sec.nuclei.match` — template hit
- `sec.nuclei.ok` / `sec.nuclei.completed`
