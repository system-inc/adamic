'use strict';
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process');
const [repo,tree,scratch,baselinePath,output]=process.argv.slice(2);
const baseline=JSON.parse(fs.readFileSync(baselinePath,'utf8'));
const file=path.join(tree,baseline.roots[0].file),original=fs.readFileSync(file,'utf8');
const line=original.split('\n').length+1;
const addition='\nfunction __voidReadProbe(): void {}\nconst __voidReadValue = __voidReadProbe();\n';
const mutantPath=path.join(scratch,'mutant.json');
try {
 fs.writeFileSync(file,original+addition);
 const run=cp.spawnSync('timeout',['180',process.execPath,path.join(__dirname,'census.cjs'),repo,tree,path.join(scratch,'checker'),mutantPath],{encoding:'utf8',timeout:185000});
 if(run.status!==0)throw Error(run.stdout+run.stderr);
 const mutant=JSON.parse(fs.readFileSync(mutantPath,'utf8'));
 const key=r=>JSON.stringify([r.file,r.line,r.column,r.expression,r.form]);
 const old=new Set(baseline.sites.map(key)),now=new Set(mutant.sites.map(key));
 const added=mutant.sites.filter(r=>!old.has(key(r))),lost=baseline.sites.filter(r=>!now.has(key(r)));
 if(mutant.sites.length!==baseline.sites.length+1||added.length!==1||lost.length||added[0].line!==line+1||added[0].expression!=='__voidReadProbe()'||mutant.unionWithVoid.length!==baseline.unionWithVoid.length)throw Error('mutant delta incorrect');
 const proof={baselineCount:baseline.sites.length,mutantCount:mutant.sites.length,added,lost,expectedLine:line+1,unionCountUnchanged:true,sourceAddition:addition,restored:true};
 fs.writeFileSync(output,JSON.stringify(proof,null,2)+'\n');
 console.log('mutant caught: exactly one extra void read at '+added[0].file+':'+added[0].line);
} finally {fs.writeFileSync(file,original);}
