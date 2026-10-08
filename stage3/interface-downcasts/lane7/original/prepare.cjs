// Verify pristine official upstream and import complete emitted declarations.
// The shared adapter owns declaration emission; no cohere file is copied.
const ts = require('../../../api/node_modules/typescript');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const cp = require('node:child_process');
const args = process.argv.slice(2);
if (args.length !== 2) throw Error('usage: prepare.cjs <official-upstream> <declarations>');
const [root, declarations] = args.map(p => path.resolve(p));
const pin = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../../../source.json')));
if (cp.execFileSync('git',['-C',root,'rev-parse','HEAD'],{encoding:'utf8'}).trim() !== pin.commit) throw Error('upstream pin mismatch');
cp.execFileSync('git',['-C',root,'diff','--exit-code','HEAD','--','src']);
const hash = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
const emitted = JSON.parse(fs.readFileSync(path.join(declarations,'original-manifest.json')));
if (emitted.upstream_commit !== pin.commit) throw Error('declaration pin mismatch');
for (const [file,digest] of Object.entries(emitted.declarations)) if (hash(path.join(declarations,file)) !== digest) throw Error('declaration drift: '+file);
const witnesses = JSON.parse(fs.readFileSync(path.resolve(__dirname,'../original-read-witnesses.json'))).witnesses;
for (const site of witnesses) {
 const file = path.join(root,site.file);
 const text = fs.readFileSync(file,'utf8');
 if (hash(file) !== site.source_sha256 || text.slice(site.start,site.end) !== site.text) throw Error('read-site drift: '+JSON.stringify(site));
}
const options = {target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,strict:true,types:[]};
const program = ts.createProgram([path.join(root,'src/compiler/types.ts')],options);
const checker = program.getTypeChecker();
const source = program.getSourceFile(path.join(root,'src/compiler/types.ts'));
const moduleExports = checker.getExportsOfModule(checker.getSymbolAtLocation(source));
const fields = {};
for (const name of ['SymbolTracker','ModuleSpecifierResolutionHost','GeneratedIdentifier','EmitNode','AutoGenerateInfo','Identifier']) {
 const symbol = moduleExports.find(symbol=>symbol.name===name);
 if (!symbol) throw Error('missing original type '+name);
 fields[name] = checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(field=>field.name).sort();
}
const pairs = [];
for (const [id,name,field] of [[10236,'SymbolTracker','moduleResolverHost'],[7612,'GeneratedIdentifier','emitNode']]) {
 const symbol = moduleExports.find(symbol=>symbol.name===name);
 const receiver = checker.getDeclaredTypeOfSymbol(symbol);
 const member = checker.getPropertyOfType(receiver,field);
 const declared = checker.getTypeOfSymbolAtLocation(member,member.valueDeclaration || member.declarations[0]);
 const sites = witnesses.filter(site=>site.type_id===id && site.field===field);
 pairs.push({type_id:id,type:name,field,read_count:sites.length,sites,
   declared_type:checker.typeToString(declared),
   present_fields:checker.getPropertiesOfType(checker.getNonNullableType(declared)).map(field=>field.name).sort()});
}
fs.writeFileSync(path.join(declarations,'intersection-manifest.json'),JSON.stringify({upstream_commit:pin.commit,declarations:emitted.declarations,fields,pairs},null,2)+'\n');
console.log(`Complete declaration field sets verified; original spans ${witnesses.length}; pairs ${pairs.map(pair=>pair.read_count).join('+')}`);
