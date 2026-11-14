const fs=require('fs'),path=require('path'),ts=require('typescript');
const root=process.argv[2],input=process.argv[3];
const rows=fs.readFileSync(input,'utf8').trim().split('\n').map(JSON.parse);
const files=new Map(),out=[];
for(const row of rows)for(const f of row.findings||[]){
 const reason=f.reason||f.Reason||'';
 const m=/ seen as (.*?)(?:, (?:which|whose|a |so )|;|$)/.exec(reason);if(!m)continue;
 const where=f.where||f.Where;const loc=/^(.*):(\d+):(\d+)$/.exec(where||'');if(!loc)continue;
 const filename=loc[1];if(!filename.startsWith(root+'/'))continue;
 let sf=files.get(filename);if(!sf){sf=ts.createSourceFile(filename,fs.readFileSync(filename,'utf8'),ts.ScriptTarget.Latest,true);files.set(filename,sf)}
 const pos=sf.getPositionOfLineAndCharacter(+loc[2]-1,+loc[3]-1);
 let node=sf;function descend(n){if(n.getStart(sf)<=pos&&pos<n.end){node=n;ts.forEachChild(n,descend)}}descend(sf);
 while(node.parent&&node.parent!==sf&&node.parent.getStart(sf)===pos)node=node.parent;
 let call=node.parent;while(call&&!ts.isCallExpression(call)&&!ts.isNewExpression(call))call=call.parent;
 if(!call&&(ts.isCallExpression(node)||ts.isNewExpression(node)))call=node;
 let arg=node;if(call)while(arg.parent&&arg.parent!==call)arg=arg.parent;
 out.push({file:path.relative(root,filename),line:+loc[2],column:+loc[3],callee:call?.expression.getText(sf)||'<not a call>',argument:call?arg.getText(sf):node.getText(sf),target:m[1],kind:f.kind,text:f.text,reason,where});
}
const seen=new Set();console.log(JSON.stringify(out.filter(x=>{const key=JSON.stringify([x.kind,x.where,x.reason,x.text]);if(seen.has(key))return false;seen.add(key);return true}),null,2));
