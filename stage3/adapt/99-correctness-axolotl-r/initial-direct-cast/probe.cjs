'use strict';
// Read-only stock-checker probe. The first mint trial exists only in the host's memory.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const assert = require('node:assert/strict'), ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const tree = path.resolve(process.argv[2]);
const dir = path.join(tree, 'src/compiler'), file = path.join(dir, 'core.ts');
const raw = fs.readFileSync(file, 'utf8');
const hash = x => crypto.createHash('sha256').update(x).digest('hex');
const options = {strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true,
    erasableSyntaxOnly: false, verbatimModuleSyntax: true, allowImportingTsExtensions: true,
    noEmit: true, module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: []};
const roots = fs.readdirSync(dir, {recursive: true}).filter(n => n.endsWith('.ts')).sort().map(n => path.join(dir, n));
assert.equal(roots.length, 79);
const diagnostic = d => ({code: d.code, file: d.file && path.relative(tree, d.file.fileName),
    ...(d.file && d.start !== undefined ? (() => {const p=d.file.getLineAndCharacterOfPosition(d.start); return {line:p.line+1,column:p.character+1};})() : {}),
    message: ts.flattenDiagnosticMessageText(d.messageText, '\n')});
function check(text, rootNames = roots, targetFile = file) {
    const host = ts.createCompilerHost(options), read = host.readFile;
    host.readFile = name => path.resolve(name) === targetFile ? text : read(name);
    return ts.getPreEmitDiagnostics(ts.createProgram(rootNames, options, host)).map(diagnostic);
}
const baseline = check(raw);

const from = 'emptyArray as any as SortedReadonlyArray<T>', to = 'emptyArray as SortedReadonlyArray<T>';
assert.equal(raw.split(from).length, 2);
const trial = raw.replace(from, to), diagnostics = check(trial);
const baselineKeys = new Set(baseline.map(d => JSON.stringify(d)));
const addedDiagnostics = diagnostics.filter(d => !baselineKeys.has(JSON.stringify(d)));
assert.equal(addedDiagnostics.length, 1);
assert.equal(addedDiagnostics[0].code, 2352);
assert.equal(addedDiagnostics[0].line, 736);
assert(diagnostics.length === baseline.length + 1);
assert.equal(fs.readFileSync(file, 'utf8'), raw);
const publicText = fs.readFileSync(path.join(dir, 'corePublic.ts'), 'utf8');
const source = ts.createSourceFile('corePublic.ts', publicText, ts.ScriptTarget.Latest, true);
const brands = source.statements.filter(n => ts.isInterfaceDeclaration(n) && ['SortedArray','SortedReadonlyArray'].includes(n.name.text)).map(n => {
    assert.equal(n.members.length, 1); assert.equal(n.members[0].name.text, ' __sortedArrayBrand');
    assert.equal(n.members[0].type.kind, ts.SyntaxKind.UndefinedKeyword);
    return n.getText(source);
});
assert.equal(brands.length, 2);
const coreSource = ts.createSourceFile('core.ts', raw, ts.ScriptTarget.Latest, true);
const emptyDeclaration = coreSource.statements.find(n => ts.isVariableStatement(n) && n.declarationList.declarations.some(d => d.name.getText(coreSource) === 'emptyArray')).getText(coreSource);
const witnessFile = path.join(tree, 'sorted-brand-admission.ts');
const witnessBefore = brands.join('\n') + '\n' + emptyDeclaration + '\nfunction mint<T>() { return ' + from + '; }\n';
const witnessAfter = witnessBefore.replace(from, to);
const witnessBaselineDiagnostics = check(witnessBefore, [witnessFile], witnessFile);
const witnessTrialDiagnostics = check(witnessAfter, [witnessFile], witnessFile);
assert.equal(witnessBaselineDiagnostics.length, 0);
assert.equal(witnessTrialDiagnostics.length, 1);
assert.equal(witnessTrialDiagnostics[0].code, 2352);
const casts=[];
function walk(directory) {for(const entry of fs.readdirSync(directory, {withFileTypes:true})) {
    const name=path.join(directory,entry.name); if(entry.isDirectory()) walk(name);
    else if(name.endsWith('.ts')) {
        const text=fs.readFileSync(name,'utf8'), s=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);
        function visit(n) {if((ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)) && ts.isTypeReferenceNode(n.type) && ['SortedArray','SortedReadonlyArray'].includes(n.type.typeName.getText(s))) {
            const p=s.getLineAndCharacterOfPosition(n.getStart(s)); casts.push({file:path.relative(tree,name),line:p.line+1,column:p.character+1,text:n.getText(s)});
        } ts.forEachChild(n,visit); } visit(s);
    }
}}
walk(path.join(tree,'src'));
console.log(JSON.stringify({typescript:ts.version,roots:roots.length,options,baselineDiagnostics:baseline,
    trial:{from,to,diagnostics,addedDiagnostics},stockAmbientContext: "Full stock run omits Adamic host overlays; its ambient diagnostics are retained and unchanged by this trial.",witness:{before:witnessBefore,after:witnessAfter,baselineDiagnostics:witnessBaselineDiagnostics,trialDiagnostics:witnessTrialDiagnostics},sourceUnchanged:true,coreSha256:hash(raw),corePublicSha256:hash(publicText),brands,casts,
    decision:'STOP: stock TypeScript rejects the first direct mint. No source adaptation was written.'},null,2));
