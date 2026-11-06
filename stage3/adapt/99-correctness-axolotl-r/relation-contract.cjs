'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict'),child=require('node:child_process'),ts=require('typescript');
function checkUpcast(text,binary,directory) {
 assert(binary&&path.isAbsolute(binary)&&fs.existsSync(binary),'provide an absolute built Adamic compiler path');
 const source=ts.createSourceFile('core.ts',text,ts.ScriptTarget.Latest,true);
 assert.equal(source.parseDiagnostics.length,0);
 const owners=source.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='toSorted');assert.equal(owners.length,1);
 const owner=owners[0];assert.equal(owner.typeParameters.length,1);assert.equal(owner.typeParameters[0].getText(source),'T');
 const conditionals=[];function visit(n){if(ts.isConditionalExpression(n))conditionals.push(n);ts.forEachChild(n,visit);}visit(owner);assert.equal(conditionals.length,1);
 const arm=conditionals[0].whenTrue;assert(ts.isAsExpression(arm),'the adapted empty arm must have its explicit view');
 assert(ts.isIdentifier(arm.expression)&&arm.expression.text==='emptyArray','probe exactly the shared empty read');
 const declarations=source.statements.filter(n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>ts.isIdentifier(d.name)&&d.name.text==='emptyArray'));
 assert.equal(declarations.length,1);assert.equal(declarations[0].declarationList.declarations.length,1);
 const declaration=declarations[0].getText(source),target=arm.type.getText(source),actual=arm.getText(source);
 const root=directory||fs.mkdtempSync(path.join(os.tmpdir(),'99r-upcast-'));fs.mkdirSync(root,{recursive:true});
 const results=[];
 for(const [name,expression] of [['cast',actual],['relation',`emptyArray satisfies ${target}`]]) {
  const probe=`${declaration}\nfunction probe<T>(): readonly T[] { return ${expression}; }\nconsole.log(String(probe<number>().length));\nconsole.log(String(probe<string>().length));\n`;
  const file=path.join(root,name+'.ts');fs.writeFileSync(file,probe);
  const row={name,expression,file,commands:[]};
  for(const command of ['types','c']) {
   const result=child.spawnSync(binary,[command,file],{encoding:'utf8',maxBuffer:64*1024*1024});assert.ifError(result.error);assert.equal(result.signal,null,'compiler killed by signal');
   fs.writeFileSync(path.join(root,name+'-'+command+'.stdout'),result.stdout);fs.writeFileSync(path.join(root,name+'-'+command+'.stderr'),result.stderr);
   row.commands.push({command:[binary,command,file],exit:result.status,stderr:result.stderr});
  }
  results.push(row);
 }
 const result={declaration,adaptedArm:actual,target,compiler:binary,probeDirectory:root,results};fs.writeFileSync(path.join(root,'result.json'),JSON.stringify(result,null,2)+'\n');
 assert(results.every(r=>r.commands.every(c=>c.exit===0)),'Adamic upcast relation rejected: '+JSON.stringify(result));
 return result;
}
module.exports={checkUpcast};
if(require.main===module){assert.equal(process.argv.length,4,'usage: node relation-contract.cjs TREE BUILT_ADAMIC');console.log(JSON.stringify(checkUpcast(fs.readFileSync(path.join(process.argv[2],'src/compiler/core.ts'),'utf8'),path.resolve(process.argv[3]),process.env.ADAMIC_RELATION_EVIDENCE),null,2));}
