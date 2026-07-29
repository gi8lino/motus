import test from "node:test";
import assert from "node:assert/strict";
import { resolveThemeMode } from "../src/utils/themeMode.ts";

test("theme controller resolves explicit and system modes", () => {
  assert.equal(resolveThemeMode("auto", true), "dark");
  assert.equal(resolveThemeMode("auto", false), "light");
  assert.equal(resolveThemeMode("light", true), "light");
});
