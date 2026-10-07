"use strict";
// Test-only executable model of Adamic's checked !. Not a source adaptation.
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict");
const [tree] = process.argv.slice(2);
assert(tree, "usage: instrument-range.cjs <disposable-adapted-tree>");
const file = path.join(tree, "src/compiler/core.ts"), source = fs.readFileSync(file, "utf8");
const needle = "to.push(from[i]!)";
assert.equal(source.split(needle).length, 2, "expected exactly one asserted range read");
assert(!source.includes("function parserRangeCheckedRead"), "refusing duplicate instrumentation");
// Leave the push lookup before evaluating the checked argument.
fs.writeFileSync(file, source.replace(needle, "to.push(parserRangeCheckedRead(from[i]))") +
    "\nfunction parserRangeCheckedRead<T>(value: T | undefined): T {\n" +
    "    if (value === undefined) throw new Error(\"parser addRange second read missing\");\n" +
    "    return value;\n}\n");
console.log("Installed test-only checked second read; preserve push lookup order.");
