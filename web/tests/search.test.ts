import assert from "node:assert/strict";
import test from "node:test";

import { matchesSearch } from "../src/utils/search.ts";

test("exercise search matches names and labels case-insensitively", () => {
  const item = {
    label: "Swing",
    searchText: "kettlebell hinge conditioning",
  };

  assert.equal(matchesSearch(item, "swi"), true);
  assert.equal(matchesSearch(item, "KETTLEBELL"), true);
  assert.equal(matchesSearch(item, "hinge"), true);
  assert.equal(matchesSearch(item, "push"), false);
});
