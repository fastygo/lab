#!/usr/bin/env node
"use strict";

const { chromium } = require("playwright");
const { AxeBuilder } = require("@axe-core/playwright");

const target = process.env.LAB_TARGET_URL || "http://127.0.0.1/";
const gate = process.env.LAB_GATE_ID || "";
const check = process.env.LAB_CHECK_ID || "";

(async () => {
  const browser = await chromium.launch({
    headless: true,
    args: ["--no-sandbox", "--disable-gpu"],
  });
  try {
    const context = await browser.newContext();
    const page = await context.newPage();
    await page.goto(target, { waitUntil: "networkidle", timeout: 60000 });
    const results = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"])
      .analyze();

    const findings = [];
    for (const v of results.violations || []) {
      const impact = v.impact || "moderate";
      let severity = "low";
      if (impact === "critical" || impact === "serious") severity = "high";
      else if (impact === "moderate") severity = "medium";
      findings.push({
        code: `quality.axe.${v.id}`,
        gate,
        check,
        severity,
        message: `${v.help} (${v.nodes?.length || 0} nodes)`,
        target,
        evidence: { impact, description: v.description || "" },
      });
    }
    if (findings.length === 0) {
      findings.push({
        code: "quality.axe.ok",
        gate,
        check,
        severity: "info",
        message: "no critical/serious axe violations",
        target,
      });
    }
    console.log(JSON.stringify({ findings }));
  } finally {
    await browser.close();
  }
})().catch(async (err) => {
  console.log(
    JSON.stringify({
      findings: [
        {
          code: "quality.axe.exec_failed",
          gate,
          check,
          severity: "high",
          message: String(err).slice(0, 500),
          target,
        },
      ],
    })
  );
  process.exit(0);
});
