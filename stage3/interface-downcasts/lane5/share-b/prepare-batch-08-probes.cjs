// Original array declarations and read expressions for batch 08.
const fs=require('fs'),path=require('path'),ts=require('/tmp/lane5-b-original/lib/typescript.js'),crypto=require('crypto');
const members=require('./batch-08-original.json').members,arrays=[],admissions=[];
function scaffold(receiver,element) {
 receiver=receiver.trim();
 const target='Target<'+element+'>';
 if(receiver==='getPropertiesOfType(target)')return 'const target={value:3};function getPropertiesOfType(target:Symbol):'+target+' {return value as '+target+';}';
 if(receiver==='getTypeChecker().getGlobalDiagnostics()')return 'const checker={getGlobalDiagnostics:():'+target+'=>value as '+target+'};function getTypeChecker():typeof checker {return checker;}';
 const parts=receiver.replace(/[?!]$/,'').split('.');let holder='value as '+target;
 for(const part of parts.slice(1).reverse())holder='{'+part+':'+holder+'}';
 return 'const '+parts[0]+'='+holder+';';
}
for(const m of members.filter(m=>[1102,1150,1165,1168,1171,1174,1177,1306,1309,1363,1366,1372,1375].includes(m.rank))) {
 const element=m.type.replace(/ \| undefined$/,'').replace(/^readonly /,'').replace(/\[\]$/,'');
 const scalar=['number','unknown'].includes(element),carriers=scalar?'':'interface '+element+' {readonly value:number;}';
 const filename='rank-'+m.rank+'/'+(['lastIndexOf','indexOf','slice'].includes(m.field)?'actual-array':'contract-probe')+'.a';fs.mkdirSync(path.join(__dirname,'rank-'+m.rank),{recursive:true});
 const readonly=m.type.startsWith('readonly ') ? 'readonly ' : '';
 const args=m.field==='slice'?'0,1':m.field==='lastIndexOf'?'3,0':m.field==='indexOf'||m.field==='push'?'{value:3}':m.field==='filter'?'(item:'+element+'):boolean=>true':m.field==='forEach'?'(item:'+element+',index:number,array:'+readonly+element+'[]):void=>{}':'(item:'+element+',index:number,array:'+readonly+element+'[]):number=>3';
 const actual=['lastIndexOf','indexOf','slice'].includes(m.field);
 fs.writeFileSync(path.join(__dirname,filename),'// Original array declaration/read; adjacent elements reduced.\n'+carriers+'\ninterface Base {readonly '+m.field+':unknown;}\ninterface Target<T> {'+m.declaration+'}\nfunction probe(value:Base):void {'+scaffold(m.read.slice(0,m.read.lastIndexOf('.')),element)+m.read+'('+args+');console.log("called");}\n'+(actual?'probe('+(element==='number'?'[3,4]':'[{value:3},{value:4}]')+');':'probe({'+m.field+':7});')+'\n');
 const probe={...m,filename,output:actual?'called\n':'',refusal:m.rank===1177?"stage 0 can't lower a value of type unknown yet":m.field==='filter'?'adamic/no-type-predicate':actual?'':'checked view read of field '+m.field+' with unsupported callable contract'};
 (actual?arrays:admissions).push(probe);
}
const m=members.find(m=>m.rank===1105),sourceFile='src/compiler/types.ts',bytes=fs.readFileSync('/tmp/lane5-b-original/'+sourceFile),source=ts.createSourceFile(sourceFile,bytes.toString('utf8'),ts.ScriptTarget.Latest,true);let declaration;
function visit(n){if(ts.isTypeAliasDeclaration(n)&&n.name.text==='Path')declaration=n.getText(source).replace(/^export\s+/,'');ts.forEachChild(n,visit);}visit(source);
const filename='rank-1105/contract-probe.a';fs.mkdirSync(path.join(__dirname,'rank-1105'),{recursive:true});
fs.writeFileSync(path.join(__dirname,filename),'// Original callable declaration/read; original ReadonlyMap and Path retained.\n'+declaration+'\ninterface SymlinkedDirectory {readonly value:number;}\ninterface Base {readonly getSymlinkedDirectories:unknown;}\ninterface Target {'+m.declaration+'}\nfunction probe(value:Base):void {const symlinkCache=value as Target;'+m.read+'();}\nprobe({getSymlinkedDirectories:7});\n');
admissions.push({...m,filename,output:'',refusal:'a primitive brand member __pathBrand whose type is not void',carrierDeclarations:[declaration],carrierEvidence:[{declaration,sourceFile,sha256:crypto.createHash('sha256').update(bytes).digest('hex')}]});
fs.writeFileSync(path.join(__dirname,'batch-08-array-probes.json'),JSON.stringify(arrays,null,2)+'\n');
fs.writeFileSync(path.join(__dirname,'batch-08-signature-probes.json'),JSON.stringify(admissions,null,2)+'\n');
