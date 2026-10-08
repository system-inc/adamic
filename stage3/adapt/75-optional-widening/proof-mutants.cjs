#!/usr/bin/env node
'use strict';
const fs=require('node:fs');
const os=require('node:os');
const path=require('node:path');
const assert=require('node:assert/strict');
const ts=require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
const {spawnSync}=require('node:child_process');
const [before,after]=process.argv.slice(2);
assert(before && after);
const tree=fs.mkdtempSync(path.join(os.tmpdir(),'optional-proof-mutants-'));
const results=[];
try {
    fs.cpSync(path.join(after,'src'),path.join(tree,'src'),{recursive:true});
    fs.cpSync(path.join(after,'built/local'),path.join(tree,'built/local'),{recursive:true});
    fs.symlinkSync(path.join(after,'node_modules'),path.join(tree,'node_modules'));
    function verify(name,expected){
        const result=spawnSync(process.execPath,[path.join(__dirname,'verify.cjs'),before,tree],{encoding:'utf8',env:process.env});
        assert.notEqual(result.status,0,`${name} survived`);
        assert(result.stderr.includes(expected),`${name} failed for a different reason: ${result.stderr}`);
        results.push({name,exit:result.status,caught_by:expected});
    }
    const js=path.join(tree,'built/local/watchGuard.js');
    const original=fs.readFileSync(js);
    fs.appendFileSync(js,'\n// emitted-byte mutant\n');
    verify('javascript-byte-change','emitted JavaScript changed');
    fs.writeFileSync(js,original);
    const file=path.join(tree,'src/compiler/types.ts');
    const text=fs.readFileSync(file,'utf8');
    const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
    let property;
    function visit(n){if(ts.isInterfaceDeclaration(n) && n.name.text==='Type')property=n.members.find(m=>m.name?.getText(source)==='resolvedBaseConstraint');ts.forEachChild(n,visit);}visit(source);
    assert(property?.type);
    fs.writeFileSync(file,text.slice(0,property.type.end)+' | undefined'+text.slice(property.type.end));
    verify('non-idempotent-input',"files: 1");
    console.log(JSON.stringify(results,null,2));
} finally {fs.rmSync(tree,{recursive:true,force:true});}
