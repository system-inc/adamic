// Original tracing namespace headers, enum values, and recursive Args dictionary.
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const ts=require('/tmp/lane5-b-original/lib/typescript.js');
const original='/tmp/lane5-b-original',filename='src/compiler/tracing.ts';
const bytes=fs.readFileSync(path.join(original,filename),'utf8');
const source=ts.createSourceFile(filename,bytes,ts.ScriptTarget.Latest,true);
const functions=new Map(),carriers=[];
function scan(n){if(ts.isFunctionDeclaration(n)&&n.name)functions.set(n.name.text,n);if(ts.isEnumDeclaration(n)&&n.name.text==='Phase'||ts.isInterfaceDeclaration(n)&&n.name.text==='Args')carriers.push(n.getText(source).replace(/^export\s+/,'').replace(/^const\s+enum/,'enum').replace(/\r\n/g,'\n'));ts.forEachChild(n,scan);}scan(source);
const probes=[],hash=x=>crypto.createHash('sha256').update(x).digest('hex');
for(const pair of require('../unknown-callable-pairs-ranked.json').filter(p=>[64,265].includes(p.rank))){
 const n=functions.get(pair.field);if(!n)throw Error('missing function');
 const header=source.text.slice(n.getStart(source),n.body.getStart(source)).trim();
 const parameters=n.parameters.map(p=>p.initializer?p.name.getText(source)+'?: boolean':p.getText(source)).join(', ');
 const declaration='readonly '+pair.field+': ('+parameters+') => '+n.type.getText(source)+';';
 const readSource=ts.createSourceFile(pair.witness.file,fs.readFileSync(path.join(original,pair.witness.file),'utf8'),ts.ScriptTarget.Latest,true);let read;
 function visit(node){if(ts.isPropertyAccessExpression(node)&&node.name.text===pair.field){const lc=readSource.getLineAndCharacterOfPosition(node.getStart(readSource));if(lc.line+1===pair.witness.line&&lc.character+1===pair.witness.column)read=node;}ts.forEachChild(node,visit);}visit(readSource);if(!read)throw Error('missing read');
 const receiver=read.expression.getText(readSource);
 if(!/^[A-Za-z_$][\w$]*$/.test(receiver))throw Error('complex receiver '+receiver);
 const fixture='rank-'+pair.rank+'/contract-probe.a';fs.mkdirSync(path.join(__dirname,'rank-'+pair.rank),{recursive:true});
 fs.writeFileSync(path.join(__dirname,fixture),'// Original tracing declarations; false default represented as optional boolean.\n'+carriers.join('\n')+'\ninterface Base {readonly '+pair.field+':unknown;}\ninterface Target {'+declaration+'}\nfunction probe(value:Base):void {const '+receiver+':Target=value as Target;console.log(typeof '+read.getText(readSource)+');}\nprobe({'+pair.field+':7});\n');
 probes.push({rank:pair.rank,type:pair.type,field:pair.field,candidateReads:pair.reads,witness:pair.witness,read:read.getText(readSource),declaration,originalFunctionHeaders:[header],carrierDeclarations:carriers,declarationFile:filename,declarationSha256:hash(bytes),fileSha256:hash(readSource.text),utf16Start:read.getStart(readSource),utf16End:read.end,filename:fixture,output:'number\n',refusal:'checked view read of field '+pair.field+' with unsupported callable contract'});
}
fs.writeFileSync(path.join(__dirname,'batch-04-tracing-probes.json'),JSON.stringify(probes,null,2)+'\n');
