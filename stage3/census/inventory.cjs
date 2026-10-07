// Source inventory uses the pinned stock TypeScript parser and checker as an outside oracle.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require(process.env.CENSUS_TYPESCRIPT);
const source = path.resolve(process.argv[2]);
const output = path.resolve(process.argv[3]);
const configName = path.join(source, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configName, ts.sys.readFile);
if (read.error) throw Error(ts.flattenDiagnosticMessageText(read.error.messageText, '\n'));
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configName), undefined, configName);
if (config.errors.length) throw Error(JSON.stringify(config.errors));
const program = ts.createProgram(config.fileNames, config.options);
const checker = program.getTypeChecker();
const root = path.join(source, 'src/compiler');
const files = program.getSourceFiles().filter(f => f.fileName.startsWith(root + '/'));
const rules = [];
const sites = [];
const hosts = [];
const nodeDerived = [];
const shapes = [];
const stringLookups = [];
const functionExpandos = [];
const graph = [];
const inventory = [];
const sourceLines = new Map();
const lowerSource = fs.readFileSync(path.resolve(__dirname, '../../internal/lower/refusals.go'), 'utf8');
const syntax = new Map();
const operators = new Map();
for (const match of lowerSource.matchAll(/ast\.Kind(\w+):\s*\{"([^"]+)", "([^"]+)"\}/g)) {
 (match[1].endsWith('Token') || match[1] === 'InKeyword' ? operators : syntax).set(ts.SyntaxKind[match[1]], {reason:match[2],fix:match[3]});
}
function location(file, node) {
 const start = node.getStart(file);
 const p = file.getLineAndCharacterOfPosition(start);
 return {file:path.relative(source, file.fileName),line:p.line+1,column:p.character+1,end_line:file.getLineAndCharacterOfPosition(node.end).line+1,source:(sourceLines.get(file.fileName) || file.text.split(/\r?\n/))[p.line],start,end:node.end};
}
function add(file, node, reason, evidence, extra={}) { sites.push({...location(file,node),reason,evidence,...extra}); }
function symbol(node) {
 let value = checker.getSymbolAtLocation(node);
 if (value && value.flags & ts.SymbolFlags.Alias) value = checker.getAliasedSymbol(value);
 return value;
}
function systemMember(node) {
 const value = symbol(node);
 return value?.declarations?.some(d => ts.isInterfaceDeclaration(d.parent) && d.parent.name.text === 'System' && d.getSourceFile().fileName === path.join(root,'sys.ts'));
}
function isCall(node) {
 let parent = node.parent;
 while (parent && (ts.isParenthesizedExpression(parent) || ts.isNonNullExpression(parent))) parent = parent.parent;
 return parent && ts.isCallExpression(parent) && parent.expression.pos <= node.pos && parent.expression.end >= node.end;
}
const nodeRoots = new Set(['fsRealpath','_fs','_path','_os','_crypto','process','Buffer','global','require','__filename','__dirname','setTimeout','clearTimeout']);
const kindNames = new Set(['IfStatement','WhileStatement','DoStatement','ForStatement','ConditionalExpression']);
for (const file of files.sort((a,b)=>a.fileName.localeCompare(b.fileName))) {
 const rel = path.relative(source,file.fileName);
 sourceLines.set(file.fileName,file.text.split(/\r?\n/));
 inventory.push({file:rel,bytes:Buffer.byteLength(file.text),lines:file.text.split(/\r?\n/).length,sha256:crypto.createHash('sha256').update(fs.readFileSync(file.fileName)).digest('hex'),generated:rel.includes('.generated.')});
 function visit(node) {
  if (ts.isFunctionDeclaration(node) || ts.isClassDeclaration(node)) {
   let ancestor=node.parent;
   while (ancestor && !ts.isFunctionLike(ancestor)) ancestor=ancestor.parent;
   if (ancestor) add(file,node,ts.isFunctionDeclaration(node)?'a function inside a function (a closure)':'a class inside a function','statement NotYet branch');
  }
  const direct = syntax.get(node.kind);
  if (direct) add(file,node,direct.reason,'syntax refusal table',{fix:direct.fix});
  if (ts.isBinaryExpression(node) && operators.has(node.operatorToken.kind)) {
   const r=operators.get(node.operatorToken.kind); add(file,node.operatorToken,r.reason,'operator refusal table',{fix:r.fix});
  }
  if (node.kind === ts.SyntaxKind.AnyKeyword) add(file,node,'explicit any','inventory only: representation/use determines actual gate');
  if (ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) add(file,node,'type assertion sites','inventory only: includes supported const/upcasts/checked downcasts');
  if (ts.isTypeReferenceNode(node) && ts.isIdentifier(node.typeName) && node.typeName.text === 'Record' && node.typeArguments?.[0]?.kind === ts.SyntaxKind.StringKeyword) add(file,node,'Record<string, T>','design inventory');
  if (ts.isMappedTypeNode(node)) {
   const t=checker.getTypeAtLocation(node.typeParameter.constraint);
   if (t.flags & ts.TypeFlags.String) add(file,node,'string-keyed mapped type','design inventory');
  }
  if ((ts.isPropertyDeclaration(node) || ts.isVariableDeclaration(node)) && node.exclamationToken) add(file,node.exclamationToken,'a definite assignment assertion !','syntax refusal branch');
  if (ts.isObjectLiteralExpression(node)) node.properties.forEach((p,i)=> { if (ts.isSpreadAssignment(p) && i !== 0) add(file,p,'a spread after the first field','objectLiteral refusal'); });
  if (ts.isExportDeclaration(node)) add(file,node,'an ExportDeclaration','statement refusal');
  if (ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) {
   if (node.moduleSpecifier) {
    const value=symbol(node.moduleSpecifier);
    const target=value?.declarations?.find(ts.isSourceFile);
    graph.push({...location(file,node),target:target?path.relative(source,target.fileName):null,specifier:node.moduleSpecifier.text,type_only:!!node.importClause?.isTypeOnly || !!node.isTypeOnly});
   }
  }
  if (kindNames.has(ts.SyntaxKind[node.kind])) {
   const condition=ts.isForStatement(node)?node.condition:ts.isConditionalExpression(node)?node.condition:node.expression;
   if (condition) {
    const type=checker.getTypeAtLocation(condition);
    const parts=type.isUnion()?type.types:[type];
    if (parts.some(t=>!(t.flags & ts.TypeFlags.BooleanLike))) add(file,condition,'non-boolean control condition','typed candidate: upstream checker type, not a stage-0 observation',{type:checker.typeToString(type)});
   }
  }
  if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isPropertyAccessExpression(node.left) && ts.isIdentifier(node.left.expression)) {
   const decl=symbol(node.left.expression)?.valueDeclaration;
   if (decl && (ts.isFunctionDeclaration(decl) || ts.isFunctionExpression(decl) || ts.isArrowFunction(decl))) functionExpandos.push({...location(file,node),object:node.left.expression.text,property:node.left.name.text,declaration:location(decl.getSourceFile(),decl)});
   if (decl && ts.isVariableDeclaration(decl) && decl.initializer) {
    let initial=decl.initializer;
    while (ts.isAsExpression(initial) || ts.isParenthesizedExpression(initial)) initial=initial.expression;
    if ((ts.isObjectLiteralExpression(initial) && !initial.properties.some(p=>p.name && p.name.getText(file)===node.left.name.text)) || (ts.isArrayLiteralExpression(initial) && node.left.name.text !== 'length')) shapes.push({...location(file,node),initial_kind:ts.isArrayLiteralExpression(initial)?'array':'object',object:node.left.expression.text,property:node.left.name.text,declaration:location(decl.getSourceFile(),decl),declared_property:!!checker.getPropertyOfType(checker.getTypeAtLocation(node.left.expression),node.left.name.text)});
   }
  }
  if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) {
   const member=ts.isPropertyAccessExpression(node)?node.name.text:node.argumentExpression?.getText(file);
   if (ts.isElementAccessExpression(node)) {
    const indexType=checker.getIndexTypeOfType(checker.getTypeAtLocation(node.expression),ts.IndexKind.String);
    if (indexType) stringLookups.push({...location(file,node),receiver_type:checker.typeToString(checker.getTypeAtLocation(node.expression)),key_type:checker.typeToString(checker.getTypeAtLocation(node.argumentExpression)),value_type:checker.typeToString(indexType)});
   }
   const recv=node.expression.getText(file);
   const propertySymbol=symbol(ts.isPropertyAccessExpression(node)?node.name:node);
   const nodeDeclaration=propertySymbol?.declarations?.find(d=>d.getSourceFile().fileName.includes('/node_modules/@types/node/'));
   if (nodeDeclaration) nodeDerived.push({...location(file,node),receiver:recv,member,call:isCall(node),declared_owner:nodeDeclaration.parent.name?.getText() || ts.SyntaxKind[nodeDeclaration.parent.kind],declared_in:nodeDeclaration.getSourceFile().fileName.split('/node_modules/')[1]});
   if (systemMember(ts.isPropertyAccessExpression(node)?node.name:node)) hosts.push({...location(file,node),family:'System',receiver:recv,member,call:isCall(node)});
   // Record maximal Node member chains only, preserving reads as well as calls.
   let baseNode=node.expression;
   while (ts.isPropertyAccessExpression(baseNode) || ts.isElementAccessExpression(baseNode) || ts.isParenthesizedExpression(baseNode) || ts.isAsExpression(baseNode) || ts.isNonNullExpression(baseNode)) baseNode=baseNode.expression;
   const base=ts.isIdentifier(baseNode)?baseNode.text:'';
   if (nodeRoots.has(base) && !ts.isPropertyAccessExpression(node.parent) && !ts.isElementAccessExpression(node.parent)) hosts.push({...location(file,node),family:base,receiver:recv,member,call:isCall(node)});
   if (symbol(ts.isPropertyAccessExpression(node)?node.name:node)?.declarations?.some(d => ts.isInterfaceDeclaration(d.parent) && ['Performance','PerformanceTime'].includes(d.parent.name.text) && d.getSourceFile().fileName === path.join(root,'performanceCore.ts'))) hosts.push({...location(file,node),family:'performance',receiver:recv,member,call:isCall(node)});
  }
  if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && nodeRoots.has(node.expression.text)) hosts.push({...location(file,node),family:node.expression.text,receiver:'',member:'(call)',call:true,argument:node.arguments[0]?.getText(file)});
  if (ts.isInterfaceDeclaration(node) && node.name.text === 'System') {
   for(const member of node.members) rules.push({name:member.name.getText(file),optional:!!member.questionToken,...location(file,member)});
  }
  ts.forEachChild(node,visit);
 }
 visit(file);
}
const diagnostics = ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n'),...(d.file ? {file:path.relative(source,d.file.fileName),...(()=>{const p=d.file.getLineAndCharacterOfPosition(d.start);return {line:p.line+1,column:p.character+1};})()}: {})}));
fs.mkdirSync(output,{recursive:true});
for(const [name,data] of Object.entries({files:inventory,sites,host_sites:hosts,node_derived_sites:nodeDerived,system_contract:rules,shape_additions:shapes,string_lookups:stringLookups,function_expandos:functionExpandos,module_edges:graph,upstream_diagnostics:diagnostics,upstream_config:{file:'src/compiler/tsconfig.json',options:config.options,roots:config.fileNames.map(f=>path.relative(source,f)),typescript:ts.version}})) fs.writeFileSync(path.join(output,name+'.json'),JSON.stringify(data,null,2)+'\n');
console.log(JSON.stringify({files:files.length,sites:sites.length,hosts:hosts.length,shape_additions:shapes.length,upstream_diagnostics:diagnostics.length}));
