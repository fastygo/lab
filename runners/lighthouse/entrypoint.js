#!/usr/bin/env node
"use strict";

/**
 * Q1 Lighthouse — median of N runs + Core Web Vitals asserts.
 */

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

const runs = Math.max(1, Math.min(5, Number(cfg.runs || 3) || 3));

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

// CWV thresholds (seconds / unitless). Fail / warn.
const cwv = {
  lcpFail: num(cfg.lcpFail, 4.0),
  lcpWarn: num(cfg.lcpWarn, 2.5),
  clsFail: num(cfg.clsFail, 0.25),
  clsWarn: num(cfg.clsWarn, 0.1),
  tbtFail: num(cfg.tbtFail, 600), // ms — INP proxy
  tbtWarn: num(cfg.tbtWarn, 300),
  fcpFail: num(cfg.fcpFail, 3.0),
  fcpWarn: num(cfg.fcpWarn, 1.8),
};

function num(v, d) {
  const n = Number(v);
  return Number.isFinite(n) ? n : d;
}

function median(values) {
  const a = values.filter((v) => Number.isFinite(v)).slice().sort((x, y) => x - y);
  if (a.length === 0) return NaN;
  const mid = Math.floor(a.length / 2);
  return a.length % 2 === 0 ? (a[mid - 1] + a[mid]) / 2 : a[mid];
}

function runOnce(outPath) {
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
  return spawnSync("npx", ["--yes", ...args], { encoding: "utf8", timeout: 180000 });
}

function auditNumeric(report, id) {
  const a = report.audits && report.audits[id];
  if (!a) return NaN;
  if (typeof a.numericValue === "number") return a.numericValue;
  return NaN;
}

const outDir = fs.mkdtempSync(path.join(os.tmpdir(), "lh-"));
const reports = [];
const errors = [];

for (let i = 0; i < runs; i++) {
  const outPath = path.join(outDir, `report-${i}.json`);
  const r = runOnce(outPath);
  if (!fs.existsSync(outPath)) {
    errors.push((r.stderr || r.stdout || `run ${i} failed`).slice(0, 200));
    continue;
  }
  try {
    reports.push(JSON.parse(fs.readFileSync(outPath, "utf8")));
  } catch (e) {
    errors.push(String(e.message || e).slice(0, 200));
  }
}

if (reports.length === 0) {
  console.log(
    JSON.stringify({
      findings: [
        {
          code: "quality.lighthouse.exec_failed",
          gate,
          check,
          severity: "high",
          message: (errors[0] || "lighthouse failed").slice(0, 500),
          target,
          evidence: { runs: String(runs), errors: String(errors.length) },
        },
      ],
    })
  );
  process.exit(0);
}

const findings = [];
const categories = ["performance", "accessibility", "best-practices", "seo"];
for (const id of categories) {
  const scores = reports.map((rep) => {
    const cat = (rep.categories || {})[id];
    return cat && typeof cat.score === "number" ? cat.score : NaN;
  });
  const score = median(scores);
  if (!Number.isFinite(score)) continue;
  let severity = "info";
  if (score < (failBelow[id] ?? 0.7)) severity = "high";
  else if (score < (warnBelow[id] ?? 0.9)) severity = "medium";
  findings.push({
    code: `quality.lighthouse.${id}`,
    gate,
    check,
    severity,
    message: `${id} median score ${(score * 100).toFixed(0)} (n=${reports.length})`,
    target,
    evidence: {
      score: String(score),
      median: String(score),
      runs: String(reports.length),
      samples: scores.map((s) => (Number.isFinite(s) ? s.toFixed(3) : "na")).join(","),
    },
  });
}

function pushCWV(code, label, values, unit, warnAt, failAt, higherIsWorse) {
  const med = median(values);
  if (!Number.isFinite(med)) {
    findings.push({
      code: `${code}.missing`,
      gate,
      check,
      severity: "info",
      message: `${label} not present in Lighthouse report`,
      target,
    });
    return;
  }
  let severity = "info";
  if (higherIsWorse) {
    if (med > failAt) severity = "high";
    else if (med > warnAt) severity = "medium";
  } else {
    if (med < failAt) severity = "high";
    else if (med < warnAt) severity = "medium";
  }
  const display =
    unit === "ms" ? `${Math.round(med)}ms` : unit === "s" ? `${(med / 1000).toFixed(2)}s` : med.toFixed(3);
  findings.push({
    code,
    gate,
    check,
    severity,
    message: `${label} median ${display} (n=${reports.length})`,
    target,
    evidence: {
      median: String(med),
      unit,
      warnAt: String(warnAt),
      failAt: String(failAt),
      runs: String(reports.length),
    },
  });
}

// Lighthouse numericValue: LCP/FCP in ms, CLS unitless, TBT in ms
pushCWV(
  "quality.lighthouse.lcp",
  "LCP",
  reports.map((r) => auditNumeric(r, "largest-contentful-paint")),
  "s",
  cwv.lcpWarn * 1000,
  cwv.lcpFail * 1000,
  true
);
pushCWV(
  "quality.lighthouse.cls",
  "CLS",
  reports.map((r) => auditNumeric(r, "cumulative-layout-shift")),
  "score",
  cwv.clsWarn,
  cwv.clsFail,
  true
);
pushCWV(
  "quality.lighthouse.tbt",
  "TBT (INP proxy)",
  reports.map((r) => auditNumeric(r, "total-blocking-time")),
  "ms",
  cwv.tbtWarn,
  cwv.tbtFail,
  true
);
pushCWV(
  "quality.lighthouse.fcp",
  "FCP",
  reports.map((r) => auditNumeric(r, "first-contentful-paint")),
  "s",
  cwv.fcpWarn * 1000,
  cwv.fcpFail * 1000,
  true
);

findings.push({
  code: "quality.lighthouse.summary",
  gate,
  check,
  severity: "info",
  message: `Lighthouse median of ${reports.length}/${runs} run(s)`,
  target,
  evidence: { runs: String(reports.length), requested: String(runs) },
});

// Resource byte budgets (median across runs).
const budgets = {
  totalFail: num(cfg.totalByteFail, 1500000),
  totalWarn: num(cfg.totalByteWarn, 500000),
  scriptFail: num(cfg.scriptByteFail, 300000),
  scriptWarn: num(cfg.scriptByteWarn, 80000),
  styleFail: num(cfg.styleByteFail, 400000),
  styleWarn: num(cfg.styleByteWarn, 100000),
};

function resourceSummary(report) {
  const out = { total: NaN, script: NaN, stylesheet: NaN, image: NaN, font: NaN };
  const tw = auditNumeric(report, "total-byte-weight");
  if (Number.isFinite(tw)) out.total = tw;
  const rs = report.audits && report.audits["resource-summary"];
  const items = (rs && rs.details && rs.details.items) || [];
  for (const it of items) {
    const t = String(it.resourceType || "").toLowerCase();
    const size = Number(it.transferSize != null ? it.transferSize : it.size);
    if (!Number.isFinite(size)) continue;
    if (t === "total") out.total = size;
    else if (t === "script") out.script = size;
    else if (t === "stylesheet") out.stylesheet = size;
    else if (t === "image") out.image = size;
    else if (t === "font") out.font = size;
  }
  return out;
}

function pushBudget(code, label, values, warnAt, failAt) {
  const med = median(values);
  if (!Number.isFinite(med)) {
    findings.push({
      code: `${code}.missing`,
      gate,
      check,
      severity: "info",
      message: `${label} bytes not in Lighthouse report`,
      target,
    });
    return;
  }
  let severity = "info";
  if (med > failAt) severity = "high";
  else if (med > warnAt) severity = "medium";
  findings.push({
    code,
    gate,
    check,
    severity,
    message: `${label} median ${Math.round(med)} B (warn ${warnAt}, fail ${failAt})`,
    target,
    evidence: {
      median: String(med),
      warnAt: String(warnAt),
      failAt: String(failAt),
      runs: String(reports.length),
    },
  });
}

const summaries = reports.map(resourceSummary);
pushBudget(
  "quality.lighthouse.bytes_total",
  "Total transfer",
  summaries.map((s) => s.total),
  budgets.totalWarn,
  budgets.totalFail
);
pushBudget(
  "quality.lighthouse.bytes_script",
  "Script transfer",
  summaries.map((s) => s.script),
  budgets.scriptWarn,
  budgets.scriptFail
);
pushBudget(
  "quality.lighthouse.bytes_style",
  "Stylesheet transfer",
  summaries.map((s) => s.stylesheet),
  budgets.styleWarn,
  budgets.styleFail
);

console.log(JSON.stringify({ findings }));
