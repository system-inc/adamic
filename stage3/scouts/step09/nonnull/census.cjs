// Compiler-API inventory of postfix assertions, never prefix ! or declaration !.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const root = path.resolve(process.argv[2]);
const out = path.resolve(process.argv[3] || __dirname);
const membership = JSON.parse(fs.readFileSync(path.join(__dirname,'closure.json')));
const configPath = path.join(root,'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath,ts.sys.readFile);
assert.equal(read.error, undefined);
const config = ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configPath),undefined,configPath);
assert.equal(config.errors.length,0);
const options = {...config.options,noEmit:true,composite:false,isolatedDeclarations:false};
const own = ts.createProgram([path.join(root,'src/tsc/tsc.ts')],options);
assert.equal(own.getSemanticDiagnostics().length,0);
const program = ts.createProgram([path.join(root,'src/tsc/tsc.ts')],{...options,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true});
const checker = program.getTypeChecker();
const stockChecker = own.getTypeChecker();
const proven = require('./flow.cjs');
const files = program.getSourceFiles().filter(f=>f.fileName.startsWith(root+'/src/') && !f.isDeclarationFile).sort((a,b)=>a.fileName.localeCompare(b.fileName));
assert.deepEqual(files.map(f=>path.relative(root,f.fileName)).sort(),membership.stock.map(f=>f.file).sort());
const rows = [], fileRows = [];
for (const file of files) {
 const relative = path.relative(root,file.fileName);
 const sha256 = crypto.createHash('sha256').update(fs.readFileSync(file.fileName)).digest('hex');
 assert.equal(sha256,membership.stock.find(f=>f.file===relative).sha256,relative);
 let count=0, definite=0, prefix=0;
 const stockFile = own.getSourceFile(file.fileName);
 const stockNodes = new Map();
 function stockVisit(n) {if(ts.isNonNullExpression(n))stockNodes.set(n.getStart(stockFile)+':'+n.end,n);ts.forEachChild(n,stockVisit);}
 stockVisit(stockFile);
 function visit(n) {
  if (n.exclamationToken) definite++;
  if (ts.isPrefixUnaryExpression(n) && n.operator===ts.SyntaxKind.ExclamationToken) prefix++;
  if (ts.isNonNullExpression(n)) {
   count++;
   const start=n.getStart(file), pos=file.getLineAndCharacterOfPosition(start);
   const type=checker.getTypeAtLocation(n.expression);
   const original=stockNodes.get(start+':'+n.end);
   const stockType=stockChecker.getTypeAtLocation(original.expression);
   const status=proven(type,checker)?'flow-proven':'checked';
   rows.push({file:relative,line:pos.line+1,column:pos.character+1,start,end:n.end,expression:n.getText(file),operand_type:checker.typeToString(type),status,stock_project_status:proven(stockType,stockChecker)?'flow-proven':'checked',stock_project_type:stockChecker.typeToString(stockType)});
  }
  ts.forEachChild(n,visit);
 }
 visit(file);
 fileRows.push({file:relative,sha256,assertions:count,prefix_not:prefix,definite_assignment: definite});
}
const added = membership.adapted.filter(f=>!membership.stock.some(s=>s.file===f.file));
assert.deepEqual(added.map(f=>f.file),['src/compiler/hostErrors.ts']);
// Account for the added 81st file from its AST template, without inventing stock source.
const adapter = path.resolve(__dirname,'../../../adapt/47-host-errors/adapt.cjs');
const adapterFile = ts.createSourceFile(adapter,fs.readFileSync(adapter,'utf8'),ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);
let helper;
function findHelper(n){if(ts.isVariableDeclaration(n)&&n.name.getText(adapterFile)==='helper')helper=n.initializer;ts.forEachChild(n,findHelper);}
findHelper(adapterFile);assert.ok(helper && ts.isNoSubstitutionTemplateLiteral(helper));
assert.equal(crypto.createHash('sha256').update(helper.text).digest('hex'),added[0].sha256);
const helperFile=ts.createSourceFile('hostErrors.ts',helper.text,ts.ScriptTarget.Latest,true);
let helperAssertions=0;function countHelper(n){if(ts.isNonNullExpression(n))helperAssertions++;ts.forEachChild(n,countHelper);}countHelper(helperFile);
added[0].assertions=helperAssertions;added[0].stock_counterpart=false;
const counts = key => rows.reduce((a,r)=>(a[r[key]]=(a[r[key]]||0)+1,a),{});
const summary={typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',stock_files:files.length,adapted_files:membership.adapted.length,added_files_without_stock_counterpart:added,assertions:rows.length,counts:counts('status'),stock_project_counts:counts('stock_project_status'),stock_project_semantic_diagnostics:0,adamic_option_semantic_diagnostics:program.getSemanticDiagnostics().length,options:{strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true},definition:'API flow-proven is non-nullish under strict options; this is an erasure candidate, not an observed Adamic lowering of tsc.'};
fs.mkdirSync(out,{recursive:true});
for(const [name,data] of Object.entries({'sites.json':rows,'files.json':fileRows,'summary.json':summary}))fs.writeFileSync(path.join(out,name),JSON.stringify(data,null,2)+'\n');
fs.writeFileSync(path.join(out,'sites.tsv'),'file\tline\tcolumn\tstatus\toperand type\texpression\n'+rows.map(r=>[r.file,r.line,r.column,r.status,r.operand_type,r.expression.replace(/\s+/g,' ')].join('\t')).join('\n')+'\n');
console.log(JSON.stringify(summary,null,2));
