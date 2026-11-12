const fs=require('fs'),crypto=require('crypto'),ts=require('/workspace/orm-main/stage3/api/node_modules/typescript');
const out=__dirname,main='/workspace/orm-main',base='/workspace/orm-adapted',trees={base,checked:'/workspace/cbp-checked','typed-old':'/workspace/cbp-typed-old'};
const dts=fs.readFileSync(main+'/stage3/api/node_modules/typescript/lib/typescript.d.ts','utf8'),sanctioned=fs.readFileSync(main+'/stage3/lane/sanctioned-api.json','utf8');
const names=['createBuilderProgram','BuilderProgramKind','BuilderCreationParameters','getBuilderCreationParameters','BuilderProgram','SemanticDiagnosticsBuilderProgram','EmitAndSemanticDiagnosticsBuilderProgram','createSemanticDiagnosticsBuilderProgram','createEmitAndSemanticDiagnosticsBuilderProgram'];
const api=Object.fromEntries(names.map(name=>[name,{typescriptDts:new RegExp('\\b'+name+'\\b').test(dts),sanctionedApi:new RegExp('\\b'+name+'\\b').test(sanctioned)}]));
function publicTexts(tree){const sf=ts.createSourceFile('builderPublic.ts',fs.readFileSync(tree+'/src/compiler/builderPublic.ts','utf8'),ts.ScriptTarget.Latest,true);return sf.statements.filter(n=>(ts.isFunctionDeclaration(n)&&!n.body&&['createSemanticDiagnosticsBuilderProgram','createEmitAndSemanticDiagnosticsBuilderProgram'].includes(n.name?.text))||(ts.isInterfaceDeclaration(n)&&['BuilderProgram','SemanticDiagnosticsBuilderProgram','EmitAndSemanticDiagnosticsBuilderProgram'].includes(n.name.text))).map(n=>n.getText(sf));}
const texts=Object.fromEntries(Object.entries(trees).map(([label,tree])=>[label,publicTexts(tree)]));
const e=require(main+'/stage3/adapt/99-correctness-axolotl-x/engine.cjs');
const pins=require(main+'/stage3/adapt/99-correctness-axolotl-x/owners.json').filter(p=>p.file==='builder.ts');
const pinChecks=[];
for(const [label,tree] of Object.entries(trees)){
 const owners=e.ownerMap(e.parse('builder.ts',e.normalize(fs.readFileSync(tree+'/src/compiler/builder.ts','utf8'))));
 for(const pin of pins){const hash=e.digest(owners.get(pin.owner).text);pinChecks.push({label,pin:'stage3/adapt/99-correctness-axolotl-x/owners.json',owner:pin.owner,expected:pin.ownerAfterSha256,actual:hash,unchanged:hash===pin.ownerAfterSha256});}
}
const data={typescript:ts.version,names:api,publicOverloadsAndBuilderInterfacesUnchanged:Object.fromEntries(Object.entries(texts).map(([label,text])=>[label,JSON.stringify(text)===JSON.stringify(texts.base)])),publicDeclarations:texts.base,exactOwnerPins:pinChecks};
if(!pinChecks.every(p=>p.unchanged)||!Object.values(data.publicOverloadsAndBuilderInterfacesUnchanged).every(Boolean))throw Error('unexpected public declaration or pinned owner drift');
fs.writeFileSync(out+'/api-pins.json',JSON.stringify(data,null,2)+'\n');console.log('API declarations unchanged; all',pinChecks.length,'exact nested owner hashes match');
