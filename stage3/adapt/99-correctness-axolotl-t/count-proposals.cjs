const fs=require('fs'),path=require('path'),ts=require('typescript');
const root=process.argv[2];const manifest=[];
function file(name,fn){const p=path.join(root,'src/compiler',name),s=fs.readFileSync(p,'utf8');const after=fn(s);fs.writeFileSync(p,after);}
const emit=(id,condition)=>{manifest.push({id,condition});return `(globalThis as any).__debugVisibleCensus(${JSON.stringify(id)}, ${condition})`};
const method=(v,name)=>`typeof ${v} === "object" && ${v} !== null && "${name}" in ${v} && typeof (${v} as any).${name} === "function"`;
function replace(s,a,b){if(s.split(a).length!==2)throw Error('expected unique '+a);return s.replace(a,b)}
file('debug.ts',s=>{
 const expressions=[['CW144','this.source','DebugType','__debugTypeToString'],['CW145','this.target','DebugType','__debugTypeToString'],['CW148','s','DebugType','__debugTypeToString'],['CW149','t()','DebugType','__debugTypeToString'],['CW150','this.mapper1','unknown as DebugTypeMapper','__debugToString'],['CW151','this.mapper2','unknown as DebugTypeMapper','__debugToString']];
 for(const[id,v,type,m]of expressions){if(v==='t()'){s=replace(s,`(t() as ${type}).${m}()`,`((value => (${emit(id,method('value',m))}, (value as ${type}).${m}()))(t()))`)}else s=replace(s,`(${v} as ${type}).${m}()`,`(${emit(id,method(v,m))}, (${v} as ${type}).${m}())`)}
 for(const[id,v]of [['CW146','this.sources'],['CW147','this.targets']]){
 const code=`((values => { ${emit(id+'-array',`${v==='this.targets'?'values === undefined || ':''}Array.isArray(values)`)}; if (values) for (const value of values) ${emit(id+'-element',method('value','__debugTypeToString'))}; return values; })(${v}) as readonly DebugType[])`;
 s=replace(s,`${v} as readonly DebugType[]`,code);
 }
 const call='text = sourceFile ? getSourceTextOfNodeFromSourceFile(sourceFile, parseNode, includeTrivia) : "";';
 s=replace(s,call,`if (sourceFile) ${emit('CW143','sourceFile.kind === SyntaxKind.SourceFile')};\r\n                                ${call}`);
 const old='return !!(f.flags & FlowFlags.Label) && !!(f as FlowLabel).antecedent;';
 s=replace(s,old,`if (f.flags & FlowFlags.Label) ${emit('CW154','"antecedent" in f && (f.antecedent === undefined || Array.isArray(f.antecedent))')};\r\n            ${old}`);
 return s;
});
const typeChecks={
 DeclarationWithTypeParameterChildren:['CallSignature','ConstructSignature','MethodSignature','IndexSignature','FunctionType','ConstructorType','JSDocFunctionType','ClassDeclaration','ClassExpression','InterfaceDeclaration','TypeAliasDeclaration','JSDocTemplateTag','FunctionDeclaration','MethodDeclaration','Constructor','GetAccessor','SetAccessor','FunctionExpression','ArrowFunction'],
 DeclarationWithTypeParameters:['CallSignature','ConstructSignature','MethodSignature','IndexSignature','FunctionType','ConstructorType','JSDocFunctionType','ClassDeclaration','ClassExpression','InterfaceDeclaration','TypeAliasDeclaration','JSDocTemplateTag','FunctionDeclaration','MethodDeclaration','Constructor','GetAccessor','SetAccessor','FunctionExpression','ArrowFunction','JSDocCallbackTag','JSDocTypedefTag','JSDocSignature'],
 HasInferredType:['Parameter','PropertySignature','PropertyDeclaration','BindingElement','PropertyAccessExpression','ElementAccessExpression','BinaryExpression','VariableDeclaration','ExportAssignment','PropertyAssignment','ShorthandPropertyAssignment','JSDocParameterTag','JSDocPropertyTag'],
 PrimitiveLiteral:['TrueKeyword','FalseKeyword','NumericLiteral','StringLiteral','NoSubstitutionTemplateLiteral','BigIntLiteral'],
 WithIsolatedDeclarationDiagnostic:['GetAccessor','SetAccessor','ShorthandPropertyAssignment','SpreadAssignment','ComputedPropertyName','ArrayLiteralExpression','SpreadElement','FunctionDeclaration','FunctionExpression','ArrowFunction','MethodDeclaration','ConstructSignature','BindingElement','VariableDeclaration','PropertyDeclaration','Parameter','PropertyAssignment','ClassExpression'],
 'ArrowFunction | FunctionExpression':['ArrowFunction','FunctionExpression']
};
for(const name of ['utilities.ts','expressionToTypeNode.ts','transformers/declarations/diagnostics.ts'])file(name,s=>{
 for(const[type,kinds]of Object.entries(typeChecks)){
 const call=`Debug.type<${type}>(node);`;if(!s.includes(call))continue;
 let c=kinds.map(k=>`node.kind === SyntaxKind.${k}`).join(' || ');
 if(type==='PrimitiveLiteral')c+=` || (node.kind === SyntaxKind.PrefixUnaryExpression && (((node as any).operator === SyntaxKind.PlusToken && (node as any).operand.kind === SyntaxKind.NumericLiteral) || ((node as any).operator === SyntaxKind.MinusToken && ((node as any).operand.kind === SyntaxKind.NumericLiteral || (node as any).operand.kind === SyntaxKind.BigIntLiteral))))`;
 s=replace(s,call,`${emit(type,c)}; ${call}`);
 }return s;
});
file('moduleSpecifiers.ts',s=>{const call='Debug.type<ModuleResolutionHost>(host);';return replace(s,call,`${emit('ModuleResolutionHost','typeof host.readFile === "function" && (!("getDirectories" in host) || host.getDirectories === undefined || typeof host.getDirectories === "function")')}; ${call}`)});
fs.writeFileSync('/workspace/scratch/99t-counter-manifest.json',JSON.stringify(manifest,null,2)+'\n');
