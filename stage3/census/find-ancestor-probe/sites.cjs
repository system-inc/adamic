const ts=require('/workspace/scratch/probe-apply/stage3/api/node_modules/typescript');
const fs=require('fs'),path=require('path');
const evidence='/workspace/scratch/probe-evidence',root='/workspace/scratch/fa-adapted/src';
const records=[],cache=new Map();
function source(file){if(!cache.has(file)){const text=fs.readFileSync(path.join(root,file),'utf8');cache.set(file,ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true));}return cache.get(file);}
function loc(sf,pos){let p=sf.getLineAndCharacterOfPosition(pos);return sf.fileName+':'+(p.line+1)+':'+(p.character+1);}
function record(file,line,col,category){const sf=source(file),pos=sf.getPositionOfLineAndCharacter(line-1,col-1);let call=null,node=null;
 function walk(n){if(n.getStart(sf)<=pos&&pos<n.end){node=n;if(ts.isCallExpression(n))call=n;ts.forEachChild(n,walk);}}walk(sf);
 // No enclosing call exists for the two hook assignments. Retain the exact assignment expression as a bounded absence check.
 let region=call||node;while(!call&&region.parent&&!ts.isBinaryExpression(region)&&!ts.isStatement(region)){region=region.parent;}
 return {category,site:file+':'+line+':'+col,file,line,col,call:call?loc(sf,call.getStart(sf)):null,start:region.getStart(sf),end:region.end,region_kind:ts.SyntaxKind[region.kind],source:region.getText(sf),start_where:loc(sf,region.getStart(sf)),end_where:loc(sf,region.end)};
}
for(const side of ['parent','head'])for(const line of fs.readFileSync(path.join(evidence,'findings-only-'+side+'.txt'),'utf8').split('\n')){if(!line.includes(' | refusal_scan | '))continue;const where=line.split(' | ')[3];const m=where.match(/^(.*):(\d+):(\d+)$/);records.push(record(m[1],+m[2],+m[3],side==='parent'?'cleared-array':'new-method'));}
// First unchanged append call in three affected files, in lexical order. Require identical findings inside on the original sides.
const originals=['parent','head'].map(s=>fs.readFileSync(path.join(evidence,s+'-entry-run.jsonl'),'utf8').trim().split('\n').flatMap(l=>JSON.parse(l).findings||[]));
for(const file of ['compiler/transformers/es2015.ts','compiler/transformers/es2018.ts','compiler/transformers/module/module.ts']){
 const sf=source(file),calls=[];function walk(n){if(ts.isCallExpression(n)&&n.expression.getText(sf)==='append')calls.push(n);ts.forEachChild(n,walk);}walk(sf);
 const call=calls.find(n=>!records.some(r=>r.file===file&&r.start<n.end&&n.getStart(sf)<r.end)&&JSON.stringify(originals[0].filter(f=>{const m=f.where.match(/\/src\/(.*):(\d+):(\d+)$/);if(!m||m[1]!==file)return false;const pos=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1);return n.getStart(sf)<=pos&&pos<n.end;}))===JSON.stringify(originals[1].filter(f=>{const m=f.where.match(/\/src\/(.*):(\d+):(\d+)$/);if(!m||m[1]!==file)return false;const pos=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1);return n.getStart(sf)<=pos&&pos<n.end;})));
 if(!call)throw Error('no control '+file);const p=sf.getLineAndCharacterOfPosition(call.getStart(sf));records.push(record(file,p.line+1,p.character+1,'control'));
}
fs.writeFileSync('/workspace/scratch/probe-report/stage3/census/find-ancestor-probe/sites.json',JSON.stringify(records,null,2)+'\n');
console.log('sites child_exit=0 count='+records.length+' calls='+records.filter(r=>r.call).length);
