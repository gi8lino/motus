import test from "node:test";
import assert from "node:assert/strict";
import { resolveAuthEntry, resolveAuthView } from "../src/utils/authEntry.ts";

test("fresh proxy installation prioritizes setup over proxy denial", () => {
  assert.equal(
    resolveAuthEntry(
      { setupRequired: true, authHeaderEnabled: true },
      "Forbidden",
    ),
    "setup",
  );
});

test("fresh local installation requires setup", () => {
  assert.equal(
    resolveAuthEntry({ setupRequired: true, authHeaderEnabled: false }, null),
    "setup",
  );
});

test("fresh OIDC and no-auth installations also require setup", () => {
  // Both modes report authHeaderEnabled=false in /api/config.
  assert.equal(
    resolveAuthEntry(
      { setupRequired: true, authHeaderEnabled: false },
      "Unauthorized",
    ),
    "setup",
  );
});

test("completed setup restores proxy denial handling", () => {
  assert.equal(
    resolveAuthEntry(
      { setupRequired: false, authHeaderEnabled: true },
      "Forbidden",
    ),
    "proxy-denied",
  );
});

test("authenticated proxy without errors reaches application", () => {
  assert.equal(
    resolveAuthEntry({ setupRequired: false, authHeaderEnabled: true }, null),
    "ready",
  );
});

test("completed local or OIDC setup reaches regular authentication", () => {
  assert.equal(
    resolveAuthEntry({ setupRequired: false, authHeaderEnabled: false }, null),
    "ready",
  );
});

test("fresh setup overrides a bookmarked training or admin view", () => {
  assert.equal(resolveAuthView("setup", "training"), "login");
  assert.equal(resolveAuthView("setup", "admin"), "login");
});

test("normal navigation is unchanged after setup", () => {
  assert.equal(resolveAuthView("ready", "training"), "training");
  assert.equal(resolveAuthView("proxy-denied", "admin"), "admin");
});
