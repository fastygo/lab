#!/usr/bin/env node
"use strict";

/**
 * Org Gate 4 — Playwright keyboard smoke:
 * skip-link, primary nav, mobile sheet, search.
 */

const { chromium } = require("playwright");

const target = process.env.LAB_TARGET_URL || "http://127.0.0.1/";
const gate = process.env.LAB_GATE_ID || "";
const check = process.env.LAB_CHECK_ID || "";

function finding(code, severity, message, evidence) {
  return {
    code,
    gate,
    check,
    severity,
    message,
    target,
    evidence: evidence || {},
  };
}

function stringifyEvidence(obj) {
  const out = {};
  for (const [k, v] of Object.entries(obj || {})) {
    out[k] = v == null ? "" : String(v);
  }
  return out;
}

async function tabUntil(page, predicate, maxTabs) {
  for (let i = 1; i <= maxTabs; i++) {
    await page.keyboard.press("Tab");
    const hit = await predicate(i);
    if (hit) return { ok: true, tabs: i, ...hit };
  }
  return { ok: false, tabs: maxTabs };
}

async function activeInfo(page) {
  return page.evaluate(() => {
    const el = document.activeElement;
    if (!el) return { tag: "", id: "", name: "", className: "", role: "", href: "" };
    return {
      tag: el.tagName.toLowerCase(),
      id: el.id || "",
      name: el.getAttribute("name") || "",
      className: typeof el.className === "string" ? el.className : "",
      role: el.getAttribute("role") || "",
      href: el.getAttribute("href") || "",
    };
  });
}

async function resetFocus(page) {
  await page.evaluate(() => {
    const active = document.activeElement;
    if (active && typeof active.blur === "function") active.blur();
    const body = document.body;
    if (body) {
      body.setAttribute("tabindex", "-1");
      body.focus();
      body.removeAttribute("tabindex");
    }
  });
}

async function runSkip(page) {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto(target, { waitUntil: "domcontentloaded", timeout: 60000 });
  await resetFocus(page);

  const first = await tabUntil(
    page,
    async () => {
      const info = await activeInfo(page);
      const isSkip =
        info.className.includes("skip-link") ||
        (info.href || "").includes("#content");
      return isSkip ? { info } : null;
    },
    8
  );

  if (!first.ok) {
    return [
      finding(
        "org.keyboard.skip_missing",
        "high",
        "Skip link was not the first focusable control within 8 Tabs",
        stringifyEvidence({ tabs: first.tabs })
      ),
    ];
  }

  await page.keyboard.press("Enter");
  await page.waitForTimeout(200);
  const after = await page.evaluate(() => {
    const hash = location.hash || "";
    const main = document.getElementById("content");
    const active = document.activeElement;
    const activeIsContent =
      !!main &&
      !!active &&
      (active === main || main.contains(active) || active.id === "content");
    return {
      hash,
      hasContent: !!main,
      activeIsContent,
      activeId: active && active.id ? active.id : "",
    };
  });

  if (!after.hasContent) {
    return [
      finding(
        "org.keyboard.skip_target_missing",
        "high",
        "Skip link activated but #content target is missing",
        stringifyEvidence(after)
      ),
    ];
  }

  const hashOk = after.hash === "#content";
  if (!hashOk && !after.activeIsContent) {
    return [
      finding(
        "org.keyboard.skip_failed",
        "high",
        "Skip link Enter did not reach #content (hash/focus)",
        stringifyEvidence(after)
      ),
    ];
  }

  return [
    finding(
      "org.keyboard.skip_ok",
      "info",
      "Skip link is first focusable and targets #content",
      stringifyEvidence({ tabs: first.tabs, hash: after.hash, activeId: after.activeId })
    ),
  ];
}

async function runNav(page) {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto(target, { waitUntil: "domcontentloaded", timeout: 60000 });
  await resetFocus(page);

  const navMeta = await page.evaluate(() => {
    const nav = document.querySelector("header nav");
    if (!nav) return { count: 0, display: "", visible: false };
    const style = getComputedStyle(nav);
    const visible = style.display !== "none" && style.visibility !== "hidden";
    return {
      count: nav.querySelectorAll("a[href]").length,
      display: style.display,
      visible,
    };
  });

  if (navMeta.count === 0) {
    return [
      finding(
        "org.keyboard.nav_missing",
        "high",
        "No primary desktop nav links (assign a menu to location primary)",
        stringifyEvidence({ viewport: "1280x800" })
      ),
    ];
  }

  if (!navMeta.visible) {
    return [
      finding(
        "org.keyboard.nav_hidden",
        "high",
        "Desktop primary nav is in the DOM but not visible at 1280px (check md:block CSS)",
        stringifyEvidence(navMeta)
      ),
    ];
  }

  const hit = await tabUntil(
    page,
    async () => {
      return page.evaluate(() => {
        const el = document.activeElement;
        if (!el) return null;
        const nav = el.closest("header nav");
        if (!nav) return null;
        return {
          id: el.id || "",
          text: (el.textContent || "").trim().slice(0, 80),
          href: el.getAttribute("href") || "",
        };
      });
    },
    40
  );

  if (!hit.ok) {
    return [
      finding(
        "org.keyboard.nav_unreachable",
        "high",
        "Primary nav links not reachable via Tab within 40 presses",
        stringifyEvidence({ tabs: hit.tabs, navCount: navMeta.count })
      ),
    ];
  }

  return [
    finding(
      "org.keyboard.nav_ok",
      "info",
      "Primary nav reachable via keyboard",
      stringifyEvidence({ tabs: hit.tabs, href: hit.href || "", text: hit.text || "" })
    ),
  ];
}

async function runSheet(page) {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(target, { waitUntil: "domcontentloaded", timeout: 60000 });

  const trigger = page.locator(
    '#classic-mobile-sheet-trigger, #app-shell-mobile-sheet-trigger, [data-ui8kit-dialog-open="true"]'
  ).first();
  if ((await trigger.count()) === 0) {
    return [
      finding(
        "org.keyboard.sheet_missing",
        "high",
        "Mobile sheet trigger not found (primary menu empty?)",
        stringifyEvidence({ viewport: "390x844" })
      ),
    ];
  }

  await trigger.focus();
  await trigger.click();
  await page.waitForTimeout(250);

  const openState = await page.evaluate(() => {
    const panel = document.querySelector(
      '#classic-mobile-sheet-panel, #app-shell-mobile-sheet-panel, [data-ui8kit-dialog="true"]'
    );
    const triggerEl =
      document.querySelector("#classic-mobile-sheet-trigger") ||
      document.querySelector("#app-shell-mobile-sheet-trigger") ||
      document.querySelector('[data-ui8kit-dialog-open="true"]');
    if (!panel) return { found: false };
    return {
      found: true,
      hidden: panel.hasAttribute("hidden"),
      state: panel.getAttribute("data-state") || "",
      ariaExpanded: triggerEl ? triggerEl.getAttribute("aria-expanded") || "" : "",
      role: panel.getAttribute("role") || "",
    };
  });

  if (!openState.found || openState.hidden || openState.state === "closed") {
    return [
      finding(
        "org.keyboard.sheet_open_failed",
        "high",
        "Mobile sheet did not open",
        stringifyEvidence(openState)
      ),
    ];
  }

  if (openState.ariaExpanded && openState.ariaExpanded !== "true") {
    return [
      finding(
        "org.keyboard.sheet_aria_failed",
        "medium",
        "Mobile sheet open but aria-expanded is not true",
        stringifyEvidence(openState)
      ),
    ];
  }

  await page.keyboard.press("Escape");
  await page.waitForTimeout(250);

  const closedState = await page.evaluate(() => {
    const panel = document.querySelector(
      '#classic-mobile-sheet-panel, #app-shell-mobile-sheet-panel, [data-ui8kit-dialog="true"]'
    );
    const triggerEl =
      document.querySelector("#classic-mobile-sheet-trigger") ||
      document.querySelector("#app-shell-mobile-sheet-trigger") ||
      document.querySelector('[data-ui8kit-dialog-open="true"]');
    if (!panel) return { found: false };
    return {
      found: true,
      hidden: panel.hasAttribute("hidden"),
      state: panel.getAttribute("data-state") || "",
      ariaExpanded: triggerEl ? triggerEl.getAttribute("aria-expanded") || "" : "",
    };
  });

  const closed =
    closedState.hidden ||
    closedState.state === "closed" ||
    closedState.ariaExpanded === "false";
  if (!closed) {
    return [
      finding(
        "org.keyboard.sheet_close_failed",
        "high",
        "Mobile sheet did not close on Escape",
        stringifyEvidence(closedState)
      ),
    ];
  }

  return [
    finding(
      "org.keyboard.sheet_ok",
      "info",
      "Mobile sheet opens, exposes aria state, closes on Escape",
      stringifyEvidence({ openAria: openState.ariaExpanded, closedState: closedState.state || "hidden" })
    ),
  ];
}

async function runSearch(page) {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto(target, { waitUntil: "domcontentloaded", timeout: 60000 });
  await resetFocus(page);

  const searchCount = await page.locator('input[name="s"], #latte-search-field').count();
  if (searchCount === 0) {
    return [
      finding(
        "org.keyboard.search_missing",
        "high",
        "Search field input[name=s] not found",
        stringifyEvidence({})
      ),
    ];
  }

  const hit = await tabUntil(
    page,
    async () => {
      const info = await activeInfo(page);
      if (info.name === "s" || info.id === "latte-search-field") {
        return { info };
      }
      return null;
    },
    50
  );

  if (!hit.ok) {
    return [
      finding(
        "org.keyboard.search_unreachable",
        "high",
        "Search field not reachable via Tab within 50 presses",
        stringifyEvidence({ tabs: hit.tabs, searchCount })
      ),
    ];
  }

  return [
    finding(
      "org.keyboard.search_ok",
      "info",
      "Search field reachable via keyboard",
      stringifyEvidence({ tabs: hit.tabs, id: hit.info.id || "", name: hit.info.name || "" })
    ),
  ];
}

(async () => {
  const browser = await chromium.launch({
    headless: true,
    args: ["--no-sandbox", "--disable-gpu"],
  });
  const findings = [];
  try {
    const page = await browser.newPage();
    findings.push(...(await runSkip(page)));
    findings.push(...(await runNav(page)));
    findings.push(...(await runSheet(page)));
    findings.push(...(await runSearch(page)));

    const fails = findings.filter((f) => f.severity === "high" || f.severity === "medium");
    if (fails.length === 0) {
      findings.push(
        finding(
          "org.keyboard.ok",
          "info",
          "Keyboard scenarios passed (skip, nav, sheet, search)",
          stringifyEvidence({ scenarios: "4" })
        )
      );
    }
    findings.push(
      finding(
        "org.keyboard.summary",
        "info",
        `Keyboard scenarios: ${fails.length} issue(s)`,
        stringifyEvidence({ issues: fails.length })
      )
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
          "org.keyboard.exec_failed",
          "high",
          String(err && err.message ? err.message : err).slice(0, 500),
          stringifyEvidence({})
        ),
      ],
    })
  );
});
