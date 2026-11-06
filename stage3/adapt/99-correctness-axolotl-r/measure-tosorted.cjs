'use strict';
const fs=require('node:fs'),ts=require('typescript'),crypto=require('node:crypto');
const {compiler,nodes}=require('../47-host-errors/adapt.cjs');
const tree=process.argv[2],base=compiler(tree,ts),name=tree+'/src/compiler/core.ts',original=fs.readFileSync(name,'utf8');
const candidates=['emptyArray','emptyArray as readonly never[]','emptyArray as readonly T[]','emptyArray as ReadonlyArray<T>','<readonly T[]>emptyArray','emptyArray as readonly []','emptyArray satisfies readonly T[]','(emptyArray as readonly never[]) satisfies readonly T[]','conditional satisfies readonly T[]'];
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
const emit=(text,removeComments)=>ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
const results=[];
for(const spelling of candidates){
 const text=spelling==='conditional satisfies readonly T[]' ? original.replace('array.slice().sort(comparer)) as unknown as SortedReadonlyArray<T>', 'array.slice().sort(comparer)) satisfies readonly T[] as unknown as SortedReadonlyArray<T>') : original.replace('array.length === 0 ? emptyArray : array.slice().sort(comparer)','array.length === 0 ? '+spelling+' : array.slice().sort(comparer)');
 const host=ts.createCompilerHost(base.getCompilerOptions(),true),get=host.getSourceFile.bind(host);
 host.getSourceFile=(file,lang,...args)=>file===name?ts.createSourceFile(file,text,lang,true):get(file,lang,...args);
 const program=ts.createProgram(base.getRootFileNames(),base.getCompilerOptions(),host),checker=program.getTypeChecker(),file=program.getSourceFile(name);
 const owner=file.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='toSorted');
 const conditional=nodes(ts,owner,ts.isConditionalExpression)[0];
 const show=t=>t?checker.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation):null;
 const describe=n=>({text:n.getText(file),type:show(checker.getTypeAtLocation(n)),contextualType:show(checker.getContextualType(n)),...(()=>{const p=file.getLineAndCharacterOfPosition(n.getStart(file));return {line:p.line+1,column:p.character+1};})()});
 const casts=nodes(ts,conditional.whenTrue,n=>ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)).map(n=>({text:n.getText(file),source:show(checker.getTypeAtLocation(n.expression)),target:show(checker.getTypeAtLocation(n)),assignable:checker.isTypeAssignableTo(checker.getTypeAtLocation(n.expression),checker.getTypeAtLocation(n))}));
 const diagnostics=program.getSemanticDiagnostics(file).filter(d=>d.start>=owner.pos&&d.start<owner.end).map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n'));
 results.push({spelling,conditional:describe(conditional),trueArm:describe(conditional.whenTrue),falseArm:describe(conditional.whenFalse),emptyArrayRead:describe(nodes(ts,conditional.whenTrue,n=>ts.isIdentifier(n)&&n.text==='emptyArray')[0]),casts,diagnostics,javascript:[false,true].map(removeComments=>{const a=emit(original,removeComments),b=emit(text,removeComments);return {removeComments,byteIdentical:a===b,sha256:hash(b),bytes:Buffer.byteLength(b)};})});
}
console.log(JSON.stringify({typescript:ts.version,tree,results},null,2));
