// The ruled installed-layout-path divergence (#redxes2): a CommonJS build's __filename and __dirname, printed raw.
// Natively __filename is the running image's real path, lib/_tsc, where Node's is lib/_tsc.js. tsc never prints
// either; TestInstalledLayoutReadsAreNodes holds what it derives from them. Built only as a CommonJS program.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
console.log(__filename);
console.log(__dirname);
