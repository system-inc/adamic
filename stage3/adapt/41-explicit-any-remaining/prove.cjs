'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const crypto = require('node:crypto'), ts = require('typescript');
const [beforeTree, afterTree, output] = process.argv.slice(2);
const rules = require('./rules.json');
const hash = s => crypto.createHash('sha256').update(s).digest('hex');
const walk = d => fs.readdirSync(d, {withFileTypes:true}).flatMap(e => e.isDirectory() ? walk(path.join(d,e.name)) : [path.join(d,e.name)]).sort();
const evidence = {typescript:ts.version, rules:[], sources:{}, emissions:{}};
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
fs.writeFileSync(output,JSON.stringify(evidence,null,2)+'\n');
console.log(JSON.stringify({sourceFiles:Object.keys(evidence.sources).length,identicalEmissions:Object.keys(evidence.emissions).length}));
