// Namespace functions use their complete original parameter/return declarations.
// Default initializer sugar is represented as an optional parameter in Target.
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const ts=require('/tmp/lane5-b-original/lib/typescript.js');
const ledger=require('../unknown-callable-pairs-ranked.json');
const original='/tmp/lane5-b-original';
const filename='src/compiler/debug.ts';
const source=ts.createSourceFile(filename,fs.readFileSync(path.join(original,filename),'utf8'),ts.ScriptTarget.Latest,true);
const functions=new Map();
function scan(node){if(ts.isFunctionDeclaration(node)&&node.name)functions.set(node.name.text,[...(functions.get(node.name.text)||[]),node]);ts.forEachChild(node,scan);}scan(source);
const probes=[];
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
for(const pair of ledger.filter(p=>[19,43,46,262].includes(p.rank))){
 const declarations=functions.get(pair.field);if(!declarations)throw Error('missing function');
 const overloads=declarations.some(n=>!n.body)?declarations.filter(n=>!n.body):declarations;
 const headers=overloads.map(n=>source.text.slice(n.getStart(source),n.body?.getStart(source)||n.end).trim());
 const signatures=overloads.map(n=>{
  const params=n.parameters.map(p=>{
   if(p.initializer){if(!ts.isStringLiteral(p.initializer))throw Error('unhandled default');return p.name.getText(source)+'?: string';}
   return p.getText(source);
  }).join(', ');
  return (n.typeParameters?.length?'<'+n.typeParameters.map(p=>p.getText(source)).join(', ')+'>':'')+'('+params+'): '+n.type.getText(source)+';';
 });
 const declaration='readonly '+pair.field+': '+(signatures.length===1 ? signatures[0].replace(/\): /,') => ').replace(/;$/,'') : '{'+signatures.join('\n')+'}')+';';
 const readSource=ts.createSourceFile(pair.witness.file,fs.readFileSync(path.join(original,pair.witness.file),'utf8'),ts.ScriptTarget.Latest,true);let read;
 function visit(n){if(ts.isPropertyAccessExpression(n)&&n.name.text===pair.field){const lc=readSource.getLineAndCharacterOfPosition(n.getStart(readSource));if(lc.line+1===pair.witness.line&&lc.character+1===pair.witness.column)read=n;}ts.forEachChild(n,visit);}visit(readSource);if(!read)throw Error('missing read');
 const args=pair.rank===19?'7 as never':pair.rank===46?'{value:3},(node:Node):node is Node=>true':pair.rank===262?'3,4':'3';
 const fixture='rank-'+pair.rank+'/contract-probe.a';fs.mkdirSync(path.join(__dirname,'rank-'+pair.rank),{recursive:true});
 fs.writeFileSync(path.join(__dirname,fixture),'// Original namespace header; defaults made optional; adjacent Node carrier reduced.\ntype AnyFunction = (...args: never[]) => void;\ninterface Node {readonly value:number;}\ninterface Base {readonly '+pair.field+':unknown;}\ninterface Target {'+declaration+'}\nfunction probe(value:Base):void {const Debug:Target=value as Target;console.log(typeof '+read.getText(readSource)+');}\nprobe({'+pair.field+':7});\n');
 probes.push({rank:pair.rank,type:pair.type,field:pair.field,candidateReads:pair.reads,witness:pair.witness,read:read.getText(readSource),declaration,originalFunctionHeaders:headers,defaultConversion:'string literal initializer becomes optional string parameter',declarationFile:filename,declarationSha256:hash(source.text),fileSha256:hash(readSource.text),utf16Start:read.getStart(readSource),utf16End:read.end,filename:fixture,output:'number\n',refusal:pair.rank===43||pair.rank===46?'adamic/no-type-predicate':'checked view read of field '+pair.field+' with unsupported callable contract'});
}
fs.writeFileSync(path.join(__dirname,'batch-03-debug-probes.json'),JSON.stringify(probes,null,2)+'\n');
