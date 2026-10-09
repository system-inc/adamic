'use strict';
// Stock TypeScript decides owner/caller validity. Mutants change actual source types.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto'), ts = require('typescript');
const {plan} = require('./adapt.cjs');
if (ts.version !== '6.0.3' || process.argv.length !== 6) throw Error('usage: NODE_PATH=<6.0.3 modules> node prove.cjs BEFORE_TREE AFTER_TREE DRIVER OUTPUT');
const [beforeTree, afterTree, driverFile, output] = process.argv.slice(2).map(p => path.resolve(p));
fs.mkdirSync(output, {recursive: true});
const relative = 'src/compiler/scanner.ts';
const before = fs.readFileSync(path.join(beforeTree,relative),'utf8'), after = fs.readFileSync(path.join(afterTree,relative),'utf8');
const driver = fs.readFileSync(driverFile,'utf8');
function assert(value,message) { if (!value) throw Error(message); }
assert(plan(before).text === after, 'after is not the exact reviewed transformation of main');
assert(plan(after).text === after, 'adapter is not idempotent');
function emit(text) { return ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true}}).outputText; }
assert(emit(before) === emit(after),'scanner JavaScript changed');
const oldDriver = driver.replace('arg0: string | number | undefined','arg0');
assert(oldDriver !== driver && emit(oldDriver) === emit(driver),'driver annotation changed JavaScript');
const scannerFile = path.join(afterTree,relative), configFile = path.join(afterTree,'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configFile,ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config,ts.sys,path.dirname(configFile));
const options = {...parsed.options,noEmit:true,allowImportingTsExtensions:true,typeRoots:[path.join(afterTree,"node_modules/@types")]};
const virtualDriver = path.join(afterTree,'src/compiler/scanner-driver-proof.ts');
function checkerDriver(text) {
    return text.replace("import type {} from 'node:util';", '')
        .replace("'./adapted/src/compiler/scanner.ts'", "'./scanner.ts'")
        .replace("'./adapted/src/compiler/types.ts'", "'./types.ts'")
        .replace("import { tokenNames } from './token-names.a';", 'declare const tokenNames: {readonly [index: number]: string | undefined};')
        .replace("import { panic, programArguments, readTextFile } from 'adamic';", 'declare function panic(message: string): never; declare function programArguments(): string[]; declare function readTextFile(path: string): {kind: "Error"; message: string} | {kind: "Ok"; text: string};');
}
function program(scannerText, driverText) {
    const host = ts.createCompilerHost(options), originalRead = host.readFile.bind(host), originalExists = host.fileExists.bind(host);
    host.readFile = file => path.resolve(file) === scannerFile ? scannerText : path.resolve(file) === virtualDriver ? checkerDriver(driverText) : originalRead(file);
    host.fileExists = file => path.resolve(file) === virtualDriver || originalExists(file);
    return ts.createProgram([...parsed.fileNames,virtualDriver],options,host);
}
function observations(p) {
    return ts.getPreEmitDiagnostics(p).map(d => ({code:d.code,file:d.file && path.relative(afterTree,d.file.fileName),start:d.start,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
}
const p = program(after,driver), checker = p.getTypeChecker(), sf = p.getSourceFile(scannerFile);
const baseline = observations(p);
fs.writeFileSync(path.join(output,'stock-diagnostics.json'),JSON.stringify(baseline,null,2)+'\n');
assert(baseline.length === 0,'stock checker rejects the adapted compiler or driver');
const payloads = [], map = [], publicTypes = [];
function visit(n) {
    if (ts.isTypeAliasDeclaration(n) && n.name.text === 'ErrorCallback') publicTypes.push(n.getText(sf));
    if (ts.isVariableDeclaration(n) && n.name.getText(sf) === 'textToKeyword') map.push({type:checker.typeToString(checker.getTypeAtLocation(n.name)),initializer:n.initializer.getText(sf)});
    if (ts.isCallExpression(n) && n.expression.getText(sf) === 'error' && n.arguments.length === 4) {
        const arg = n.arguments[3], type = checker.getTypeAtLocation(arg);
        assert(!(type.flags & (ts.TypeFlags.Any|ts.TypeFlags.Unknown)),'unproved error caller');
        payloads.push({line:sf.getLineAndCharacterOfPosition(arg.getStart(sf)).line+1,expression:arg.getText(sf),type:checker.typeToString(type)});
    }
    ts.forEachChild(n,visit);
}
visit(sf);
assert(map.length === 1 && map[0].type === 'Map<string, KeywordSyntaxKind>','keyword map type changed');
assert(payloads.some(r => r.expression === 'numberOfCapturingGroups' && r.type === 'number'),'numeric payload witness missing');
const beforeSF=ts.createSourceFile('before.ts',before,ts.ScriptTarget.Latest,true);
const beforePublic = beforeSF.statements.filter(n => ts.isTypeAliasDeclaration(n) && n.name.text === 'ErrorCallback').map(n=>n.getText(beforeSF));
assert(JSON.stringify(publicTypes) === JSON.stringify(beforePublic),'public callback declaration changed');
const mutants = [
    {site:'scanner-error',scanner:after.replace(/(function error\(message: DiagnosticMessage[^\n]*arg0\?: )string \| number/g,'$1string'),driver},
    {site:'keyword-map',scanner:after.replace('new Map<string, KeywordSyntaxKind>','new Map<string, string>'),driver},
    {site:'driver-error',scanner:after,driver:driver.replace('arg0: string | number | undefined','arg0: string | undefined')},
];
const results = [];
for (const m of mutants) {
    const diagnostics = observations(program(m.scanner,m.driver));
    fs.writeFileSync(path.join(output,m.site+'-mutant.json'),JSON.stringify(diagnostics,null,2)+'\n');
    assert(diagnostics.some(d => [2322,2345,2769].includes(d.code)),m.site+' mutant escaped stock checking');
    results.push({site:m.site,caughtBy:'stock TypeScript source/consumer checking',diagnosticCodes:[...new Set(diagnostics.map(d=>d.code))]});
}
// Drift and duplicate owners must fail the adapter, rather than applying a guess.
for (const [name,text] of [ ['drift',before.replace('Object.entries(textToKeywordObj)','Object.values(textToKeywordObj)')], ['duplicate',before+'\nconst textToKeyword = new Map(Object.entries(textToKeywordObj));\n'] ]) {
    let caught=false;try {plan(text);} catch {caught=true;} assert(caught,name+' adapter mutant escaped');
}
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
const report={stockDiagnostics:baseline.length,scannerJavaScriptIdentical:true,driverJavaScriptIdentical:true,publicCallbackIdentical:true,idempotent:true,mainInputHash:hash(before),afterHash:hash(after),map,payloads,mutants:results,guardMutants:['drift','duplicate']};
fs.writeFileSync(path.join(output,'proof.json'),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report));
