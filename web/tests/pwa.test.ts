import test from "node:test";
import assert from "node:assert/strict";
import { serviceWorkerUrl } from "../src/utils/pwa.ts";

test("service worker URL respects deployments under a route prefix", () => {
  assert.equal(serviceWorkerUrl(""), "/sw.js");
  assert.equal(serviceWorkerUrl("/motus/"), "/motus/sw.js");
});
