const ts = require('typescript');
const fs = require('node:fs');
const path = require('node:path');
const root = process.argv[2];
const sites = JSON.parse(fs.readFileSync(process.argv[3], 'utf8'));
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configPath, ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath));
parsed.options.typeRoots = [path.join(process.env.NODE_PATH, '@types')];
const program = ts.createProgram(parsed.fileNames, parsed.options);
const checker = program.getTypeChecker();
function families(type) {
 if (type.isUnion()) return [...new Set(type.types.flatMap(families))];
 if (type.flags & ts.TypeFlags.Null) return ['null'];
 if (type.flags & ts.TypeFlags.Undefined) return ['undefined'];
 if (type.flags & ts.TypeFlags.BooleanLike) return ['boolean'];
 if (type.flags & ts.TypeFlags.NumberLike) return ['number'];
 if (type.flags & ts.TypeFlags.StringLike) return ['string'];
 if (checker.getSignaturesOfType(type, ts.SignatureKind.Call).length) return ['function'];
 if (checker.isArrayType(type) || checker.isTupleType(type)) return ['array'];
 return ['object'];
}
const output = sites.map(row => {
 const [file,line,column] = row.where.split(':');
 const source = program.getSourceFile(path.join(root,file));
 const offset = ts.getPositionOfLineAndCharacter(source,+line-1,+column-1);
 const candidates = [];
 function visit(node) {
  let condition;
  if (ts.isIfStatement(node) || ts.isWhileStatement(node) || ts.isDoStatement(node)) condition=node.expression;
  if (ts.isForStatement(node) || ts.isConditionalExpression(node)) condition=node.condition;
  if (ts.isPrefixUnaryExpression(node) && node.operator===ts.SyntaxKind.ExclamationToken) condition=node.operand;
  if (condition && condition.getStart(source)===offset) candidates.push(condition);
  ts.forEachChild(node,visit);
 }
 visit(source);
 if(candidates.length!==1) throw Error(row.where+': condition matches '+candidates.length+' line '+source.text.split(/\r?\n/)[+line-1]);
 const node=candidates[0], type=checker.getTypeAtLocation(node);
 return {...row,expression:node.getText(source),context:ts.SyntaxKind[node.parent.kind],type:checker.typeToString(type),families:families(type),enum_constant:ts.isBinaryExpression(node) ? checker.getConstantValue(node.right) : undefined,source_sha256:require('node:crypto').createHash('sha256').update(source.text).digest('hex')};
});
fs.writeFileSync(process.argv[4], JSON.stringify({typescript:ts.version,checker_diagnostics:ts.getPreEmitDiagnostics(program).length,sites:output},null,2)+'\n');
console.log(JSON.stringify({sites:output.length,null_sites:output.filter(r=>r.families.includes('null'))}));
