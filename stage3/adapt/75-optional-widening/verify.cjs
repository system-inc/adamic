#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const { execFileSync } = require('node:child_process');
const [before,after] = process.argv.slice(2);
assert(before && after);
function hashes(root, suffix) {
    const result = {};
    function walk(dir) {
        for (const entry of fs.readdirSync(dir,{withFileTypes:true})) {
            const file=path.join(dir,entry.name);
            if (entry.isDirectory()) walk(file);
            else if (entry.isFile() && file.endsWith(suffix)) result[path.relative(root,file)] = crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
        }
    }
    walk(root);return result;
}
const sourceBefore=hashes(path.join(before,'src'),'.ts');
const sourceAfter=hashes(path.join(after,'src'),'.ts');
assert.deepEqual(Object.keys(sourceBefore).sort(),Object.keys(sourceAfter).sort());
const changes=Object.keys(sourceBefore).filter(k=>sourceBefore[k]!==sourceAfter[k]);
assert.deepEqual(changes,['compiler/types.ts']);
const jsBefore=hashes(path.join(before,'built/local'),'.js');
const jsAfter=hashes(path.join(after,'built/local'),'.js');
assert(Object.keys(jsBefore).length>0);
assert.deepEqual(jsBefore,jsAfter,'emitted JavaScript changed');
const apiAdditions=require('./api-additions.cjs')(fs.readFileSync(path.join(before,'built/local/typescript.d.ts'),'utf8'),fs.readFileSync(path.join(after,'built/local/typescript.d.ts'),'utf8'));
const result=JSON.parse(execFileSync(process.execPath,[path.join(__dirname,'adapt.cjs'),after],{encoding:'utf8'}));
assert.deepEqual(result,{files:0,additions:[]});
assert.deepEqual(hashes(path.join(after,'src'),'.ts'),sourceAfter,'idempotence changed source');
console.log(JSON.stringify({status:'pass',source_files:Object.keys(sourceAfter).length,changed_sources:changes,javascript_files:Object.keys(jsAfter).length,javascript:jsAfter,public_api_additions:apiAdditions,idempotence:result},null,2));
