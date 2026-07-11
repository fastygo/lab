#!/usr/bin/env node
"use strict";

/**
 * Q6 modern extras — viewports, console, prefers-reduced-motion, broken-link crawl.
 */

const { chromium } = require("playwright");

const target = process.env.LAB_TARGET_URL || "http://127.0.0.1/";
const gate = process.env.LAB_GATE_ID || "";
const check = process.env.LAB_CHECK_ID || "";
let cfg = {};
try {
  cfg = JSON.parse(process.env.LAB_CONFIG_JSON || "{}");
} catch (_) {}

const VIEWPORTS = [
  { width: 360, height: 640 },
  { width: 768, height: 1024 },
  { width: 1280, height: 800 },
];
const linkLimit = Math.max(1, Number(cfg.linkLimit || 40) || 40);

function finding(code, severity, message, evidence) {
  const ev = {};
  for (const [k, v] of Object.entries(evidence || {})) {
    ev[k] = v == null ? "" : String(v);
  }
  return { code, gate, check, severity, message, target, evidence: ev };
}

function sameOrigin(base, href) {
  try {
    const b = new URL(base);
    const u = new URL(href, base);
    if (u.protocol !== "http:" && u.protocol !== "https:") return false;
    if (u.hash && u.pathname === b.pathname && u.search === b.search) return false; // pure hash
    return u.origin === b.origin;
  } catch {
    return false;
  }
}

async function runViewportsAndConsole(browser, findings) {
  let issues = 0;
  for (const vp of VIEWPORTS) {
    const page = await browser.newPage({ viewport: vp });
    const consoleErrors = [];
    const pageErrors = [];
    page.on("console", (msg) => {
      if (msg.type() === "error") consoleErrors.push(msg.text());
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
          finding("quality.viewport.empty", "high", `Viewport ${vp.width}x${vp.height}: page appears empty`, {
            width: vp.width,
            height: vp.height,
          })
        );
      } else {
        findings.push(
          finding("quality.viewport.ok", "info", `Viewport ${vp.width}x${vp.height} rendered`, {
            width: vp.width,
            height: vp.height,
            hasMain: mainOk.hasMain,
          })
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
        finding("quality.console.error", "high", `Console/page errors at ${vp.width}px: ${errs[0].slice(0, 200)}`, {
          width: vp.width,
          count: errs.length,
          sample: errs[0].slice(0, 300),
        })
      );
    } else {
      findings.push(finding("quality.console.ok", "info", `Console clean at ${vp.width}px`, { width: vp.width }));
    }
    await page.close();
  }
  return issues;
}

async function runReducedMotion(browser, findings) {
  const context = await browser.newContext({
    viewport: { width: 1280, height: 800 },
    reducedMotion: "reduce",
  });
  const page = await context.newPage();
  try {
    await page.goto(target, { waitUntil: "domcontentloaded", timeout: 60000 });
    const result = await page.evaluate(() => {
      const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
      const animated = Array.from(
        document.querySelectorAll("[data-lab-animate], .lab-animate, [class*='animate']")
      ).slice(0, 20);
      const probes = animated.map((el) => {
        const cs = getComputedStyle(el);
        return {
          tag: el.tagName,
          animation: cs.animationName || "none",
          duration: cs.animationDuration || "0s",
          iteration: cs.animationIterationCount || "1",
        };
      });
      // Also sample body transitions
      const body = getComputedStyle(document.body);
      return {
        mqMatches: mq.matches,
        probes,
        bodyTransition: body.transitionDuration || "0s",
        hasReduceRule: Array.from(document.styleSheets).some((ss) => {
          try {
            return Array.from(ss.cssRules || []).some(
              (r) => r.media && String(r.media.mediaText || "").includes("prefers-reduced-motion")
            );
          } catch {
            return false;
          }
        }),
      };
    });

    if (!result.mqMatches) {
      findings.push(
        finding("quality.motion.emulation_failed", "medium", "prefers-reduced-motion: reduce not matched by browser", {
          mqMatches: result.mqMatches,
        })
      );
      return 1;
    }

    const stillMoving = (result.probes || []).filter((p) => {
      const dur = parseFloat(p.duration) || 0;
      const infinite = String(p.iteration).includes("infinite");
      return p.animation && p.animation !== "none" && dur > 0 && infinite;
    });

    if (stillMoving.length > 0) {
      findings.push(
        finding(
          "quality.motion.unreduced",
          "high",
          `Infinite animation still active under reduced-motion (${stillMoving[0].animation})`,
          { count: stillMoving.length, sample: stillMoving[0].animation }
        )
      );
      return 1;
    }

    findings.push(
      finding(
        "quality.motion.ok",
        "info",
        result.hasReduceRule
          ? "prefers-reduced-motion respected (media query present)"
          : "prefers-reduced-motion emulated; no infinite animations detected",
        { hasReduceRule: result.hasReduceRule, probes: (result.probes || []).length }
      )
    );
    return 0;
  } catch (err) {
    findings.push(
      finding("quality.motion.failed", "high", String(err && err.message ? err.message : err).slice(0, 400), {})
    );
    return 1;
  } finally {
    await context.close();
  }
}

async function runBrokenLinks(browser, findings) {
  const page = await browser.newPage({ viewport: { width: 1280, height: 800 } });
  try {
    await page.goto(target, { waitUntil: "domcontentloaded", timeout: 60000 });
    const hrefs = await page.$$eval("a[href]", (els) => els.map((a) => a.getAttribute("href") || ""));
    const urls = [];
    const seen = new Set();
    for (const href of hrefs) {
      if (!href || href.startsWith("mailto:") || href.startsWith("tel:") || href.startsWith("javascript:")) continue;
      if (!sameOrigin(target, href)) continue;
      const abs = new URL(href, target).href.replace(/#.*$/, "");
      if (seen.has(abs)) continue;
      seen.add(abs);
      urls.push(abs);
      if (urls.length >= linkLimit) break;
    }

    if (urls.length === 0) {
      findings.push(finding("quality.links.none", "info", "No same-origin links to crawl", { limit: linkLimit }));
      return 0;
    }

    let broken = 0;
    const samples = [];
    for (const u of urls) {
      if (/lab-404|does-not-exist|p=999999/.test(u)) continue;
      try {
        const res = await page.request.get(u, { maxRedirects: 5, timeout: 15000 });
        const status = res.status();
        if (status >= 400) {
          broken++;
          if (samples.length < 8) samples.push(`${status} ${u}`);
          findings.push(
            finding("quality.links.broken", "high", `Broken link HTTP ${status}`, {
              url: u,
              status: String(status),
            })
          );
        }
      } catch (err) {
        broken++;
        if (samples.length < 8) samples.push(`err ${u}`);
        findings.push(
          finding("quality.links.fetch_failed", "high", `Link fetch failed: ${String(err.message || err).slice(0, 120)}`, {
            url: u,
          })
        );
      }
    }

    if (broken === 0) {
      findings.push(
        finding("quality.links.ok", "info", `Broken-link crawl clean (${urls.length} URL(s))`, {
          checked: urls.length,
        })
      );
    }
    findings.push(
      finding("quality.links.summary", "info", `Broken-link crawl: ${broken} broken / ${urls.length} checked`, {
        broken: String(broken),
        checked: String(urls.length),
      })
    );
    return broken > 0 ? 1 : 0;
  } catch (err) {
    findings.push(
      finding("quality.links.failed", "high", String(err && err.message ? err.message : err).slice(0, 400), {})
    );
    return 1;
  } finally {
    await page.close();
  }
}

(async () => {
  const browser = await chromium.launch({
    headless: true,
    args: ["--no-sandbox", "--disable-gpu"],
  });
  const findings = [];
  let issues = 0;

  try {
    issues += await runViewportsAndConsole(browser, findings);
    issues += await runReducedMotion(browser, findings);
    issues += await runBrokenLinks(browser, findings);

    if (issues === 0) {
      findings.push(
        finding("quality.extras.ok", "info", "Viewports, console, reduced-motion, links OK", {
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
        finding("quality.extras.exec_failed", "high", String(err && err.message ? err.message : err).slice(0, 500), {}),
      ],
    })
  );
});
