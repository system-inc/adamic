#!/usr/bin/env node
'use strict';
const fs=require('node:fs');
const path=require('node:path');
const assert=require('node:assert/strict');
const {execFileSync}=require('node:child_process');
const [pristine,before,after,optional,readonly,mode]=process.argv.slice(2);
assert(pristine && before && after && optional && readonly && (!mode || mode==='--accept-api'));
const prerequisite=JSON.parse(execFileSync(process.execPath,[path.join(__dirname,'../30-indexed-reads/check-api-baselines.cjs'),pristine,before,optional],{
    encoding:'utf8',env:{...process.env,TSC_ADAPT_READONLY_REPORT:readonly},
}));
assert.equal(prerequisite.adaptation20_lines.length,189);
assert.equal(prerequisite.adaptation40_lines.length,28);
assert.equal(prerequisite.adaptation70_lines.length,1);
const base=fs.readFileSync(path.join(before,'built/local/typescript.d.ts'),'utf8');
const actual=fs.readFileSync(path.join(after,'built/local/typescript.d.ts'),'utf8');
const additions=require('./api-additions.cjs')(base,actual);
assert.equal(additions.length,1);
function baselines(root){const map=new Map();function walk(dir){for(const e of fs.readdirSync(dir,{withFileTypes:true})){const file=path.join(dir,e.name);if(e.isDirectory())walk(file);else if(e.isFile())map.set(path.relative(root,file),fs.readFileSync(file));}}walk(root);return map;}
const original=baselines(path.join(pristine,'tests/baselines/reference'));
const adapted=baselines(path.join(after,'tests/baselines/reference'));
assert.deepEqual([...original.keys()].sort(),[...adapted.keys()].sort());
for(const [name,bytes] of adapted) {
    if(name==='api/typescript.d.ts')assert([original.get(name),Buffer.from(base),Buffer.from(actual)].some(allowed=>allowed.equals(bytes)),'unverified API reference edit');
    else assert(bytes.equals(original.get(name)),`unsanctioned reference baseline ${name}`);
}
if(mode==='--accept-api')fs.writeFileSync(path.join(after,'tests/baselines/reference/api/typescript.d.ts'),actual);
console.log(JSON.stringify({status:'pass',prerequisite_lines:{optional:189,adaptation40:28,readonly:1},additions,other_reference_baselines_identical:adapted.size-1,accepted:mode==='--accept-api'},null,2));
