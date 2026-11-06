const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict');
const ts=require(path.resolve(__dirname,'../../../../api/node_modules/typescript'));
const repo=path.resolve(__dirname,'../../../../..'), tree=process.argv[2];
assert(tree,'provide the baseline full apply tree');
const rules=require(repo+'/stage3/adapt/40-explicit-any/class-rules.json')['enum-display'];
const baseline=JSON.parse(cp.execFileSync('git',['show','63e14a0b2b61cf251ad74292e42d5fd09af804b3:stage3/adapt/40-explicit-any/class-rules.json'],{cwd:repo,encoding:'utf8'}))['enum-display'];
const {planEnum}=require(repo+'/stage3/adapt/40-explicit-any/classes.cjs');
const results=[];
for(const file of [...new Set(rules.map(r=>r.file))]) {
 const before=cp.execFileSync('git',['-C',tree,'show','refs/stage3/40-explicit-any-before:'+file],{encoding:'utf8'});
 const actual=planEnum(before,file);
 assert.equal(actual.removed,rules.filter(r=>r.file===file).length);
 const replay=planEnum(actual.text,file);assert.equal(replay.removed,0);assert.equal(replay.text,actual.text);
 let old=before;
 for(const r of baseline.filter(r=>r.file===file)) old=old.replace(r.originalLine,r.originalLine.slice(0,r.column-1)+r.replacement+r.originalLine.slice(r.column+2));
 for(const removeComments of [false,true]) {
 const opts={compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
 assert.equal(ts.transpileModule(old,opts).outputText,ts.transpileModule(actual.text,opts).outputText);
 }
 const first=rules.find(r=>r.file===file);
 assert.throws(()=>planEnum(before.replace(first.originalLine,first.originalLine.replace('any','boolean')),file),/unreviewed enum owner/);
 results.push({file,applied:actual.removed,replay_edits:replay.removed,javascript_identical_with_comments_and_without:true,foreign_type_mutant_refused:true});
}
console.log(JSON.stringify({status:'pass',results},null,2));
