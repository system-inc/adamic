// Source-level assignability candidates, a deliberately named upper bound on IR records.
const m=require('./measure.cjs');
const {ts,fs,path,checker,functions,sites,root,out,loc,key,textType}=m;
const rows=[];
for(const site of sites){
 const type=site.type;
 const typed=functions.filter(f=>checker.isTypeAssignableTo(f.type,type));
 const named=site.kind==='bind'&&site.memberName?functions.filter(f=>ts.isMethodDeclaration(f.node)&&ts.isClassLike(f.node.parent)&&f.node.name?.getText()===site.memberName):[];
 const fInfo=f=>({key:f.key,name:f.name,N:f.N,rest:f.rest,readsArguments:f.readsArguments,generic:!!f.node.typeParameters?.length,thisParameter:f.thisParameter,signature:f.signature});
 const extra=typed.filter(f=>f.N>Math.max(...site.M));
 rows.push({...loc(site.node),member:site.member,kind:site.kind,type:site.typeText,M:site.M,typed:typed.map(fInfo),extra:extra.map(fInfo),methods:named.map(fInfo)});
}
fs.writeFileSync(path.join(out,'candidates.json'),JSON.stringify(rows,null,2)+'\n');
console.log('CANDIDATES');
for(const r of rows) console.log(JSON.stringify({file:r.file,line:r.line,member:r.member,M:r.M,typed:r.typed.length,extra:r.extra.map(f=>[f.key,f.name,f.N,f.rest?'rest':'',f.generic?'generic':'']),methods:r.methods}));
