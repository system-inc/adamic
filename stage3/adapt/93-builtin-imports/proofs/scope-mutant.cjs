'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const {TestBuiltinSourceScope,TestBuiltinTranspileScope}=require('./check.cjs');
const before=path.resolve(process.argv[2]),after=path.resolve(process.argv[3]);
const file=path.join(after,'src/compiler/sys.ts'),saved=fs.readFileSync(file);
const from='declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;';
assert.equal(saved.toString().split(from).length,2);
try {
 fs.writeFileSync(file,saved.toString().replace(from,'declare function clearTimeout(handle: NodeJS.Timeout | undefined | null): void;'));
 TestBuiltinTranspileScope(before,after);
 console.log('PASS TestBuiltinTranspileScope: the extra type-only edit is invisible in both JavaScript comparisons');
 TestBuiltinSourceScope(before,after);
} finally {
 fs.writeFileSync(file,saved);
 assert.deepEqual(fs.readFileSync(file),saved);
}
