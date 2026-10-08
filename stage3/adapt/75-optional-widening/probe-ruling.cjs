const fs=require('fs'),path=require('path'),zlib=require('zlib'),ts=require('typescript'),cp=require('child_process'),assert=require('assert/strict');
const root=__dirname, tree=process.argv[2];
assert(tree && fs.existsSync(path.join(tree,'src/compiler/sys.ts')), 'pass a disposable copy of TREE/src with matching node_modules');
const input=path.join(root,'evidence/read-write/analysis.json.gz');
const tests=[];
for(const name of ['wrong-tag','escaped-literal','missing-row']) {
 const rows=JSON.parse(zlib.gunzipSync(fs.readFileSync(input)));let changed;
 if(name==='wrong-tag') {
  const file=path.join(tree,'src/compiler/symbolWalker.ts'),original=fs.readFileSync(file,'utf8');const sf=ts.createSourceFile(file,original,ts.ScriptTarget.Latest,true);let member;
  function visit(n){if(ts.isIfStatement(n)&&ts.isBinaryExpression(n.expression)&&n.expression.operatorToken.kind===ts.SyntaxKind.AmpersandToken&&ts.isPropertyAccessExpression(n.expression.right)&&n.expression.right.name.text==='TypeParameter') member=n.expression.right.name;ts.forEachChild(n,visit);}visit(sf);assert(member);
  fs.writeFileSync(file,original.slice(0,member.getStart())+'Union'+original.slice(member.end));changed={file,original};
 } else if(name==='escaped-literal') {
  // Pass the primary binding to an external call before the initializing write.
  const file=path.join(tree,'src/compiler/sys.ts'),original=fs.readFileSync(file,'utf8');const sf=ts.createSourceFile(file,original,ts.ScriptTarget.Latest,true);let node;
  function visit(n){if(ts.isBinaryExpression(n)&&ts.isIdentifier(n.left)&&n.left.text==='customLevels'&&ts.isObjectLiteralExpression(n.right))node=n;ts.forEachChild(n,visit);}visit(sf);assert(node);
  let write=node.parent;while(write && !(ts.isBinaryExpression(write) && ts.isElementAccessExpression(write.left)))write=write.parent;assert(write);
  const replacement=ts.createPrinter().printNode(ts.EmitHint.Expression,ts.factory.createPropertyAccessExpression(ts.factory.createCallExpression(ts.factory.createPropertyAccessExpression(ts.factory.createIdentifier('Object'),'keys'),undefined,[ts.factory.createIdentifier('customLevels')]),'length'),sf);
  fs.writeFileSync(file,original.slice(0,write.right.getStart())+replacement+original.slice(write.right.end));changed={file,original};
 } else {const at=rows.rows.findIndex(r=>r.classification==='write');rows.rows.splice(at,1);}
 const mutated='/tmp/optional-widening-'+name+'.json.gz',result='/tmp/optional-widening-'+name+'-result.json',log='/tmp/optional-widening-'+name+'.log';
 fs.writeFileSync(mutated,zlib.gzipSync(JSON.stringify(rows)));
 const fd=fs.openSync(log,'w');let run;
 try {run=cp.spawnSync(process.execPath,[path.join(root,'rebucket.cjs'),tree,mutated,result],{env:process.env,stdio:['ignore',fd,fd]});} finally {fs.closeSync(fd);if(changed)fs.writeFileSync(changed.file,changed.original);}
 if(name==='missing-row'){assert.notEqual(run.status,0);tests.push({name,exit:run.status,caught:'100-site partition assertion'});}
 else {assert.equal(run.status,0,fs.readFileSync(log,'utf8'));const actual=JSON.parse(fs.readFileSync(result));const expected=name==='wrong-tag'?{a:2,b:47,c:16,rest:35}:{a:1,b:47,c:17,rest:35};assert.deepEqual(actual.counts,expected);tests.push({name,exit:run.status,counts:actual.counts,caught:'control bucket counts differ; the affected site is refused',mutation:name==='wrong-tag'?'Actual symbolWalker tag test TypeParameter changed to Union':'Actual sys numeric write RHS uses Object.keys(customLevels).length and passes the alias before the initializing write'});}
}
if(process.argv[3])fs.writeFileSync(process.argv[3],JSON.stringify(tests,null,2)+'\n');console.log(JSON.stringify(tests));
