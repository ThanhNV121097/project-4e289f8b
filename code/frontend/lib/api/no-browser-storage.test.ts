import assert from "node:assert/strict";
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";
import { test } from "node:test";

const sourceRoots = ["app", "components", "lib"];
const blocked = ["localStorage", "sessionStorage", "indexedDB", "document.cookie"];

function files(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name);
    const stat = statSync(path);
    return stat.isDirectory() ? files(path) : [path];
  });
}

test("task board frontend does not persist tasks in browser storage", () => {
  const matches = sourceRoots
    .flatMap(files)
    .filter((path) => /\.(ts|tsx|js|jsx)$/.test(path))
    .flatMap((path) => {
      const text = readFileSync(path, "utf8");
      return blocked.filter((token) => text.includes(token)).map((token) => `${path}: ${token}`);
    });

  assert.deepEqual(matches, []);
});
