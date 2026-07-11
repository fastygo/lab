#!/usr/bin/env node
"use strict";

/**
 * Q3 CSS lint — parse errors + forbidden expression()/behavior:.
 * Scans LAB_CSS_DIR and/or unzipped LAB_THEME_ZIP.
 */

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");
const stylelint = require("stylelint");

const gate = process.env.LAB_GATE_ID || "";
const check = process.env.LAB_CHECK_ID || "";
const target = process.env.LAB_TARGET_URL || "";
const cssDir = process.env.LAB_CSS_DIR || "";
const themeZip = process.env.LAB_THEME_ZIP || "";

function finding(code, severity, message, evidence) {
  const ev = {};
  for (const [k, v] of Object.entries(evidence || {})) {
    ev[k] = v == null ? "" : String(v);
  }
  return { code, gate, check, severity, message, target, evidence: ev };
}

function collectCss(root, out) {
  if (!root || !fs.existsSync(root)) return;
  const entries = fs.readdirSync(root, { withFileTypes: true });
  for (const ent of entries) {
    const full = path.join(root, ent.name);
    if (ent.isDirectory()) {
      if (ent.name === "node_modules" || ent.name === "vendor" || ent.name === ".git") continue;
      collectCss(full, out);
      continue;
    }
    if (/\.(css|scss)$/i.test(ent.name)) {
      out.push(full);
    }
  }
}

function unzipTheme(zipPath) {
  const dest = "/tmp/lab-theme-css";
  fs.rmSync(dest, { recursive: true, force: true });
  fs.mkdirSync(dest, { recursive: true });
  try {
    execFileSync("unzip", ["-q", "-o", zipPath, "-d", dest], { stdio: "ignore" });
  } catch {
    // busybox unzip may differ; try python
    execFileSync(
      "python3",
      ["-c", "import zipfile,sys; zipfile.ZipFile(sys.argv[1]).extractall(sys.argv[2])", zipPath, dest],
      { stdio: "ignore" }
    );
  }
  return dest;
}

function scanForbidden(source, file) {
  const hits = [];
  if (/expression\s*\(/i.test(source)) {
    hits.push({ file, kind: "expression()" });
  }
  if (/behavior\s*:/i.test(source)) {
    hits.push({ file, kind: "behavior:" });
  }
  return hits;
}

(async () => {
  const roots = [];
  if (cssDir) roots.push(cssDir);
  if (themeZip && fs.existsSync(themeZip)) {
    roots.push(unzipTheme(themeZip));
  }

  const files = [];
  for (const root of roots) collectCss(root, files);

  if (files.length === 0) {
    console.log(
      JSON.stringify({
        findings: [
          finding("quality.css.no_files", "medium", "No CSS files found to lint", {
            cssDir,
            themeZip: themeZip ? "set" : "",
          }),
          finding("quality.css.summary", "info", "CSS lint: 0 file(s), 0 issue(s)", {
            files: "0",
            issues: "0",
          }),
        ],
      })
    );
    return;
  }

  const findings = [];
  let issues = 0;

  for (const file of files) {
    const source = fs.readFileSync(file, "utf8");
    for (const hit of scanForbidden(source, file)) {
      issues++;
      findings.push(
        finding("quality.css.forbidden", "high", `Forbidden CSS ${hit.kind} in ${path.basename(file)}`, {
          file,
          kind: hit.kind,
        })
      );
    }

    const result = await stylelint.lint({
      code: source,
      codeFilename: file,
      configFile: "/runner/stylelint.config.mjs",
    });
    for (const r of result.results || []) {
      for (const w of r.warnings || []) {
        if (w.rule === "CssSyntaxError" || /CssSyntaxError/i.test(w.text || "") || w.severity === "error") {
          // With empty rules, stylelint still reports parse errors as CssSyntaxError
          if (!/CssSyntaxError/i.test(String(w.rule || "")) && !/CssSyntaxError/i.test(w.text || "")) {
            continue;
          }
          issues++;
          findings.push(
            finding("quality.css.parse_error", "high", w.text || "CSS parse error", {
              file,
              line: String(w.line || ""),
              column: String(w.column || ""),
            })
          );
        }
      }
      // parseErrors field on some versions
      for (const pe of r.parseErrors || []) {
        issues++;
        findings.push(
          finding("quality.css.parse_error", "high", pe.text || "CSS parse error", {
            file,
            line: String(pe.line || ""),
          })
        );
      }
    }
  }

  if (issues === 0) {
    findings.push(
      finding("quality.css.ok", "info", `CSS parse clean (${files.length} file(s))`, {
        files: String(files.length),
      })
    );
  }
  findings.push(
    finding("quality.css.summary", "info", `CSS lint: ${files.length} file(s), ${issues} issue(s)`, {
      files: String(files.length),
      issues: String(issues),
    })
  );

  console.log(JSON.stringify({ findings }));
})().catch((err) => {
  console.log(
    JSON.stringify({
      findings: [
        finding("quality.css.exec_failed", "high", String(err && err.message ? err.message : err).slice(0, 500), {}),
      ],
    })
  );
});
