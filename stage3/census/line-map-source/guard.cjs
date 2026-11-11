'use strict';
const assert = require('node:assert/strict'), cp = require('node:child_process'), path = require('node:path'), fs = require('node:fs');
const ts = require('/workspace/adamic/stage3/api/node_modules/typescript');
const { plan } = require('/workspace/adamic/stage3/adapt/71-writable-views/line-map-source.cjs');
const pins = require('/workspace/adamic/stage3/adapt/71-writable-views/line-map-source.json');
const tree = '/virtual/line-map-source';
const types = path.join(tree,'src/compiler/types.ts'), scanner = path.join(tree,'src/compiler/scanner.ts');
const before = new Map([types,scanner].map(name=>[name,cp.execFileSync('git',['-C','/workspace/scratch/line-map-before','show','refs/stage3/71-writable-views-after:'+path.relative(tree,name)],{encoding:'utf8',maxBuffer:8*1024*1024})]));
const changes = plan(tree,ts,before);assert.equal(changes.size,2);
const after = new Map([...before,...changes]);assert.equal(plan(tree,ts,after).size,0);
const mutants = [
 ['missing interface',types,after.get(types).replace(pins.lineMapSource+'\r\n\r\n','')],
 ['missing import',scanner,after.get(scanner).replace('type LineAndCharacter, type LineMapSource,','type LineAndCharacter,')],
 ['old getLineStarts',scanner,after.get(scanner).replace(pins.starts.after,pins.starts.before)],
 ['old overload set',scanner,after.get(scanner).replace(pins.position.after,pins.position.before)],
 ['public overload first',scanner,after.get(scanner).replace(pins.position.after,pins.position.after.split('\r\n').slice(2,3).concat(pins.position.after.split('\r\n').slice(0,2),pins.position.after.split('\r\n').slice(3)).join('\r\n'))],
 ['missing internal overload',scanner,after.get(scanner).replace('/** @internal */\r\nexport function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;\r\n','')],
 ['changed cache body',scanner,after.get(scanner).replace('return sourceFile.lineMap || (sourceFile.lineMap = computeLineStarts(sourceFile.text));','return computeLineStarts(sourceFile.text);')],
 ['changed position body',scanner,after.get(scanner).replace('return computeLineAndCharacterOfPosition(getLineStarts(sourceFile), position);','return computeLineAndCharacterOfPosition(getLineStarts(sourceFile), position + 1);')],
 ['changed SourceFileLike documentation',types,after.get(types).replace('Subset of properties from SourceFile that are used in multiple utility functions','Changed documentation')],
 ['changed LineMapSource',types,after.get(types).replace(pins.lineMapSource,pins.lineMapSource.replace('readonly text: string;','text: string;'))],
 ['displaced SourceFileLike documentation',types,after.get(types).replace(pins.lineMapSource+'\r\n\r\n'+pins.sourceFileLike,pins.sourceFileLike.slice(0,pins.sourceFileLike.indexOf('export interface'))+pins.lineMapSource+'\r\n\r\n'+pins.sourceFileLike.slice(pins.sourceFileLike.indexOf('export interface')))],
];
const results = [];
for (const [name,file,text] of mutants) {
 assert.notEqual(text,after.get(file),name+' did not mutate');
 const candidate = new Map(after);candidate.set(file,text);let caught;
 try { plan(tree,ts,candidate); } catch (e) { caught=e.message; }
 assert(caught,name+' accepted');assert.equal(candidate.get(file),text);
 results.push({name,rejected:true,reason:caught,pure_plan_no_writes:true});
}
fs.writeFileSync(path.join(__dirname,'guard-results.json'),JSON.stringify({idempotent:true,changed_files:changes.size,mutants:results},null,2)+'\n');
console.log(JSON.stringify({idempotent:true,mutants_rejected:results.length}));
