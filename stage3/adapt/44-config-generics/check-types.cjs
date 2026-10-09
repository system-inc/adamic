'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict'), ts = require('typescript');
const tree = path.resolve(process.argv[2]), output = path.resolve(process.argv[3]);
const file = path.join(tree, 'src/compiler/commandLineParser.ts');
const configFile = path.join(tree, 'src/compiler/tsconfig.json');
const config = ts.parseJsonConfigFileContent(ts.readConfigFile(configFile, ts.sys.readFile).config, ts.sys, path.dirname(configFile), undefined, configFile);
assert.equal(config.errors.length, 0);
let source = fs.readFileSync(file, 'utf8');
for (const r of require('./rules.json')) {
    const nl = source.includes('\r\n') ? '\r\n' : '\n';
    source = source.replace(r.before.replace(/\r?\n/g, nl), r.after.replace(/\r?\n/g, nl));
}
const probe = `
function adamicConfigGenericProbe<TRaw, TValues extends readonly unknown[]>(raw: TRaw, defaults: CompilerOptions | TypeAcquisition, values: TValues, errors: Diagnostic[]) {
    const result: CompilerOptions | TypeAcquisition | undefined = convertOptionsFromJson(getCommandLineCompilerOptionsMap(), raw, "/", defaults, compilerOptionsDidYouMeanDiagnostics, errors);
    const watch: WatchOptions | undefined = convertOptionsFromJson(getCommandLineWatchOptionsMap(), raw, "/", undefined, watchOptionsDidYouMeanDiagnostics, errors);
    const compiler: CompilerOptions = convertCompilerOptionsFromJsonWorker(raw, "/", errors);
    const acquisition: TypeAcquisition = convertTypeAcquisitionFromJsonWorker(raw, "/", errors);
    const watchWorker: WatchOptions | undefined = convertWatchOptionsFromJsonWorker(raw, "/", errors);
    const save: boolean = convertCompileOnSaveOptionFromJson({compileOnSave: raw}, "/", errors);
    const normalized: CompilerOptionsValue = convertJsonOption(compileOnSaveCommandLineOption, raw, "/", errors);
    convertJsonOptionOfListType(extendsOptionDeclaration, values, "/", errors, undefined, undefined, undefined);
}
`;
function check(text) {
    const host = ts.createCompilerHost(config.options), read = host.readFile;
    host.readFile = f => f === file ? text : read(f);
    return ts.getPreEmitDiagnostics(ts.createProgram(config.fileNames, config.options, host)).map(d => ({file: d.file ? path.relative(tree, d.file.fileName) : null, code: d.code, start: d.start, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}));
}
const control = check(source + probe);
assert.equal(control.length, 0, JSON.stringify(control));
const parsed = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true);
const worker = parsed.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'convertCompilerOptionsFromJsonWorker');
assert(worker?.type);
const wrongReturn = source.slice(0, worker.type.getStart(parsed)) + 'string' + source.slice(worker.type.end);
const checkerMutant = check(wrongReturn + probe);
assert(checkerMutant.some(d => d.code === 2322), 'wrong actual worker result must fail stock checking');
// Normalization mutates values: it must never promise the caller's narrower literal subtype.
const literalProbe = `
function adamicConfigLiteralProbe(errors: Diagnostic[]) { const preserved: { strict: true } = convertOptionsFromJson(getCommandLineCompilerOptionsMap(), {strict: false}, "/", {strict: true}, compilerOptionsDidYouMeanDiagnostics, errors)!; }
`;
const subtypeWitness = check(source + literalProbe);
assert(subtypeWitness.some(d => d.code === 2322), 'mutated default cannot retain caller literal subtype');
const declined = [];
for (const name of ['parseOwnConfigOfJson']) {
    const original = 'function ' + name + '(\r\n    json: any,';
    assert(source.includes(original), 'declined site exists: ' + name);
    const diagnostics = check(source.replace(original, 'function ' + name + '<TRaw>(\r\n    json: TRaw,'));
    assert(diagnostics.length > 0, 'unchecked generic root must fail: ' + name);
    declined.push({site: name, diagnostics});
}
fs.mkdirSync(path.dirname(output), {recursive: true});
fs.writeFileSync(output, JSON.stringify({typescript: ts.version, consumerDiagnostics: control.length, genericCallerProbe: 'all converter callers accept arbitrary raw types; normalized default union and undefined preserved; narrower subtypes are not promised', checkerMutant, subtypeWitness, declined}, null, 2) + '\n');
console.log(JSON.stringify({consumerDiagnostics: control.length, checkerMutant: checkerMutant.length, subtypeWitness: subtypeWitness.length, declined: declined.map(d => ({site: d.site, diagnostics: d.diagnostics.length}))}));
