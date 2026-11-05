'use strict';
// Stock checker audit: own operand types and the exact unknown-hop brand shape.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const {owners,digest}=require('./adapt.cjs');
assert.equal(ts.version,'6.0.3');
const tree=path.resolve(process.argv[2]),dir=path.join(tree,'src/compiler'),file=path.join(dir,'core.ts');
const raw=fs.readFileSync(file,'utf8'),text=raw.replace(/\r\n/g,'\n');let after=text;
for(const pin of owners){assert(after.includes(pin.before));after=after.replace(pin.before,pin.after);}
if(raw.includes('\r\n')) after=after.replace(/\n/g,'\r\n');
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
const roots=fs.readdirSync(dir,{recursive:true}).filter(n=>n.endsWith('.ts')).sort().map(n=>path.join(dir,n));assert.equal(roots.length,79);
function program(code,target=file,rootNames=roots){const host=ts.createCompilerHost(options),read=host.readFile,exists=host.fileExists;host.readFile=f=>path.resolve(f)===target?code:read(f);host.fileExists=f=>path.resolve(f)===target||exists(f);return ts.createProgram(rootNames,options,host);}
function diagnostics(p){return ts.getPreEmitDiagnostics(p).map(d=>({code:d.code,file:d.file&&path.relative(tree,d.file.fileName),...(d.file&&d.start!==undefined?(()=>{const pos=d.file.getLineAndCharacterOfPosition(d.start);return {line:pos.line+1,column:pos.character+1};})():{}),message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));}
const beforeDiagnostics=diagnostics(program(raw)),p=program(after),afterDiagnostics=diagnostics(p);assert.deepEqual(afterDiagnostics,beforeDiagnostics,'new stock checker diagnostic');
const checker=p.getTypeChecker(),source=p.getSourceFile(file),sites=[];
const strip=n=>ts.skipParentheses(n);
const arrayType=t=>t.isUnion()?t.types.every(arrayType):(t.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown))===0&&(checker.isArrayType(t)||checker.isTupleType(t));
for(const node of source.statements){
 if(!ts.isFunctionDeclaration(node)||!owners.some(pin=>pin.owner===node.name?.text))continue;
 function visit(n){if(ts.isAsExpression(n)&&ts.isTypeReferenceNode(n.type)&&['SortedArray','SortedReadonlyArray'].includes(n.type.typeName.getText(source))){
  const hop=strip(n.expression);assert(ts.isAsExpression(hop));assert.equal(hop.type.kind,ts.SyntaxKind.UnknownKeyword);
  const operand=strip(hop.expression);assert(!ts.isAsExpression(operand));const type=checker.getTypeAtLocation(operand);assert(arrayType(type),'non-array own operand');
  if(node.name.text==='toSorted')assert(ts.isConditionalExpression(operand),'toSorted must mint the complete conditional');
  const pos=source.getLineAndCharacterOfPosition(n.getStart(source));sites.push({owner:node.name.text,line:pos.line+1,column:pos.character+1,expression:n.getText(source),operand:operand.getText(source),ownType:checker.typeToString(type,operand,ts.TypeFormatFlags.NoTruncation),typeFlags:type.flags,isArrayOrUnionOfArrays:arrayType(type),isAny:!!(type.flags&ts.TypeFlags.Any),isUnknown:!!(type.flags&ts.TypeFlags.Unknown),shape:'as Brand -> as unknown -> array operand (parentheses ignored)'});
 }ts.forEachChild(n,visit);}visit(node);
}
assert.equal(sites.length,4);assert.deepEqual(sites.map(s=>s.line),[736,759,764,1039]);
const pubRaw=fs.readFileSync(path.join(dir,'corePublic.ts'),'utf8'),pub=ts.createSourceFile('corePublic.ts',pubRaw,ts.ScriptTarget.Latest,true);
const brands=pub.statements.filter(n=>ts.isInterfaceDeclaration(n)&&['SortedArray','SortedReadonlyArray'].includes(n.name.text)).map(n=>{assert.equal(n.members.length,1);assert.equal(n.members[0].name.text,' __sortedArrayBrand');assert.equal(n.members[0].type.kind,ts.SyntaxKind.UndefinedKeyword);return n.getText(pub);});assert.equal(brands.length,2);
const original=ts.createSourceFile('core.ts',raw,ts.ScriptTarget.Latest,true),empty=original.statements.find(n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(original)==='emptyArray')).getText(original);
const witness=brands.join('\n')+'\n'+empty+'\nfunction allMints<T>(array: readonly T[], comparer?: (a:T,b:T)=>number) {\nconst deduplicated: T[] = [];\n'+sites.map((s,i)=>'const mint'+i+' = '+s.expression+';').join('\n')+'\nreturn [mint0,mint1,mint2,mint3];\n}\n';
const witnessFile=path.join(tree,'sorted-brand-unknown-witness.ts'),witnessDiagnostics=diagnostics(program(witness,witnessFile,[witnessFile]));assert.deepEqual(witnessDiagnostics,[]);
assert.equal(fs.readFileSync(file,'utf8'),raw);
console.log(JSON.stringify({typescript:ts.version,roots:roots.length,options,sites,brands,sourceUnchanged:true,coreBeforeSha256:digest(raw),coreAfterSha256:digest(after),corePublicSha256:digest(pubRaw),beforeDiagnostics,afterDiagnostics,stockAmbientContext:'The unchanged full-stock diagnostics are retained; Adamic supplies its embedded host overlay in the independent types checks.',newStockDiagnostics:0,witness,witnessDiagnostics},null,2));
