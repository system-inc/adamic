'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const {TestCryptoSourceScope,TestCryptoTranspileScope}=require('./check.cjs');
const before=path.resolve(process.argv[2]),after=path.resolve(process.argv[3]);
const file=path.join(after,'src/compiler/sys.ts'),saved=fs.readFileSync(file);
const targets=[
    ['declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;', 'declare function clearTimeout(handle: NodeJS.Timeout | undefined | null): void;'],
    ['clearTimeout?(timeoutId: unknown): void;', 'clearTimeout?(timeoutId: unknown | null): void;'],
].filter(([from])=>saved.toString().includes(from));
assert.equal(targets.length,1,'one current erased timer type target');
const [from,to]=targets[0];assert.equal(saved.toString().split(from).length,2);
try {
    fs.writeFileSync(file,saved.toString().replace(from,to));
    TestCryptoTranspileScope(before,after);
    console.log('PASS TestCryptoTranspileScope: the extra type-only edit is invisible in both JavaScript comparisons');
    TestCryptoSourceScope(before,after);
} finally {fs.writeFileSync(file,saved);assert.deepEqual(fs.readFileSync(file),saved);}
