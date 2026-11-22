const assert = require("node:assert/strict");
assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
process.argv.push(require.resolve(process.env.CENSUS_TYPESCRIPT || "typescript"));
require("./apply-family20.cjs");
