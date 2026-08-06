import assert from "node:assert/strict";
import test from "node:test";
import { isThematicTitle } from "./release-title.js";

test("version-string titles are patch releases (compact hero)", () => {
  assert.equal(isThematicTitle("Inx 1.17.14", "1.17.14"), false);
  assert.equal(isThematicTitle("Inx v1.17.14", "1.17.14"), false);
  assert.equal(isThematicTitle("  inx   1.17.14  ", "1.17.14"), false);
});

test("titles that merely start with the version stay thematic", () => {
  assert.equal(isThematicTitle("Inx 1.18.0 — Better agents", "1.18.0"), true);
  assert.equal(
    isThematicTitle("A safer delivery loop, trusted MCP execution, and offline recovery", "1.17.13"),
    true,
  );
});

test("edge cases stay thematic (full hero is the safe default)", () => {
  assert.equal(isThematicTitle("", "1.17.14"), true);
  assert.equal(isThematicTitle("Inx 1.17.14", ""), true);
});
