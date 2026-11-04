'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const {compiler}=require('../../47-host-errors/adapt.cjs');
const {validateBindings}=require('../adapt.cjs');
const tree=path.resolve(process.argv[2]),file=path.join(tree,'src/compiler/sys.ts'),saved=fs.readFileSync(file);
try {
    const text=saved.toString();assert.equal(text.split('createHash: createSHA256Hash,').length,2);
    fs.writeFileSync(file,text.replace('createHash: createSHA256Hash,','createHash: _crypto ? createSHA256Hash : generateDjb2Hash,'));
    const p=compiler(tree,ts);
    // Bypass the scoped digest guard, proving the semantic guard itself is red.
    validateBindings(p,p.getSourceFile(file),true);
} finally {fs.writeFileSync(file,saved);assert.deepEqual(fs.readFileSync(file),saved);}
