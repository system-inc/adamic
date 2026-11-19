const fs=require('node:fs'),path=require('node:path');
const ts=require('/tmp/adamic-g02/main/stage3/api/node_modules/typescript');
if(ts.version!=='6.0.3')throw Error(ts.version);
const root=process.argv[2],output=process.argv[3];
const config=path.join(root,'src/compiler/tsconfig.json');const raw=ts.readConfigFile(config,ts.sys.readFile);if(raw.error)throw Error('config');
const parsed=ts.parseJsonConfigFileContent(raw.config,ts.sys,path.dirname(config),{},config);
if(parsed.errors.length)throw Error(JSON.stringify(parsed.errors));
const baseOptions=process.env.FAMILY5_STOCK_STRICT==='1'?{strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:['node']}:{...parsed.options,noEmit:true};
const options={...baseOptions,typeRoots:['/tmp/adamic-g02/main/internal/load/node_types/node_modules/@types']};
const program=ts.createProgram(parsed.fileNames,options);
const ds=ts.getPreEmitDiagnostics(program).map(d=>{const lc=d.file&&d.file.getLineAndCharacterOfPosition(d.start);return {code:d.code,file:d.file?.fileName.replace(root+'/',''),line:lc?.line+1,column:lc?.character+1,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};});
const report={version:ts.version,roots:parsed.fileNames.map(p=>path.relative(root,p)),root_count:parsed.fileNames.length,options,diagnostic_count:ds.length,diagnostics:ds};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');console.log('stock roots='+report.root_count+' diagnostics='+ds.length);
const crypto = require('node:crypto');
if (ds.length || parsed.fileNames.length !== 79) process.exit(1);
const emitted = {};
const emitProgram = ts.createProgram(parsed.fileNames, {
    ...options, noEmit: false, composite: false, incremental: false,
    declaration: false, declarationMap: false, sourceMap: false, inlineSourceMap: false,
    rewriteRelativeImportExtensions: true,
    outDir: path.join(root, 'stock-js'), rootDir: path.join(root, 'src'),
});
const result = emitProgram.emit(undefined, (name, data, bom) => {
    if (!name.endsWith('.js')) return;
    emitted[path.relative(path.join(root, 'stock-js'), name)] =
        crypto.createHash('sha256').update(bom ? '\ufeff' + data : data).digest('hex');
    const destination = path.join(output + '.js-tree', path.relative(path.join(root, 'stock-js'), name));
    fs.mkdirSync(path.dirname(destination), {recursive: true});
    fs.writeFileSync(destination, bom ? '\ufeff' + data : data);
});
const ordered = Object.fromEntries(Object.entries(emitted).sort(([left], [right]) => left.localeCompare(right)));
fs.writeFileSync(output + '.emit.json', JSON.stringify(ordered, null, 2) + '\n');
console.log('emit diagnostics=' + result.diagnostics.length + ' JavaScript files=' + Object.keys(emitted).length);
if (result.diagnostics.length) console.error(ts.formatDiagnosticsWithColorAndContext(result.diagnostics, ts.sys));
if (result.emitSkipped || result.diagnostics.length || !Object.keys(emitted).length) process.exit(1);
