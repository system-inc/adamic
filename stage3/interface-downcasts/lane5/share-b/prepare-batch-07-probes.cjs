// Original Map result contract and Path declaration.
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const ts=require('/tmp/lane5-b-original/lib/typescript.js');
const m=require('./batch-07-original.json').members.find(m=>m.rank===994);
const sourceFile='src/compiler/types.ts',bytes=fs.readFileSync('/tmp/lane5-b-original/'+sourceFile),source=ts.createSourceFile(sourceFile,bytes.toString('utf8'),ts.ScriptTarget.Latest,true);let declaration;
function visit(n){if(ts.isTypeAliasDeclaration(n)&&n.name.text==='Path')declaration=n.getText(source).replace(/^export\s+/,'');ts.forEachChild(n,visit);}visit(source);
const filename='rank-994/contract-probe.a';fs.mkdirSync(path.join(__dirname,'rank-994'),{recursive:true});
fs.writeFileSync(path.join(__dirname,filename),'// Original callable declaration/read; Map and primitive Path retained.\n'+declaration+'\ninterface PackageJsonInfoCacheEntry {readonly value:number;}\ninterface Base {readonly getInternalMap:unknown;}\ninterface Target {'+m.declaration+'}\nfunction probe(value:Base):void {const moduleResolutionCache={getPackageJsonInfoCache:():Target=>value as Target};'+m.read+'();}\nprobe({getInternalMap:7});\n');
fs.writeFileSync(path.join(__dirname,'batch-07-signature-probes.json'),JSON.stringify([{...m,filename,output:'',refusal:'a primitive brand member __pathBrand whose type is not void',carrierDeclarations:[declaration],carrierEvidence:[{declaration,sourceFile,sha256:crypto.createHash('sha256').update(bytes).digest('hex')}]}],null,2)+'\n');
