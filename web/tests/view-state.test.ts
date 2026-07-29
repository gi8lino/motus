import test from "node:test";
import assert from "node:assert/strict";
import { isValidView } from "../src/hooks/useViewState.ts";

test("view names use the canonical training terminology", () => {
  assert.equal(isValidView("training"), true);
  assert.equal(isValidView("train"), false);
});
