const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const cp = require("node:child_process");
test("pinned compiler call families, shapes, locations and reserved instance calls", () => {
  const root = path.resolve(
    __dirname,
    "../../../cohere/TypeScript/tsc/testdata/fixtures/compiler",
  );
  const result = JSON.parse(
    cp.execFileSync(
      process.execPath,
      [path.join(__dirname, "census.cjs"), root],
      { maxBuffer: 8 * 1024 * 1024 },
    ),
  );
  const expected = JSON.parse(
    fs.readFileSync(
      path.join(__dirname, "testdata/census.expected.json"),
      "utf8",
    ),
  );
  const actual = Object.fromEntries(
    Object.keys(expected)
      .filter((x) => !["locations", "borrowedLocations"].includes(x))
      .map((x) => [x, result[x]]),
  );
  actual.locations = result.sites.map(({ file, line, column, name }) => ({
    file,
    line,
    column,
    name,
  }));
  actual.borrowedLocations = result.borrowed.map(
    ({ file, line, column, name }) => ({ file, line, column, name }),
  );
  assert.deepEqual(actual, expected);
  // A static call miscount must fail even if the aggregate call total is unchanged.
  const wrong = structuredClone(actual);
  wrong.counts["Math.max"]--;
  wrong.counts["Math.min"]++;
  assert.throws(() => assert.deepEqual(wrong, expected), assert.AssertionError);
});
