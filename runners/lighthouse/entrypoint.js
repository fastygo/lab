#!/usr/bin/env node
"use strict";

const { spawnSync } = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");

const target = process.env.LAB_TARGET_URL || "http://127.0.0.1/";
const gate = process.env.LAB_GATE_ID || "";
const check = process.env.LAB_CHECK_ID || "";
let cfg = {};
try {
  cfg = JSON.parse(process.env.LAB_CONFIG_JSON || "{}");
} catch (_) {}

const failBelow = {
  performance: num(cfg.perfFail, 0.7),
  accessibility: num(cfg.a11yFail, 0.9),
  "best-practices": num(cfg.bpFail, 0.9),
  seo: num(cfg.seoFail, 0.9),
};
const warnBelow = {
  performance: num(cfg.perfWarn, 0.9),
  accessibility: num(cfg.a11yWarn, 0.95),
  "best-practices": num(cfg.bpWarn, 0.95),
  seo: num(cfg.seoWarn, 0.95),
};

function num(v, d) {
  const n = Number(v);
  return Number.isFinite(n) ? n : d;
}

const outDir = fs.mkdtempSync(path.join(os.tmpdir(), "lh-"));
const outPath = path.join(outDir, "report.json");
const args = [
  "lighthouse",
  target,
  "--only-categories=performance,accessibility,best-practices,seo",
  "--form-factor=mobile",
  "--screenEmulation.mobile",
  "--chrome-flags=--headless --no-sandbox --disable-gpu",
  "--output=json",
  `--output-path=${outPath}`,
  "--quiet",
];
const r = spawnSync("npx", ["--yes", ...args], { encoding: "utf8", timeout: 180000 });
if (r.status !== 0 && !fs.existsSync(outPath)) {
  console.log(
    JSON.stringify({
      findings: [
        {
          code: "quality.lighthouse.exec_failed",
          gate,
          check,
          severity: "high",
          message: (r.stderr || r.stdout || "lighthouse failed").slice(0, 500),
          target,
        },
      ],
    })
  );
  process.exit(0);
}

const report = JSON.parse(fs.readFileSync(outPath, "utf8"));
const cats = report.categories || {};
const findings = [];
for (const [id, cat] of Object.entries(cats)) {
  const score = typeof cat.score === "number" ? cat.score : 0;
  const code = `quality.lighthouse.${id}`;
  let severity = "info";
  if (score < (failBelow[id] ?? 0.7)) severity = "high";
  else if (score < (warnBelow[id] ?? 0.9)) severity = "medium";
  findings.push({
    code,
    gate,
    check,
    severity,
    message: `${id} score ${(score * 100).toFixed(0)}`,
    target,
    evidence: { score: String(score) },
  });
}
console.log(JSON.stringify({ findings }));
