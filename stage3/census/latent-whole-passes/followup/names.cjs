const fs=require('fs'),ts=require('typescript');
const rows=fs.readFileSync(process.argv[2],'utf8').trim().split('\n').map(JSON.parse), cache=new Map(),out={};
for(const row of rows) for(const entry of row.entries||[]) {
 const w=entry.where;if(!w||out[w])continue;const m=/^(.*):(\d+):(\d+)$/.exec(w);if(!m)continue;
 let sf=cache.get(m[1]);if(!sf){sf=ts.createSourceFile(m[1],fs.readFileSync(m[1],'utf8'),ts.ScriptTarget.Latest,true);cache.set(m[1],sf)}
 const pos=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1);let fn,kind=[];
 function visit(n){if(n.getStart(sf)===pos)kind.push(n.kind===ts.SyntaxKind.TypePredicate?"TypePredicate":ts.SyntaxKind[n.kind]);if(n.pos<=pos&&pos<n.end){if(ts.isFunctionLike(n)&&n.name)fn=n.name.getText(sf);ts.forEachChild(n,visit)}}visit(sf);
 out[w]={function:fn||null,node_kinds_at_start:kind};
}
console.log(JSON.stringify(out,null,2));
