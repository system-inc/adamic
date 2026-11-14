const ts=require('/tmp/destructure-price-scratch/work/stage3/api/node_modules/typescript'),fs=require('fs'),path=require('path');
const base='/tmp/destructure-price-scratch/adapted',dest='/workspace/scratch/destructure-price-f2989762/counting';
const sites=JSON.parse(fs.readFileSync('/tmp/destructure-price-scratch/ast-sites.json'));
const cfg=ts.getParsedCommandLineOfConfigFile(base+'/src/compiler/tsconfig.json',{}, {...ts.sys,onUnRecoverableConfigFileDiagnostic:d=>{throw Error(ts.flattenDiagnosticMessageText(d.messageText,'\n'))}});const program=ts.createProgram(cfg.fileNames,cfg.options),checker=program.getTypeChecker();const groups=Map.groupBy(sites,s=>s.file);let n=0,identityTests=0;
for(const [file,group] of groups){const text=fs.readFileSync(path.join(base,file),'utf8'),sf=program.getSourceFile(path.join(base,file)),edits=[],constLists=new Set(),originals=new Map();
 const insert=(position,text)=>edits.push({start:position,end:position,text});
 function walk(node){if(ts.isObjectBindingPattern(node)){
 const line=sf.getLineAndCharacterOfPosition(node.getStart()).line+1,site=group.find(s=>s.line===line);
 if(site){n++;const id=file+':'+line;
 let fn=node.parent;while(fn&&!ts.isFunctionLike(fn))fn=fn.parent;
 const enter=`\n(globalThis as any).__dpEnter(${JSON.stringify(id)});\n`;
 const assignments=node.elements.map((e,index)=>{if(!ts.isIdentifier(e.name))throw Error('nested binding');const member=(e.propertyName||e.name).getText(sf),local=e.name.getText(sf);const original='__dpOriginal_'+line+'_'+index;originals.set(checker.getSymbolAtLocation(e.name),original);return `const ${original} = ${local};\n${local} = (<T,>(value:T):T => (globalThis as any).__dpCount(${JSON.stringify(id+'::'+member)}, value))(${local});`;}).join('\n')+'\n';
 if(ts.isParameter(node.parent)){if(!fn?.body||!ts.isBlock(fn.body))throw Error('parameter body');insert(fn.body.getStart()+1,enter+assignments);}
 else {if(!ts.isVariableDeclaration(node.parent))throw Error('unexpected binding');const list=node.parent.parent,statement=list.parent;if(!ts.isVariableStatement(statement))throw Error('unexpected statement');
 if(list.flags&ts.NodeFlags.Const)constLists.add(list);
 insert(statement.end,'\n'+assignments);
 if(fn){if(!fn.body||!ts.isBlock(fn.body))throw Error('unexpected enclosing body');insert(fn.body.getStart()+1,enter);}
 else insert(statement.getStart(),enter);
 }
 }}ts.forEachChild(node,walk);}walk(sf);
 function identities(node){if(ts.isBinaryExpression(node)&&[ts.SyntaxKind.EqualsEqualsToken,ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(node.operatorToken.kind)){for(const operand of [node.left,node.right])if(ts.isIdentifier(operand)){const alias=originals.get(checker.getSymbolAtLocation(operand));if(alias){edits.push({start:operand.getStart(),end:operand.end,text:alias});identityTests++;}}}ts.forEachChild(node,identities);}identities(sf);
 for(const list of constLists){const start=list.getStart();if(text.slice(start,start+5)!=='const')throw Error('const token mismatch');edits.push({start,end:start+5,text:'let'});}
 let result=text;for(const edit of edits.sort((a,b)=>b.start-a.start))result=result.slice(0,edit.start)+edit.text+result.slice(edit.end);
 const check=ts.createSourceFile(file,result,99,true);if(check.parseDiagnostics.length)throw Error('invalid rewritten syntax '+file);
 fs.writeFileSync(path.join(dest,file),result);
}
if(n!==29)throw Error('site count '+n);console.log(JSON.stringify({sites:n,files:groups.size,identityTests,strategy:'wrap local bindings after destructuring, defaults included; all function-valued members counted'}));
