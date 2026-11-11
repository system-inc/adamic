const fs=require('fs'),cp=require('child_process'),crypto=require('crypto');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const root='/workspace/adamic/stage3/adapt/';
for(const [dir,owner,binding,callBefore,callAfter] of [
 ['62-temporary-tracing-write','dumpTypes','descriptor','fs.writeSync(typesFd, JSON.stringify(descriptor))','fs.writeSync(typesFd, JSON.stringify(descriptor)!)'],
 ['63-temporary-tracing-legend','dumpLegend','legend','fs.writeFileSync(legendPath, JSON.stringify(legend))','fs.writeFileSync(legendPath, JSON.stringify(legend)!)']]){
 const text=cp.execFileSync('git',['-C','/workspace/tracing-before','show','refs/stage3/'+dir+'-before:src/compiler/tracing.ts'],{encoding:'utf8'});
 const source=ts.createSourceFile('tracing.ts',text,ts.ScriptTarget.Latest,true);let body,literal;
 function walk(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===owner)body=n.getText(source);if(ts.isVariableDeclaration(n)&&n.name.getText(source)===binding)literal=n.getText(source);ts.forEachChild(n,walk)}walk(source);
 const after=body.replace(callBefore,callAfter);if(after===body)throw Error('missing call');
 const sha=x=>crypto.createHash('sha256').update(x).digest('hex');
 fs.writeFileSync(root+dir+'/pins.json',JSON.stringify({version:ts.version,commit:'050880ce59e30b356b686bd3144efe24f875ebc8',owner,binding,literal,callBefore,callAfter,before:body,after,beforeSha256:sha(body),afterSha256:sha(after)},null,2)+'\n');
 fs.copyFileSync(require('path').join(__dirname,'adapter-template.cjs'),root+dir+'/adapt.cjs');
}
const dir=root+'99-correctness-tracing-dumptypes/';const p=JSON.parse(fs.readFileSync(dir+'pins.json'));
for(const state of ['before','after']){const old='(fs as { writeSync(fd: number, data: string | undefined): number }).writeSync(typesFd, JSON.stringify(descriptor))';if(!p[state].includes(old))throw Error('99 pin drift');p[state]=p[state].replace(old,'fs.writeSync(typesFd, JSON.stringify(descriptor)!)');p[state+'Sha256']=crypto.createHash('sha256').update(p[state]).digest('hex')}
fs.writeFileSync(dir+'pins.json',JSON.stringify(p,null,2)+'\n');
