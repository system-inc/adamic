'use strict';
const path = require('node:path'), ts = require(path.resolve(process.argv[2]));
const files = new Map([
    ['/95g/exported.js', "/** @type {string} */\nlet key = 'xyz';\nclass Computed { static [key] = 1; }\nmodule.exports = /** @type {typeof Computed} */ (class {});\nmodule.exports.extra = 2;\n"],
    ['/95g/consumer.ts', "import v = require('./exported');\nv.xyz;\n"],
]);
const result = [];
for (const noImplicitAny of [false, true]) {
    const options={strict:true,noImplicitAny,allowJs:true,checkJs:true,noEmit:true,target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.CommonJS,moduleResolution:ts.ModuleResolutionKind.Node10,ignoreDeprecations:'6.0'};
    const host=ts.createCompilerHost(options),getSourceFile=host.getSourceFile,read=host.readFile,exists=host.fileExists;
    host.getSourceFile=(name,...args)=>files.has(name)?ts.createSourceFile(name,files.get(name),ts.ScriptTarget.Latest,true):getSourceFile.call(host,name,...args);
    host.readFile=name=>files.has(name)?files.get(name):read.call(host,name);host.fileExists=name=>files.has(name)||exists.call(host,name);host.directoryExists=name=>name==='/95g'||ts.sys.directoryExists(name);
    const program=ts.createProgram(['/95g/consumer.ts'],options,host),checker=program.getTypeChecker();
    const diagnostics=ts.getPreEmitDiagnostics(program);const row={noImplicitAny,diagnosticOutput:ts.formatDiagnostics(diagnostics,{getCanonicalFileName:f=>f,getCurrentDirectory:()=> '/',getNewLine:()=> '\n'}),diagnostics:diagnostics.map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}))};
    try { const symbol=checker.getSymbolAtLocation(program.getSourceFile('/95g/consumer.ts').statements[1].expression);row.result=symbol?{name:symbol.name,parent:symbol.parent?.name}:'undefined'; }
    catch(error){row.error=error.name+': '+error.message;row.stack=error.stack;}
    result.push(row);
}
console.log(JSON.stringify(result,null,2));
if (process.argv.includes('--expect-fixed')) {
    const assert = require('node:assert/strict');
    assert.equal(result.length, 2);
    for (const row of result) { assert.equal(row.result, 'undefined', row.error); assert.deepEqual(row.diagnostics, []); }
}
