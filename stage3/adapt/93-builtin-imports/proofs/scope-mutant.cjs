'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const {TestBuiltinSourceScope}=require('./check.cjs');
const before=path.resolve(process.argv[2]),after=path.resolve(process.argv[3]);
const file=path.join(after,'src/compiler/sys.ts'),saved=fs.readFileSync(file);
const inside=process.argv[4]==='inside';
const from=inside?'function getNodeSystem(): System {':'export function getFileWatcherEventKind(';
assert.equal(saved.toString().split(from).length,2,'mutant anchor moved');
const to=inside?from+'/* scope-inside-mutant */':'/* scope-outside-mutant */'+from;
try {
 fs.writeFileSync(file,saved.toString().replace(from,to));
 TestBuiltinSourceScope(before,after);
} finally {
 fs.writeFileSync(file,saved);
 assert.deepEqual(fs.readFileSync(file),saved);
}
