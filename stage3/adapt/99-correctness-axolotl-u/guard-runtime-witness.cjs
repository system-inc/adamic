const fs=require('fs'),assert=require('assert/strict'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const compiler=require('/workspace/scratch/99u-adapted/built/local/typescript.js');
const manifest=JSON.parse(fs.readFileSync('/workspace/scratch/99u-guard-manifest.json','utf8')).filter(r=>r.id!=='G11');
const kinds=['PrefixUnaryExpression','GetAccessor','ComputedPropertyName','ArrayLiteralExpression','MethodDeclaration','BindingElement','PropertyDeclaration','Parameter','PropertyAssignment','ClassExpression'];const results=[];
for(const level of [0,1,2,3]){compiler.Debug.setAssertionLevel(level);
 for(let i=0;i<manifest.length;i++){
  const site=manifest[i],file='/workspace/scratch/99u-adapted/src/compiler/'+site.file,text=fs.readFileSync(file,'utf8'),sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);let owner;function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===site.owner)owner=n;ts.forEachChild(n,visit);}visit(sf);assert(owner);
  const code=ts.transpileModule(owner.getText(sf).replace(/^export /,''),{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
  const stubs={findAncestor:()=>undefined,isPartOfTypeNode:()=>false,isTypeQueryNode:()=>false,isEntityName:()=>false,isEntityNameExpression:()=>false};
  for(const name of ['createAccessorTypeError','createObjectLiteralError','createArrayLiteralError','createReturnTypeError','createBindingElementError','createVariableOrPropertyError','createParameterError','createExpressionError','createClassExpressionError'])stubs[name]=()=>undefined;
  const fn=new Function('scope',`with(scope) { ${code};return ${site.owner}; }`)({...compiler,...stubs});let reads=0;const input={get kind(){return ++reads===1?ts.SyntaxKind[kinds[i]]:ts.SyntaxKind.Unknown;},parent:undefined};let message;
  try{fn(input);}catch(error){message=error.message;}assert.equal(message,'Debug Failure. False expression.',site.id);results.push({id:site.id,assertionLevel:level,failed:true,message,kindReads:reads});
 }
}
fs.writeFileSync('/workspace/scratch/99u-guard-runtime-witness.json',JSON.stringify({source:'complete applied owners, actual existing tsc guards and Debug.assert, unrelated closure dependencies stubbed',results},null,2)+'\n');console.log('40 expected actual assertion failures across levels 0, 1, 2 and 3');
