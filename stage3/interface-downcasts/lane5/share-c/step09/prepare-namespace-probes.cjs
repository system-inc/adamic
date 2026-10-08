const fs=require('fs'),path=require('path');
const root=path.resolve(__dirname,'..'),e=JSON.parse(fs.readFileSync(path.join(root,'original-members.json'))).members;
const carriers=JSON.parse(fs.readFileSync('/tmp/lane5-c-carrier-declarations.json'));
const ts=require('/tmp/lane5-c-original/lib/typescript.js');
const rows=[];
for(const m of e){
 if(!/^typeof Debug$/.test(m.type)||m.declarations.length!==1||!m.declarations[0].text.startsWith('export function '))continue;
 const header=m.declarations[0].text.replace(/^export /,'');
 const source=ts.createSourceFile('original.ts',header,ts.ScriptTarget.Latest,true),fn=source.statements[0];
 if(!fn.parameters||!fn.type)continue;
 const definitions=new Map();let blocked=false;
 function types(n){
  if(ts.isTypeReferenceNode(n)){
   const name=n.typeName.getText();if(!/^\w+$/.test(name)){blocked=true;return;}
   if(name==='T'||['Array','ReadonlyArray'].includes(name))return;
   const original=carriers[name];
   if(original?.kind==='TypeAliasDeclaration')definitions.set(name,original.text.replace(/\r\n/g,'\n'));
   else if(original?.kind==='InterfaceDeclaration')definitions.set(name,'interface '+name+' {readonly value:number;'+(original.kindValue===undefined?'':'readonly kind:'+original.kindValue+';')+'}');
   else blocked=true;
  }
  ts.forEachChild(n,types);
 }
 for(const p of fn.parameters)types(p.type);types(fn.type);if(blocked)continue;
 let body='{}',invocation='';
 if(fn.type.kind===ts.SyntaxKind.NeverKeyword){body='{throw new Error("original failure control");}';invocation='try {'+m.read+'('+fn.parameters.map((p,i)=>i===0&&p.type.kind===ts.SyntaxKind.StringKeyword?'"message"':p.type.getText()==='Node'?'{value:3}':'undefined').join(',')+');} catch {console.log("completed");}';}
 else if(fn.type.kind===ts.SyntaxKind.TypeReference && fn.type.getText()==='T'){
  body='{if(value===undefined || value===null){throw new Error("missing control");}return value;}';invocation=m.read+'(3);console.log("completed");';
 }else if(fn.type.kind===ts.SyntaxKind.VoidKeyword){invocation=m.read+'('+fn.parameters.map(p=>p.type.getText()==='T'?'3':p.type.kind===ts.SyntaxKind.StringKeyword?'"message"':p.type.kind===ts.SyntaxKind.NumberKeyword?'3':'undefined').join(',')+');console.log("completed");';}
 else continue;
 const text='// Original namespace function header; body and adjacent object carriers reduced.\n'+[...definitions.values()].join('\n')+'\n'+header.slice(0,-1)+body+'\ninterface Base {readonly '+m.field+':unknown;}\ninterface Target {readonly '+m.field+':typeof '+m.field+';}\nfunction probe(value:Base):void {const Debug=value as Target;'+invocation+'}\nprobe({'+m.field+':'+m.field+'});\n';
 const directory=path.join(root,'families','rank-'+m.rank);fs.mkdirSync(directory,{recursive:true});fs.writeFileSync(path.join(directory,'original-context.a'),text);
 rows.push({rank:m.rank,reads:m.reads,type:m.type,read:m.read,originalHeader:m.declarations[0].text,status:'original namespace control prepared'});
}
fs.writeFileSync(path.join(__dirname,'namespace-probes.json'),JSON.stringify(rows,null,2)+'\n');console.log('Prepared '+rows.length+' namespace controls.');
