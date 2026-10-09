'use strict';
// Stock checker witnesses for the consumer work that an honest parse contract exposes.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const [treeArg, outputArg] = process.argv.slice(2);
const tree = path.resolve(treeArg), output = path.resolve(outputArg);
const configFile = path.join(tree, 'src/compiler/tsconfig.json');
const config = ts.parseJsonConfigFileContent(ts.readConfigFile(configFile, ts.sys.readFile).config, ts.sys, path.dirname(configFile), undefined, configFile);
assert.equal(config.errors.length, 0);
function diagnostics(change) {
    const host = ts.createCompilerHost(config.options), read = host.readFile;
    host.readFile = file => {
        let text = read(file);
        if (change && file === path.join(tree, 'src/compiler/utilities.ts')) text = text.replace('export function tryParseJson(text: string): any', 'type AdamicParsedJson = string | number | boolean | null | AdamicParsedJson[] | { [key: string]: AdamicParsedJson };\nexport function tryParseJson(text: string): AdamicParsedJson | undefined');
        return text;
    };
    return ts.getPreEmitDiagnostics(ts.createProgram(config.fileNames, config.options, host)).map(d => ({file: d.file ? path.relative(tree, d.file.fileName) : null, code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}));
}
const baseline = diagnostics(false), candidate = diagnostics(true);
const identities = new Set(baseline.map(d => JSON.stringify(d)));
const added = candidate.filter(d => !identities.has(JSON.stringify(d)));
assert(added.some(d => d.file === 'src/compiler/utilities.ts' && d.code === 2322), 'actual JSON primitives contradict readJsonOrUndefined object contract');
assert(added.some(d => d.file === 'src/compiler/moduleSpecifiers.ts'), 'actual package JSON callers need narrowing');
const api = fs.readFileSync(path.join(tree, 'built/local/typescript.d.ts'), 'utf8');
assert.match(api, /function convertToObject\(sourceFile: JsonSourceFile, errors: Diagnostic\[\]\): any;/);
assert.equal(JSON.parse('7'), 7); assert.equal(JSON.parse('null'), null);
fs.mkdirSync(output, {recursive: true});
fs.writeFileSync(path.join(output, 'declines.json'), JSON.stringify({baselineDiagnostics: baseline.length, candidateNewDiagnostics: added, publicDeclaration: 'convertToObject(sourceFile: JsonSourceFile, errors: Diagnostic[]): any', tryParseJsonTruthfulType: 'AdamicParsedJson | undefined', witnesses: ['JSON.parse("7") === 7', 'JSON.parse("null") === null']}, null, 2) + '\n');
console.log(JSON.stringify({baselineDiagnostics: baseline.length, candidateNewDiagnostics: added.length, publicDeclarationUnchanged: true}));
