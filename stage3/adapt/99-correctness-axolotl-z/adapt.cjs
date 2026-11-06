"use strict";
// Measurement-only checkpoint. No elementAt source change is accepted here.
// Numeric adaptation directories need an entry point even for a stopped trial.
if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
if (!require("node:fs").statSync(process.argv[2]).isDirectory()) throw new Error("tree is not a directory");
console.log(JSON.stringify({ files: 0, edits: 0, javascriptChanging: false, measurementOnly: true }));
