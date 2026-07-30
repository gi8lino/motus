import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { serviceWorkerUrl } from "../src/utils/pwa.ts";

test("service worker URL respects deployments under a route prefix", () => {
  assert.equal(serviceWorkerUrl(""), "/sw.js");
  assert.equal(serviceWorkerUrl("/motus/"), "/motus/sw.js");
});

test("service worker refreshes navigations before using the offline shell", async () => {
  const source = await readFile(
    new URL("../public/sw.js", import.meta.url),
    "utf8",
  );
  assert.match(source, /request\.mode === "navigate"/);
  assert.match(source, /fetch\(event\.request\)[\s\S]*catch/);
});
