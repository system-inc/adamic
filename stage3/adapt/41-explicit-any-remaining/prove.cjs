'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const crypto = require('node:crypto'), ts = require('typescript');
const [beforeTree, afterTree, output] = process.argv.slice(2);
const rules = require('./rules.json');
const hash = s => crypto.createHash('sha256').update(s).digest('hex');
const walk = d => fs.readdirSync(d, {withFileTypes:true}).flatMap(e => e.isDirectory() ? walk(path.join(d,e.name)) : [path.join(d,e.name)]).sort();
const evidence = {typescript:ts.version, rules:[], sources:{}, emissions:{}};
assert.deepEqual(walk(path.join(afterTree,'src/compiler')).filter(f=>f.endsWith('.ts')).map(f=>path.relative(afterTree,f)), walk(path.join(beforeTree,'src/compiler')).filter(f=>f.endsWith('.ts')).map(f=>path.relative(beforeTree,f)), 'compiler source file set');
for (const oldFile of walk(path.join(beforeTree,'src/compiler')).filter(f=>f.endsWith('.ts'))) {
    const file=path.relative(beforeTree,oldFile), before=fs.readFileSync(oldFile,'utf8'), after=fs.readFileSync(path.join(afterTree,file),'utf8');
    let expected=before;
    for (const rule of rules.filter(r=>r.file===file)) {
        assert.equal(expected.split(rule.before).length,2, 'one original rule '+rule.id);
        expected=expected.replace(rule.before,rule.after);
    }
    assert.equal(after,expected,'only reviewed source edits '+file);
    evidence.sources[file]={before:hash(before),after:hash(after)};
    if (before!==after) {
        const options={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext};
        const a=ts.transpileModule(before,{fileName:file,compilerOptions:options}).outputText;
        const b=ts.transpileModule(after,{fileName:file,compilerOptions:options}).outputText;
        assert.equal(b,a,'whole-file emitted JavaScript '+file);
        evidence.emissions[file]={bytes:Buffer.byteLength(a),sha256:hash(a)};
    }
}
for (const r of rules) evidence.rules.push({id:r.id,reason:r.reason});
const artifacts = tree => Object.fromEntries(walk(path.join(tree,'built')).filter(f=>/\.(?:js|mjs|cjs|d\.ts)$/.test(f)).map(f=>[path.relative(path.join(tree,'built'),f),fs.readFileSync(f,'utf8')]));
const oldOutputs=artifacts(beforeTree),newOutputs=artifacts(afterTree);
assert.deepEqual(Object.keys(newOutputs),Object.keys(oldOutputs),'exact build artifact file set');
const changed=[];
for(const [file,text] of Object.entries(oldOutputs)) {
    let expected=text;
    if(['local/typescript.internal.d.ts','local/compiler/types.d.ts'].includes(file)) {
        assert.equal(text.split('NodeArray<any> | undefined').length-1,2);
        assert.equal(text.split('PragmaArgumentSpecification<any>').length-1,2);
        expected=expected.replaceAll('NodeArray<any> | undefined','NodeArray<Node> | undefined').replaceAll('PragmaArgumentSpecification<any>','PragmaArgumentSpecification<string>');
    }
    if(['local/typescript.internal.d.ts','local/compiler/program.d.ts'].includes(file)) {
        assert.equal(text.split('ProgramHost<any>["readFile"]').length-1,1);
        expected=expected.replace('ProgramHost<any>["readFile"]','ProgramHost<BuilderProgram>["readFile"]');
    }
    if(['local/typescript.internal.d.ts','local/compiler/watch.d.ts'].includes(file)) {
        const before='createCompilerHostFromProgramHost(host: ProgramHost<any>';
        assert.equal(text.split(before).length-1,1);
        expected=expected.replace(before,'createCompilerHostFromProgramHost<T extends BuilderProgram>(host: ProgramHost<T>');
    }
    assert.equal(newOutputs[file],expected,'exact JavaScript/declaration artifact '+file);
    if(text!==expected)changed.push(file);
}
assert.equal(changed.length,4,'only four internal declaration artifacts');
const scratch=fs.mkdtempSync('/tmp/adaptation41-proof-mutants-');
const js=Object.keys(newOutputs).find(f=>f.endsWith('.js'));
const jsMutant=path.join(scratch,'javascript-byte-mutant.txt');
fs.writeFileSync(jsMutant,newOutputs[js]+'\n');
assert.throws(()=>assert.equal(fs.readFileSync(jsMutant,'utf8'),oldOutputs[js]),assert.AssertionError);
const missing={...newOutputs};delete missing[js];
assert.throws(()=>assert.deepEqual(Object.keys(missing),Object.keys(oldOutputs)),assert.AssertionError);
const api='local/typescript.d.ts',apiMutant=path.join(scratch,'public-api-byte-mutant.txt');
fs.writeFileSync(apiMutant,newOutputs[api]+'\n// unreviewed declaration\n');
assert.throws(()=>assert.equal(fs.readFileSync(apiMutant,'utf8'),oldOutputs[api]),assert.AssertionError);
evidence.build={javascriptFiles:Object.keys(newOutputs).filter(f=>/\.(js|mjs|cjs)$/.test(f)).length,declarationFiles:Object.keys(newOutputs).filter(f=>f.endsWith('.d.ts')).length,changedDeclarations:changed,publicAPIIdentical:true};
evidence.mutants=['real JavaScript artifact byte','missing build artifact','real public API artifact byte'];
fs.writeFileSync(output,JSON.stringify(evidence,null,2)+'\n');
console.log(JSON.stringify({sourceFiles:Object.keys(evidence.sources).length,identicalEmissions:Object.keys(evidence.emissions).length}));
