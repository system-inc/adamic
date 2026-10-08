// Inventory the actual adapted tree. This is syntax/type evidence, not executable IR.
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const [root, output] = process.argv.slice(2);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath, ts.sys.readFile);
assert.equal(read.error, undefined);
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
assert.equal(config.errors.length, 0);
const program = ts.createProgram(config.fileNames, config.options), checker = program.getTypeChecker();
const parts = type => type.isUnion() ? type.types : [type];
const interfaces = type => parts(type).every(part => part.symbol?.declarations?.some(ts.isInterfaceDeclaration));
function descendant(target, source, seen = new Set(), first = true) {
 if (target.symbol && target.symbol === source.symbol) return !first;
 if (seen.has(target)) return false;
 seen.add(target);
 if (!(target.flags & ts.TypeFlags.Object) || !(target.objectFlags & (ts.ObjectFlags.ClassOrInterface | ts.ObjectFlags.Reference))) return false;
 return (checker.getBaseTypes(target) || []).some(base => descendant(base, source, seen, false));
}
const rows = [];
for (const file of program.getSourceFiles()) {
 const relative = path.relative(root, file.fileName).replaceAll('\\', '/');
 if (file.isDeclarationFile || relative.includes('.generated.') || !relative.startsWith('src/compiler/')) continue;
 function visit(node) {
  if (ts.isAsExpression(node)) {
   const source = checker.getTypeAtLocation(node.expression), target = checker.getTypeFromTypeNode(node.type);
   const downcast = !checker.isTypeAssignableTo(source, target) && checker.isTypeAssignableTo(target, source);
   let kind = 'other', tag;
   if (downcast && interfaces(source) && interfaces(target)) {
    for (const property of checker.getPropertiesOfType(target)) {
     const before = checker.getPropertyOfType(source, property.name);
     const type = checker.getTypeOfSymbolAtLocation(property, node);
     const broad = before && checker.getTypeOfSymbolAtLocation(before,node);
     if (before && !(before.flags & ts.SymbolFlags.Optional) && !(property.flags & ts.SymbolFlags.Optional) && !checker.isTypeAssignableTo(broad,type) && checker.isTypeAssignableTo(type,broad) && parts(type).every(part => part.flags & (ts.TypeFlags.StringLiteral | ts.TypeFlags.NumberLiteral | ts.TypeFlags.BooleanLiteral))) {tag = property.name;break;}
    }
    if (tag && !source.isUnion() && parts(target).every(member => descendant(member, source))) kind = 'tagged';
    else if (!tag) kind = 'untagged';
   }
   const location = file.getLineAndCharacterOfPosition(node.getStart(file));
   rows.push({file:relative,line:location.line+1,column:location.character+1,start:node.getStart(file),end:node.end,text:node.getText(file),kind,tag,source_type:checker.typeToString(source),target_type:checker.typeToString(target)});
  }
  ts.forEachChild(node, visit);
 }
 visit(file);
}
fs.writeFileSync(output, JSON.stringify(rows, null, 2)+'\n');
console.log(JSON.stringify({casts:rows.length,tagged:rows.filter(r=>r.kind==='tagged').length,untagged:rows.filter(r=>r.kind==='untagged').length,limits:'Current adapted checker-type classification. Not the old immutable 2936-site denominator and not successful lowering. Tagged requires declared ancestry and a common finite field; untagged requires an interface downcast without such a field.'},null,2));
