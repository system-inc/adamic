"use strict";
const assert=require("node:assert/strict"),path=require("node:path");
assert.equal(process.argv.length,3,"usage: node mutant.cjs TREE");
console.log(JSON.stringify(require("./adapt.cjs").rewrite(path.resolve(process.argv[2]),true)));
