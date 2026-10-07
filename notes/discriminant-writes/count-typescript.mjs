// Run with: node count-typescript.mjs <TypeScript v6.0.3 checkout> <typescript npm package directory> <independent JSON>
// Instrument the stock checker in memory to expose its narrowing test. No substitute definition.
import fs from 'node:fs';
import path from 'node:path';
import Module from 'node:module';
import { createHash } from 'node:crypto';
const [root, packageRoot] = process.argv.slice(2).map(value => path.resolve(value));
const library = path.join(packageRoot, 'lib/typescript.js');
const original = fs.readFileSync(library, 'utf8');
const anchor = 'getTypeOfSymbolAtLocation: (symbol, locationIn) => {';
if (original.split(anchor).length !== 2) throw Error('checker export anchor changed');
const instrumented = new Module(library);
instrumented.filename = library;
instrumented.paths = Module._nodeModulePaths(path.dirname(library));
instrumented._compile(original.replace(anchor, '__adamicIsDiscriminantProperty: isDiscriminantProperty,\n    __adamicPropertyName: getAccessedPropertyName,\n    ' + anchor), library);
const ts = instrumented.exports;
if (ts.version !== '6.0.3') throw Error(`want 6.0.3, got ${ts.version}`);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configPath, ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath), undefined, configPath);
const program = ts.createProgram(parsed.fileNames, parsed.options);
const checker = program.getTypeChecker();
const diagnostics=ts.getPreEmitDiagnostics(program);
const sources = program.getSourceFiles().filter(s => s.fileName.startsWith(path.join(root, 'src/compiler/')) && !s.isDeclarationFile && !s.fileName.includes('.generated.'));
const independent = JSON.parse(fs.readFileSync(process.argv[4], 'utf8'));
const reviews = JSON.parse(fs.readFileSync(process.argv[5] || new URL('./construction-review.json', import.meta.url), 'utf8'));
const eventKey = event => `${event.file}:${event.line}:${event.column}:${event.property}`;
const reviewByKey = new Map(reviews.map(event => [eventKey(event), event]));
if (reviewByKey.size !== reviews.length) throw Error('duplicate construction review event');
const catalogueDeclarations = new Map();
for (const witness of independent.catalogue) for (const member of witness.members) for (const declaration of member.declarations) {
 const key = `${declaration.file}:${declaration.start}:${declaration.end}`;
 if (!catalogueDeclarations.has(key)) catalogueDeclarations.set(key, new Set());
 catalogueDeclarations.get(key).add(witness.id);
}
const fields = new Map(), seen = new Set(), witnessedUnions = new Map();
function note(type) {
 if (!type || seen.has(type)) return;
 seen.add(type);
 if (type.flags & ts.TypeFlags.TypeParameter) note(checker.getBaseConstraintOfType(type));
 for (const signature of checker.getSignaturesOfType(type, ts.SignatureKind.Call)) note(checker.getReturnTypeOfSignature(signature));
 if (type.isUnion()) {
  const text = checker.typeToString(type, undefined, ts.TypeFormatFlags.NoTruncation);
  if (!witnessedUnions.has(text)) witnessedUnions.set(text, []);
  witnessedUnions.get(text).push(type);
 }
 if (type.isUnion()) for (const member of type.types) {
  for (const property of checker.getPropertiesOfType(member)) {
   if (checker.__adamicIsDiscriminantProperty(type, property.name)) {
    if (!fields.has(property.name)) fields.set(property.name, new Set());
    fields.get(property.name).add(member);
   }
  }
  note(member);
 }
}
function visit(node, callback) { callback(node); ts.forEachChild(node, child => visit(child, callback)); }
const nodes = new Map();
const sourceHashes = new Map(sources.map(source => [path.relative(root, source.fileName), createHash('sha256').update(fs.readFileSync(source.fileName)).digest('hex')]));
for (const source of sources) visit(source, node => {
 nodes.set(`${path.relative(root,source.fileName)}:${node.getStart(source)}:${node.end}`, node);
 if (ts.isExpressionNode(node) || ts.isTypeNode(node) || ts.isFunctionLike(node) || ts.isVariableDeclaration(node) || ts.isParameter(node) || ts.isPropertyDeclaration(node) || ts.isPropertySignature(node)) note(checker.getTypeAtLocation(node));
});
function members(holder, name) {
 if (holder.flags & ts.TypeFlags.TypeParameter) holder = checker.getBaseConstraintOfType(holder);
 if (!holder) return [];
 if (holder.isUnion() && checker.__adamicIsDiscriminantProperty(holder,name)) return holder.types;
 const possible = [...fields.get(name) || []].filter(member => checker.isTypeAssignableTo(member,holder));
 if (possible.length && !possible.includes(holder)) possible.push(holder);
 return possible;
}
const writes = independent.writes.map(event => {
 const node = nodes.get(`${event.file}:${event.start}:${event.end}`);
 if (!node) throw Error(`missing event ${event.file}:${event.line}`);
 let receiver, value;
 if (ts.isBinaryExpression(node)) {
  receiver = node.left.expression;
  value = checker.getTypeAtLocation(node.operatorToken.kind === ts.SyntaxKind.EqualsToken ? node.right : node);
 } else if (ts.isPrefixUnaryExpression(node) || ts.isPostfixUnaryExpression(node)) { receiver=node.operand.expression; value=checker.getTypeAtLocation(node); }
 const holder = receiver && checker.getTypeAtLocation(receiver);
 const possible = holder ? members(holder,event.property) : [];
 const failed = possible.filter(member => { const p=checker.getPropertyOfType(member,event.property); return !p || !checker.isTypeAssignableTo(value,checker.getTypeOfSymbol(p)); });
 const inside = event.construction === 'inside';
 const refused = !inside && failed.length>0;
 const plainRefused = !inside && possible.length>0;
 return {...event, plain:plainRefused?'refused':'accepted', ruled:refused?'refused':'accepted', actual_value_type:value && checker.typeToString(value), possible_members:possible.map(m=>checker.typeToString(m)), incompatible_members:failed.map(m=>checker.typeToString(m)), comparison:inside?'same construction classification; accepted initialization':!possible.length?'exact narrowing test has no compatible witnessed member for this receiver':refused?'exact discriminant; written type fails a possible member':'exact discriminant; written type fits every possible member'};
});
const independentAccesses=new Set(writes.filter(w=>w.target).map(w=>`${w.file}:${w.start}:${w.end}:${w.property}`));
const additional=[];
for(const source of sources) visit(source,node=>{
 if(!ts.isPropertyAccessExpression(node)&&!ts.isElementAccessExpression(node))return;
 if(!ts.isAssignmentTarget(node))return;
 const name=checker.__adamicPropertyName(node); if(!name)return;
 const holder=checker.getTypeAtLocation(node.expression), possible=members(holder,name);if(!possible.length)return;
 const location=source.getLineAndCharacterOfPosition(node.getStart(source)), file=path.relative(root,source.fileName), line=location.line+1;

 let parent=node.parent;while(ts.isParenthesizedExpression(parent))parent=parent.parent;
 let value;
 if(ts.isBinaryExpression(parent))value=checker.getTypeAtLocation(parent.operatorToken.kind===ts.SyntaxKind.EqualsToken?parent.right:parent);
 else if(ts.isPrefixUnaryExpression(parent)||ts.isPostfixUnaryExpression(parent))value=checker.getTypeAtLocation(parent);
 if(!value)return;
 if (independentAccesses.has(`${file}:${parent.getStart(source)}:${parent.end}:${name}`)) return;
 const failed=possible.filter(m=>{const p=checker.getPropertyOfType(m,name);return !p||!checker.isTypeAssignableTo(value,checker.getTypeOfSymbol(p));});
 const symbol = checker.getPropertyOfType(holder, name);
 const declarationKeys = new Set();
 if (symbol) for (const rootSymbol of [symbol, ...checker.getRootSymbols(symbol)]) for (const declaration of rootSymbol.declarations || []) {
  declarationKeys.add(`${path.relative(root, declaration.getSourceFile().fileName)}:${declaration.getStart()}:${declaration.end}`);
 }
 const rootWitnesses = [...new Set([...declarationKeys].flatMap(key => [...catalogueDeclarations.get(key) || []]))];
 const receiverToMember = rootWitnesses.map(id => {
  const witness = independent.catalogue[id-1];
  const unions = witnessedUnions.get(witness.union) || [];
  if (!unions.length) throw Error(`cannot re-resolve independent union witness ${id}`);
  return {id, union:witness.union, receiver_assignable_to_member:unions.some(union=>union.types.some(member=>checker.isTypeAssignableTo(holder,member)))};
 });
 if (receiverToMember.some(witness=>witness.receiver_assignable_to_member)) throw Error(`unexplained independent selection difference ${file}:${line}:${name}`);
 const scopeDifference = rootWitnesses.length
  ? 'Root declarations have catalogue witnesses, but the independent receiver-to-member filter does not admit this broader view; the compiler checks member-to-receiver compatibility.'
  : 'Written property roots have no declaration-indexed catalogue witness; the compiler admits the discriminant through a structurally compatible union member.';
 const event = {file,line,column:location.character+1,property:name,start:parent.getStart(source),end:parent.end,receiver_type:checker.typeToString(holder),value_type:checker.typeToString(value),incompatible_members:failed.map(m=>checker.typeToString(m)),declaration_keys:[...declarationKeys],declaration_witnesses:rootWitnesses,receiver_to_member:receiverToMember,scope_difference:scopeDifference};
 const review = reviewByKey.get(eventKey(event));
 if (!review) throw Error(`construction review missing ${eventKey(event)}`);
 const sourceLine = source.text.split(/\r?\n/)[line-1].trim();
 if (sourceHashes.get(file) !== review.source_sha256) throw Error(`review source hash changed ${eventKey(event)}`);
 if (sourceLine !== review.source_line) throw Error(`review source changed ${eventKey(event)}`);
 const inside = review.construction === 'inside';
 additional.push({...event,construction:review.construction,explanation:review.explanation,evidence:review.evidence,source_sha256:review.source_sha256,ruled:!inside && failed.length?'refused':'accepted',plain:!inside?'refused':'accepted'});

});

if (additional.length !== reviewByKey.size) throw Error('construction review and extra discovery event sets differ');
const suppliedKeys = new Set(writes.map(w=>`${w.file}:${w.start}:${w.end}:${w.property}:${w.form}`));
if (suppliedKeys.size !== writes.length) throw Error('duplicate supplied event');
if (writes.length !== 437 || sources.length !== 77) throw Error('independent census population changed');

if(diagnostics.length) throw Error(ts.formatDiagnosticsWithColorAndContext(diagnostics,{getCurrentDirectory:()=>root,getCanonicalFileName:x=>x,getNewLine:()=>"\n"}));
const count = list => ({accepted:list.filter(w=>w.ruled==='accepted').length,refused:list.filter(w=>w.ruled==='refused').length,plain_refused:list.filter(w=>w.plain==='refused').length});
console.log(JSON.stringify({version:ts.version,source_commit:independent.source_commit,sourceFiles:sources.length,diagnostics:diagnostics.length,count:count(writes),outside:count(writes.filter(w=>w.construction==='outside local construction')),uncertain:count(writes.filter(w=>w.construction==='not established')),extra_count:count(additional),construction:{supplied:writes.reduce((a,w)=>(a[w.construction]=(a[w.construction]||0)+1,a),{}),additional:additional.reduce((a,w)=>(a[w.construction]=(a[w.construction]||0)+1,a),{})},combined:count([...writes,...additional]),additional,writes},null,2));
