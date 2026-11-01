#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const {adapt}=require('./adapt.cjs');
const tree=path.resolve(process.argv[2]),file=path.join(tree,'src/compiler/checker.ts');
const original=fs.readFileSync(file,'utf8'),hash=s=>crypto.createHash('sha256').update(s).digest('hex');
try {
    const unrelated=original+'\r\n// 95b unrelated-owner isolation control\r\n';
    assert.notEqual(hash(unrelated),hash(original));fs.writeFileSync(file,unrelated);
    const first=adapt(tree),second=adapt(tree);
    assert.equal(first.assertions,8);assert.equal(second.assertions,0);
    const valid=fs.readFileSync(file,'utf8');
    const old='{ kind: TypeMapKind.Deferred, sources, targets }';
    assert.equal(valid.split(old).length-1,1);
    const broken=valid.replace(old,'{ kind: TypeMapKind.Deferred, sources, targets: targets.slice(1) }');
    fs.writeFileSync(file,broken);
    assert.throws(()=>adapt(tree),/caller\/arity dependency moved: makeDeferredTypeMapper/);
    assert.equal(fs.readFileSync(file,'utf8'),broken,'failed caller guard must write nothing');
    console.log(JSON.stringify({unrelatedWholeFileChangeAccepted:{beforeSha256:hash(original),unrelatedSha256:hash(unrelated),first,second},brokenMapperFactory:{mutation:'drop first target callback',caught:true,catcher:'caller/arity dependency moved: makeDeferredTypeMapper',writes:0}},null,2));
} finally {fs.writeFileSync(file,original);}
