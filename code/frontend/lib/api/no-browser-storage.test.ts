import assert from "node:assert/strict";
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";
import { test } from "node:test";

const sourceRoots = ["app", "components", "lib"];
const blocked = ["local" + "Storage", "session" + "Storage", "indexed" + "DB", "document" + ".cookie"];
const thisFile = join("lib", "api", "no-browser-storage.test.ts");

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
    .filter((path) => path !== thisFile)
    .filter((path) => /\.(ts|tsx|js|jsx)$/.test(path))
    .flatMap((path) => {
      const text = readFileSync(path, "utf8");
      return blocked.filter((token) => text.includes(token)).map((token) => `${path}: ${token}`);
    });

  assert.deepEqual(matches, []);
});
