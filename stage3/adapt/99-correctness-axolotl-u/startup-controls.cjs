const fs=require('fs'),assert=require('assert/strict'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
require('/workspace/scratch/99u-census-preload.cjs');
const manifest=JSON.parse(fs.readFileSync('/workspace/scratch/99u-guard-manifest.json','utf8'));
const kinds=['PrefixUnaryExpression','GetAccessor','ComputedPropertyName','ArrayLiteralExpression','MethodDeclaration','BindingElement','PropertyDeclaration','Parameter','PropertyAssignment','ClassExpression','Unknown'];
const results=[];
for(let i=0;i<manifest.length;i++){
 const site=manifest[i],file='/workspace/scratch/99u-counted/src/compiler/'+site.file,text=fs.readFileSync(file,'utf8'),sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);const found=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===site.owner)found.push(n);ts.forEachChild(n,visit);}visit(sf);assert.equal(found.length,1);
 const code=ts.transpileModule(found[0].getText(sf).replace(/^export /,''),{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
 const stubs={findAncestor:()=>undefined,isPartOfTypeNode:()=>false,isTypeQueryNode:()=>false,isEntityName:()=>false,isEntityNameExpression:()=>false};for(const name of ['createAccessorTypeError','createObjectLiteralError','createArrayLiteralError','createReturnTypeError','createBindingElementError','createVariableOrPropertyError','createParameterError','createExpressionError','createClassExpressionError'])stubs[name]=()=>undefined;
 const scope={...ts,...stubs};const fn=new Function('scope',`with(scope) { ${code}; return ${site.owner}; }`)(scope);
 let reads=0;const input={get kind(){reads++;return reads===1?ts.SyntaxKind[kinds[i]]:ts.SyntaxKind.Unknown;},parent:undefined};
 globalThis.__claimControlTitle=site.id;fn(input);delete globalThis.__claimControlTitle;results.push({id:site.id,kindReads:reads,path:'exact instrumented lexical owner',plantedKind:kinds[i],laterKind:'Unknown'});
}
fs.writeFileSync('/workspace/scratch/99u-startup-control-paths.json',JSON.stringify(results,null,2)+'\n');
