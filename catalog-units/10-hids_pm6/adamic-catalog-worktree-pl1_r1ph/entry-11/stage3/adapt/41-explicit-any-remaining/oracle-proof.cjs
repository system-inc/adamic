'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),cp=require('node:child_process'),crypto=require('node:crypto');
const [tree,control,oracle,output]=process.argv.slice(2);assert(tree&&control&&oracle&&output,'usage: oracle-proof.cjs TREE PASSING_CONTROL ORACLE NEW_OUTPUT');
const originalReport=JSON.parse(fs.readFileSync(path.join(control,'report.json')));
assert.equal(originalReport.status,'pass');assert.equal(originalReport.counts.failing,0);assert.deepEqual(originalReport.baseline_diffs,[]);
fs.mkdirSync(output,{recursive:true});
const file=path.join(tree,'tests/cases/compiler/forInStatement1.ts'),before=fs.readFileSync(file);
function run(name){const fd=fs.openSync(path.join(output,name+'.log'),'w');try{return cp.spawnSync(oracle,[tree,path.join(output,name),'--runners=compiler','--tests=forInStatement1'],{stdio:['ignore',fd,fd]}).status;}finally{fs.closeSync(fd);}}
let mutant;
try {
    fs.writeFileSync(file,Buffer.concat([before,Buffer.from('\r\nconst adaptation41InputMutant: number = "wrong";\r\n')]));
    mutant=run('mutant');
} finally {fs.writeFileSync(file,before);}
assert.equal(mutant,1,'real fixture input mutant fails oracle');
const failed=JSON.parse(fs.readFileSync(path.join(output,'mutant/report.json')));
assert.equal(failed.phases.install.exit,0);assert.equal(failed.phases.build.exit,0);
assert(failed.counts.failing>0);assert(failed.baseline_diffs.some(f=>f.includes('forInStatement1')));
assert.deepEqual(fs.readFileSync(file),before,'exact fixture input restoration');
assert.equal(run('restored'),0,'restored fixture oracle passes');
const restored=JSON.parse(fs.readFileSync(path.join(output,'restored/report.json')));
assert.deepEqual(restored.counts,originalReport.counts);assert.deepEqual(restored.baseline_diffs,[]);
const report={mutant:'append number = string to actual forInStatement1.ts input',control:originalReport.counts,mutantCounts:failed.counts,mutantBaselineDiffs:failed.baseline_diffs,restored:restored.counts,restoredBaselineDiffs:restored.baseline_diffs,fixtureSHA256:crypto.createHash('sha256').update(before).digest('hex')};
fs.writeFileSync(path.join(output,'proof.json'),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report));
