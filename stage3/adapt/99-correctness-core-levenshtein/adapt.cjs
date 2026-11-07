"use strict";
// Report only. The exact proposed initializer pair fails Adamic types.
// No source transformation is implemented.
if (process.argv.length !== 3) throw Error("usage: node adapt.cjs TREE");
console.log(JSON.stringify({files: 0, edits: 0, javascriptChanging: false}));
