'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),cp=require('node:child_process'),crypto=require('node:crypto');
const [tree,oracle,out]=process.argv.slice(2);assert(tree&&oracle&&out);
fs.mkdirSync(out,{recursive:true});
const api=path.join(tree,'tests/baselines/reference/api/typescript.d.ts'),before=fs.readFileSync(api);
const log=fs.openSync(path.join(out,'run.log'),'w');let result;
try {
    fs.appendFileSync(api,'\ntype UnreviewedClassBaselineMutant = string;\n');
    result=cp.spawnSync(oracle,[tree,path.join(out,'oracle'),'--runners=unittest','--tests=Public APIs'],{stdio:['ignore',log,log]});
} finally {fs.writeFileSync(api,before);fs.closeSync(log);}
assert.equal(result.status,1,'real API reference mutant must fail');
const report=JSON.parse(fs.readFileSync(path.join(out,'oracle/report.json')));
assert(report.counts.failing>0&&report.baseline_diffs.includes('api/typescript.d.ts'));
assert.deepEqual(fs.readFileSync(api),before,'exact reference restoration');
const walk=d=>fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(d,e.name)):[path.join(d,e.name)]);
const local=path.join(tree,'tests/baselines/local'),reference=path.join(tree,'tests/baselines/reference');
for(const file of walk(local)){const ref=path.join(reference,path.relative(local,file));assert(fs.existsSync(ref));assert.deepEqual(fs.readFileSync(file),fs.readFileSync(ref),'remaining baseline difference');}
const proof={mutant:'unrelated declaration appended to actual API reference',exit:result.status,failing:report.counts.failing,differences:report.baseline_diffs,restoredSha256:crypto.createHash('sha256').update(before).digest('hex'),remainingBaselineDifferences:0};
fs.writeFileSync(path.join(out,'proof.json'),JSON.stringify(proof,null,2)+'\n');console.log(JSON.stringify(proof));
