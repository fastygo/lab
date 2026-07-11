#!/usr/bin/env node
"use strict";

/**
 * Q6 modern extras — viewports 360/768/1280 + console clean.
 */

const { chromium } = require("playwright");

const target = process.env.LAB_TARGET_URL || "http://127.0.0.1/";
const gate = process.env.LAB_GATE_ID || "";
const check = process.env.LAB_CHECK_ID || "";
const VIEWPORTS = [
  { width: 360, height: 640 },
  { width: 768, height: 1024 },
  { width: 1280, height: 800 },
];

function finding(code, severity, message, evidence) {
  const ev = {};
  for (const [k, v] of Object.entries(evidence || {})) {
    ev[k] = v == null ? "" : String(v);
  }
  return { code, gate, check, severity, message, target, evidence: ev };
}

(async () => {
  const browser = await chromium.launch({
    headless: true,
    args: ["--no-sandbox", "--disable-gpu"],
  });
  const findings = [];
  let issues = 0;

  try {
    for (const vp of VIEWPORTS) {
      const page = await browser.newPage({ viewport: vp });
      const consoleErrors = [];
      const pageErrors = [];
      page.on("console", (msg) => {
        if (msg.type() === "error") {
          consoleErrors.push(msg.text());
        }
      });
      page.on("pageerror", (err) => {
        pageErrors.push(String(err && err.message ? err.message : err));
      });

      try {
        await page.goto(target, { waitUntil: "domcontentloaded", timeout: 60000 });
        const mainOk = await page.evaluate(() => {
          const main = document.querySelector("main, #content, [role='main']");
          const body = document.body;
          return {
            hasMain: !!main,
            bodyText: (body && body.innerText ? body.innerText : "").trim().length > 0,
          };
        });
        if (!mainOk.hasMain && !mainOk.bodyText) {
          issues++;
          findings.push(
            finding(
              "quality.viewport.empty",
              "high",
              `Viewport ${vp.width}x${vp.height}: page appears empty`,
              { width: vp.width, height: vp.height }
            )
          );
        } else {
          findings.push(
            finding(
              "quality.viewport.ok",
              "info",
              `Viewport ${vp.width}x${vp.height} rendered`,
              { width: vp.width, height: vp.height, hasMain: mainOk.hasMain }
            )
          );
        }
      } catch (err) {
        issues++;
        findings.push(
          finding(
            "quality.viewport.failed",
            "high",
            `Viewport ${vp.width}x${vp.height}: ${String(err && err.message ? err.message : err).slice(0, 300)}`,
            { width: vp.width, height: vp.height }
          )
        );
      }

      const errs = [...pageErrors, ...consoleErrors].filter(Boolean).slice(0, 10);
      if (errs.length > 0) {
        issues++;
        findings.push(
          finding(
            "quality.console.error",
            "high",
            `Console/page errors at ${vp.width}px: ${errs[0].slice(0, 200)}`,
            { width: vp.width, count: errs.length, sample: errs[0].slice(0, 300) }
          )
        );
      } else {
        findings.push(
          finding("quality.console.ok", "info", `Console clean at ${vp.width}px`, {
            width: vp.width,
          })
        );
      }

      await page.close();
    }

    if (issues === 0) {
      findings.push(
        finding("quality.extras.ok", "info", "Viewports + console clean", {
          viewports: VIEWPORTS.map((v) => v.width).join(","),
        })
      );
    }
    findings.push(
      finding("quality.extras.summary", "info", `Quality extras: ${issues} issue(s)`, {
        issues: String(issues),
      })
    );
  } finally {
    await browser.close();
  }

  console.log(JSON.stringify({ findings }));
})().catch((err) => {
  console.log(
    JSON.stringify({
      findings: [
        finding(
          "quality.extras.exec_failed",
          "high",
          String(err && err.message ? err.message : err).slice(0, 500),
          {}
        ),
      ],
    })
  );
});
