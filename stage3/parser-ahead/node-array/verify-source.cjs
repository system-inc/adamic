// The source declarations decide metadata coverage; no native acceptance is inferred.
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process');
const root=process.argv[2];
if(cp.execFileSync('git',['rev-parse','HEAD'],{cwd:root,encoding:'utf8'}).trim()!=='050880ce59e30b356b686bd3144efe24f875ebc8')throw Error('source pin differs');
const ts=require(path.join(root,'lib/typescript.js'));
const source=ts.createSourceFile('types.ts',fs.readFileSync(path.join(root,'src/compiler/types.ts'),'utf8'),ts.ScriptTarget.Latest,true);
const declarations=new Map();
for(const n of source.statements)if(ts.isInterfaceDeclaration(n))declarations.set(n.name.text,n);
function fields(name){const n=declarations.get(name);if(!n)throw Error(name);const out=[];
 for(const h of n.heritageClauses||[])for(const t of h.types){const base=t.expression.getText(source);if(base!=='ReadonlyArray'&&base!=='Array')out.push(...fields(base));}
 for(const m of n.members){if(!ts.isPropertySignature(m))throw Error('unexpected member');out.push({name:m.name.getText(source),type:m.type.getText(source),optional:!!m.questionToken});}
 return out;
}
const expected=[{name:'pos',type:'number',optional:false},{name:'end',type:'number',optional:false},{name:'hasTrailingComma',type:'boolean',optional:false},{name:'transformFlags',type:'TransformFlags',optional:false}];
for(const name of ['NodeArray','MutableNodeArray']){const got=fields(name);if(JSON.stringify(got)!==JSON.stringify(expected))throw Error(JSON.stringify({name,got}));console.log(JSON.stringify({name,fields:got,scope:'source schema only; native pending'}));}
