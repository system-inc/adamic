const fs=require('fs'),ts=require('/tmp/parser-fixed-feature-scratch/stage3/api/node_modules/typescript');
const file=process.argv[2],text=fs.readFileSync(file,'utf8'),sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);let node;
function walk(n){if(ts.isClassDeclaration(n)&&n.name?.text==='DebugTypeMapper')node=n;ts.forEachChild(n,walk);}walk(sf);if(!node)throw Error('No class');
const start=node.getStart(sf),end=node.end;
const replacement='export interface DebugTypeMapper { kind: TypeMapKind; __debugToString(): string; }\nexport const DebugTypeMapper: { prototype: DebugTypeMapper } = undefined!;'+ '\n'.repeat((text.slice(start,end).match(/\n/g)||[]).length-1);
fs.writeFileSync(file,text.slice(0,start)+replacement+text.slice(end));console.log(JSON.stringify({file,start,end,replacement,warning:'Artificial class shape and checked-undefined value stub; never run or treat as adaptation'}));
