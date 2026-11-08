// The ruled javascript-module-path divergence (#redxes2): a bundled image's __filename and __dirname, printed raw.
// The JavaScript backend's program is an ES module, so its __filename is its own path from import.meta.url,
// lib/_tsc.mjs, where Node's is lib/_tsc.js. Built only as one bundled image.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
console.log(__filename);
console.log(__dirname);
