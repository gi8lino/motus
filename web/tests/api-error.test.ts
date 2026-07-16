import assert from "node:assert/strict";
import test from "node:test";

import { ApiError, isApiError } from "../src/utils/apiError.ts";

test("ApiError retains status, backend code, and display message", () => {
  const error = new ApiError(400, "invalid_request", "Step 1 is invalid");

  assert.equal(error.name, "ApiError");
  assert.equal(error.status, 400);
  assert.equal(error.code, "invalid_request");
  assert.equal(error.message, "Step 1 is invalid");
  assert.equal(isApiError(error), true);
  assert.equal(isApiError(new Error("no")), false);
});
