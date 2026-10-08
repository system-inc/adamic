const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const { test } = require("node:test");

test("the census counts executable AST sites, not comments or emitted helper text", () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "scout22-census-"));
  try {
    fs.writeFileSync(
      path.join(root, "tsconfig.json"),
      JSON.stringify({
        compilerOptions: { strict: true, types: [], target: "es2020" },
        files: ["sample.ts"],
      }),
    );
    // This is an outside TypeScript-parser input, never an Adamic compilation unit.
    fs.writeFileSync(
      path.join(root, "sample.ts"),
      `
const object = { x: 1 };
for (const key in object) { }
Object.keys(object); Object.entries(object); Object.assign(object, {});
const hasOwnProperty = Object.prototype.hasOwnProperty;
hasOwnProperty.call(object, 'x');
object.toString(); object.valueOf();
\`\${object}\`; '' + object; JSON.stringify(object);
// Object.keys(object); for (const key in object) { }
const ignored = 'Object.keys(object)';
const helper = {text: \`for (var key in object) Object.assign({}, object);\`};
`,
    );
    const actual = spawnSync(
      process.execPath,
      [path.join(__dirname, "census.cjs"), root],
      { encoding: "utf8", env: { ...process.env, CENSUS_TYPE_ROOTS: "" } },
    );
    assert.equal(actual.status, 0, actual.stderr);
    const result = JSON.parse(actual.stdout);
    assert.equal(result.diagnostics, 0);
    assert.deepEqual(result.counts, {
      "for...in": 1,
      "Object.keys": 1,
      "Object.entries": 1,
      "Object.assign": 1,
      "hasOwnProperty member": 1,
      "prototype access": 1,
      "cached hasOwnProperty.call": 1,
      "toString call": 1,
      "valueOf call": 1,
      "template object conversion": 1,
      "+ object conversion": 1,
      "JSON.stringify": 1,
    });
    assert.deepEqual(result.helperCounts, {
      "for...in": 1,
      "Object.assign": 1,
    });
    assert.equal(result.inventory.length, 1);
    assert.match(result.inventory[0].sha256, /^[0-9a-f]{64}$/);
    // An inflated count is rejected by the independently specified expected census.
    const wrong = {
      ...result.counts,
      "Object.keys": result.counts["Object.keys"] + 1,
    };
    assert.throws(
      () => assert.deepEqual(wrong, result.counts),
      assert.AssertionError,
    );
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
});
