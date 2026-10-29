const fs=require('fs');
const ts=require('/tmp/void-meter-worktree/stage3/api/node_modules/typescript');
const root='/tmp/void-meter';
const records=fs.readFileSync(root+'/latent.jsonl','utf8').trim().split('\n').map(JSON.parse);
const findings=records.slice(1).flatMap(r=>r.findings||[]).filter(f=>f.reason.startsWith("a read of a void container's elements")||f.reason.includes('instantiated at void'));
const files=new Map(); const sites={};
for(const f of findings){
 if(sites[f.where])continue;
 const m=f.where.match(/^(.*):(\d+):(\d+)$/); if(!m)throw Error(f.where);
 const [,name,ln,col]=m;
 if(!files.has(name)) {
  const text=fs.readFileSync(name,'utf8');const sf=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);const starts=new Map();
  function visit(n){let start=n.getStart(sf);if(!starts.has(start))starts.set(start,[]);starts.get(start).push(n);ts.forEachChild(n,visit)}visit(sf);
  files.set(name,{sf,starts,text});
 }
 const {sf,starts,text}=files.get(name); const pos=sf.getPositionOfLineAndCharacter(+ln-1,+col-1);
 let candidates=(starts.get(pos)||[]).filter(n=>ts.isExpression(n)&&!ts.isParenthesizedExpression(n));
 let appropriate=candidates.filter(n=>{
  const p=n.parent;
  if(f.reason.endsWith(', as an argument'))return (ts.isCallExpression(p)||ts.isNewExpression(p))&&p.arguments?.includes(n);
  if(f.reason.endsWith(', as a returned value'))return ts.isReturnStatement(p)||ts.isArrowFunction(p)&&p.body===n;
  return true;
 });
 const n=(appropriate.length?appropriate:candidates)[0];
 if(!n)throw Error('no expression '+f.where);
 sites[f.where]={syntax:ts.SyntaxKind[n.kind],expression:n.getText(sf),parent:ts.SyntaxKind[n.parent.kind],line:text.split(/\r?\n/)[+ln-1].trim()};
}
fs.writeFileSync(root+'/shapes.json',JSON.stringify(sites));
const groups=new Map();
for(const f of findings){const s=sites[f.where];const key=f.reason+'\n'+s.syntax;if(!groups.has(key))groups.set(key,{reason:f.reason,syntax:s.syntax,rows:0,sites:new Set()});let g=groups.get(key);g.rows++;g.sites.add(f.where)}
for(const g of [...groups.values()].sort((a,b)=>b.sites.size-a.sites.size).slice(0,12)){
 console.log(g.sites.size,g.rows,g.syntax,g.reason);
 for(const w of [...g.sites].slice(0,5))console.log(w.replace(root+'/adapted/',''),JSON.stringify(sites[w]));
}
