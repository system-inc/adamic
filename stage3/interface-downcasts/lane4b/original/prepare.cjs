// Generate declaration-only input from pinned upstream, never from cohere.
const ts = require('../../../api/node_modules/typescript');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const cp = require('node:child_process');
const [root, out] = process.argv.slice(2).map(p => path.resolve(p));
if (!root || !out) throw Error('usage: prepare.cjs <pinned-TypeScript-root> <new-output>');
const pin = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../../../source.json')));
if (cp.execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== pin.commit) throw Error('upstream pin mismatch');
cp.execFileSync('git', ['-C', root, 'diff', '--exit-code', 'HEAD', '--', 'src']);
const ledger = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../resume/lazy-pair-progress.json')));
const pairs = ledger.rows.slice(0,4);
for (const pair of pairs) for (const site of pair.sites) {
 const text = fs.readFileSync(path.join(root, site.file),'utf8');
 if (text.slice(site.start,site.end) !== site.text) throw Error('read-site drift: '+JSON.stringify(site));
}
const options={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,declaration:true,emitDeclarationOnly:true,noEmitOnError:false,outDir:out,rootDir:path.join(root,'src'),strict:true,types:[]};
const program=ts.createProgram([path.join(root,'src/compiler/types.ts')],options);
const emitted=[];
const result=program.emit(undefined,(file,text)=>{fs.mkdirSync(path.dirname(file),{recursive:true});fs.writeFileSync(file,text);emitted.push(file);});
if(result.emitSkipped || result.diagnostics.length) throw Error('declaration emission failed');
const hash = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
const checker = program.getTypeChecker();
const source = program.getSourceFile(path.join(root,'src/compiler/types.ts'));
const symbols = checker.getExportsOfModule(checker.getSymbolAtLocation(source));
const fields = Object.fromEntries(['SourceFile','Node','Diagnostic','DiagnosticMessageChain','BindableStaticPropertyAssignmentExpression','BindableStaticAccessExpression','LiteralType','PseudoBigInt'].map(name => { const symbol=symbols.find(s=>s.name===name); if(!symbol) throw Error('missing upstream type '+name); return [name,checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(p=>p.name).sort()]; }));
const manifest={fields,upstream_commit:pin.commit,typescript_api:ts.version,source_types_sha256:hash(path.join(root,'src/compiler/types.ts')),declarations:Object.fromEntries(emitted.map(file=>[path.relative(out,file),hash(file)])),pairs:pairs.map(({type_id,type,field,declared_type,read_count,sites})=>({type_id,type,field,declared_type,read_count,sites}))};
fs.writeFileSync(path.join(out,'original-manifest.json'),JSON.stringify(manifest,null,2)+'\n');
console.log(`Original declarations emitted: ${emitted.length}; matched sites: ${pairs.map(p=>p.read_count).join('+')}; pin: ${pin.commit}`);
