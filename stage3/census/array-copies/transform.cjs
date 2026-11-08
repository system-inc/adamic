/* Apply only to a scratch stock source tree. No compiler API emit is used.
 * node transform.cjs TYPESCRIPT_JS STOCK_JSON SCRATCH_TREE
 * Counter preload: ARRAY_COPIES_OUTPUT=... node -r ./counter.cjs built/tsc.js ...
 */
const fs=require('fs'),path=require('path');
const [api,json,tree]=process.argv.slice(2),ts=require(path.resolve(api));
const census=JSON.parse(fs.readFileSync(json,'utf8'));
const sites=census.sites.filter(s=>s.depth>0&&s.classification!=='unresolved');
const f=ts.factory;
const call=(name,args)=>f.createCallExpression(f.createPropertyAccessExpression(f.createIdentifier('globalThis'),name),undefined,args);
const str=x=>f.createStringLiteral(x);
const groups=Map.groupBy(sites,s=>s.file);
let wrapped=0;
for(const [file,rows]of groups){
 const filename=path.join(tree,file),text=fs.readFileSync(filename,'utf8');
 const sf=ts.createSourceFile(filename,text,ts.ScriptTarget.Latest,true);
 const lookup=new Map(rows.map(s=>[`${s.start}:${s.end}`,s]));
 const result=ts.transform(sf,[context=>source=>{
  function visit(n){
   const s=lookup.get(`${n.getStart(sf)}:${n.end}`);
   const relevant=s&&n.end===s.end&&((s.method==='spread'&&ts.isSpreadElement(n))||(s.method!=='spread'&&ts.isCallExpression(n)));
   const updated=ts.visitEachChild(n,visit,context);
   if(!relevant)return updated;
   wrapped++;
   if(s.method==='spread')return f.updateSpreadElement(updated,call('__acSpread',[str(s.id),updated.expression]));
   if(n.questionDotToken||n.expression.questionDotToken)throw Error('Optional call requires explicit transform: '+s.id);
   if(s.method.startsWith('.')){
    return call('__acMethod',[str(s.id),updated.expression.expression,str(updated.expression.name.text),f.createArrowFunction(undefined,undefined,[],undefined,f.createToken(ts.SyntaxKind.EqualsGreaterThanToken),f.createArrayLiteralExpression(updated.arguments)),str(s.method)]);
   }
   return call('__acCall',[str(s.id),updated.expression,f.createArrayLiteralExpression(updated.arguments),str(s.method)]);
  }
  return ts.visitNode(source,visit);
 }]);
 fs.writeFileSync(filename,ts.createPrinter({newLine:ts.NewLineKind.LineFeed}).printFile(result.transformed[0]));
 result.dispose();
}
if(wrapped!==sites.length)throw Error(`Wrapped ${wrapped}, expected ${sites.length}`);
console.log(JSON.stringify({wrapped,files:groups.size}));
