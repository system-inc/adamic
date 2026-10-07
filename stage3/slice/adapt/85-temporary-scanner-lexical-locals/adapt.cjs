const fs=require('node:fs'),path=require('node:path'),ts=require(process.env.SLICE_TYPESCRIPT||'typescript');
const tree=path.resolve(process.argv[2]);if(!fs.existsSync(path.join(tree,'slice.json')))throw new Error('adaptation 85 requires a declaration slice');
const file=path.join(tree,'src/compiler/scanner.ts');let text=fs.readFileSync(file,'utf8');
const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),edits=[],names=new Map();
function visit(node){
 if(ts.isVariableDeclarationList(node)&&!(node.flags&ts.NodeFlags.BlockScoped)){
  const statement=node.parent,body=statement.parent,func=body.parent;
  if(!ts.isVariableStatement(statement)||!ts.isBlock(body)||!ts.isFunctionDeclaration(func)||!['createScanner','scanRegularExpressionWorker'].includes(func.name?.text))throw new Error('var is not in an authorized function body');
  const seen=names.get(func)||new Set();names.set(func,seen);
  for(const decl of node.declarations){if(!ts.isIdentifier(decl.name)||seen.has(decl.name.text))throw new Error('duplicate or destructured var');seen.add(decl.name.text);}
  const start=node.getStart(source);if(text.slice(start,start+3)!=='var')throw new Error('unexpected var spelling');edits.push(start);
 }
 ts.forEachChild(node,visit);
}
visit(source);
if(edits.length!==0&&edits.length!==23)throw new Error(`unexpected var count ${edits.length}`);
for(const start of edits.sort((a,b)=>b-a))text=text.slice(0,start)+'let'+text.slice(start+3);
if(edits.length)fs.writeFileSync(file,text);console.log(JSON.stringify({lexicalLocals:edits.length}));
