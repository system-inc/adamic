const fs=require('fs'),path=require('path'),crypto=require('crypto');
const S='/workspace/bind-extra-parameters-scratch',T=path.join(S,'adapted'),ts=require(path.join(S,'base/stage3/api/node_modules/typescript'));
const files=[];function walk(p){for(const x of fs.readdirSync(p,{withFileTypes:true})){const f=path.join(p,x.name);if(x.isDirectory())walk(f);else if(f.endsWith('.ts'))files.push(f)}}walk(path.join(T,'src/compiler'));walk(path.join(T,'src/tsc'));
let unitCount=0;const sites=[];const sources=[];
for(const file of files.sort()){
 const text=fs.readFileSync(file,'utf8'),sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);unitCount+=sf.statements.length;
 const where=n=>{const q=sf.getLineAndCharacterOfPosition(n.getStart(sf));return `${path.relative(T,file)}:${q.line+1}:${q.character+1}`};
 sources.push({file:path.relative(T,file),sha256:crypto.createHash('sha256').update(text).digest('hex'),bytes:Buffer.byteLength(text)});
 function visit(n){
  if(ts.isFunctionDeclaration(n)&&!ts.isSourceFile(n.parent))unitCount++;
  if(ts.isCallExpression(n)){
   const e=n.expression;let kind;
   if(ts.isPropertyAccessExpression(e)&&e.name.text==='bind')kind='bind';
   if(ts.isIdentifier(e)&&e.text==='maybeBind')kind='maybeBind_call';
   if(kind){let u=n;while(u.parent&&!ts.isSourceFile(u.parent)){if(ts.isFunctionLike(u))break;u=u.parent};const q=sf.getLineAndCharacterOfPosition(n.getStart(sf));sites.push({kind,site:where(n),unit:where(u),expression:n.getText(sf),line:text.split(/\r?\n/)[q.line]});}
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
}
const out={typescript:ts.version,files:sources.length,latent_candidate_units:unitCount,sources,sites};fs.writeFileSync(path.join(S,'bind-sites.json'),JSON.stringify(out,null,2)+'\n');console.log(JSON.stringify({files:out.files,units:unitCount,binds:sites.filter(s=>s.kind==='bind').length,maybeBind_calls:sites.filter(s=>s.kind==='maybeBind_call').length,files_with_sites:[...new Set(sites.map(s=>s.site.split(':')[0]))]}));
