// Supplemental frozen-queue certificate: stock checker, pristine upstream.
const ts = require('../../../api/node_modules/typescript');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const cp = require('node:child_process');
const zlib = require('node:zlib');
if (process.argv.length !== 4) throw Error('usage: prepare-ranked.cjs <upstream> <declarations>');
const [root, output] = process.argv.slice(2).map(x => path.resolve(x));
const pin = '050880ce59e30b356b686bd3144efe24f875ebc8';
const hash = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
if (cp.execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== pin) throw Error('pin mismatch');
cp.execFileSync('git', ['-C', root, 'diff', '--exit-code', 'HEAD', '--', 'src']);
const original = JSON.parse(fs.readFileSync(path.join(output, 'intersection-manifest.json')));
if (original.upstream_commit !== pin || Object.keys(original.declarations).length !== 78) throw Error('declarations mismatch');
for (const [file, digest] of Object.entries(original.declarations)) if (hash(path.join(output,file)) !== digest) throw Error('declaration drift '+file);
const inventoryPath = path.resolve(__dirname, '../../lane4/read-demand-pairs.json.gz');
const sitePath = path.resolve(__dirname, '../../lane4/read-demand-sites.json.gz');
const inventory = JSON.parse(zlib.gunzipSync(fs.readFileSync(inventoryPath)));
const demand = JSON.parse(zlib.gunzipSync(fs.readFileSync(sitePath)));
const sourceFile = path.join(root,'src/compiler/types.ts');
const utilities = path.join(root,'src/compiler/utilities.ts');
const program = ts.createProgram([sourceFile,utilities], {target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,strict:true,types:[]});
const checker = program.getTypeChecker();
const moduleExports = checker.getExportsOfModule(checker.getSymbolAtLocation(program.getSourceFile(sourceFile)));
const fields = type => checker.getPropertiesOfType(type).map(x => x.name).sort();
for (const [id,field,count] of [[9245,'argument',14],[8920,'operand',3],[40931,'operand',3],[8923,'operand',2]]) {
 const pair = inventory.find(x => x.receiver_type_id === id && x.field === field);
 if (!pair || pair.reads !== count) throw Error('inventory drift '+id);
 const sites = demand.filter(x => x.receiver_type_id === id && x.field === field).map(x => {
  const file = path.join(root,x.file), source = fs.readFileSync(file,'utf8');
  if (source.slice(x.start,x.end) !== x.text) throw Error('source span drift');
  return {file:x.file,start:x.start,end:x.end,text:x.text,line:x.line,column:x.column,source_sha256:hash(file)};
 });
 if (sites.length !== pair.reads) throw Error('site count drift');
 let receiver;
 if (id===9245) {
  receiver=checker.getDeclaredTypeOfSymbol(moduleExports.find(x => x.name===pair.type));
 } else {
  const witness=sites[0], source=program.getSourceFile(path.join(root,witness.file));
  let expression;
  function visit(node) {
   if (ts.isPropertyAccessExpression(node) && node.getStart(source)===witness.start && node.end===witness.end) expression=node;
   ts.forEachChild(node,visit);
  }
  visit(source);
  if (!expression || expression.name.text!==field) throw Error('original AST receiver missing '+id);
  receiver=checker.getTypeAtLocation(expression.expression);
  if (!checker.typeToString(receiver).includes('PrefixUnaryExpression')) throw Error('original receiver drift '+id);
 }
 const member=checker.getTypeOfSymbol(checker.getPropertyOfType(receiver,field));
 const result={upstream_commit:pin,inventory_sha256:hash(inventoryPath),sites_sha256:hash(sitePath),type_id:id,type:pair.type,field,read_count:count,sites,receiver_fields:fields(receiver),present_fields:fields(member),present_arms:member.isUnion()?member.types.map(fields):[]};
 const outputFile=id===9245?'ranked-intersection-manifest.json':`ranked-intersection-${id}-manifest.json`;
 fs.writeFileSync(path.join(output,outputFile),JSON.stringify(result,null,2)+'\n');
 console.log(`${id}.${field}: ${count} original spans; complete original receiver and member field sets; 78 declarations verified`);
}
