'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const rules = require('./rules.json');
const [before, after, mutant] = process.argv.slice(2);
assert.equal(ts.version, '6.0.3');
const read = (root, file) => fs.readFileSync(path.join(root, file), 'utf8');
const parse = (root, file) => ts.createSourceFile(file, read(root, file), ts.ScriptTarget.Latest, true);
function nodes(node, test) { const found = []; (function visit(n) { if (test(n)) found.push(n); ts.forEachChild(n, visit); })(node); return found; }
function fn(root, file, name) {
    const source = parse(root, file), found = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === name);
    assert.equal(found.length, 1, name);
    return found[0].getText(source).replace(/^export /, '');
}
const js = text => ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS}}).outputText;
const commandFile = 'src/compiler/commandLineParser.ts';
let objectCode = fn(after, commandFile, 'isAdamicJsonObject') + '\n' + fn(after, commandFile, 'checkedAdamicJsonObject');
if (mutant === 'object') objectCode = objectCode.replace('typeof value === "object"', 'typeof value === "string"');
const object = Function(js(objectCode) + '; return {isAdamicJsonObject, checkedAdamicJsonObject};')();
for (const value of [{}, {files: []}, [], Object.assign(function () {}, {files: []})]) {
    assert.equal(object.checkedAdamicJsonObject(value), value, 'object check preserves identity');
}
for (const value of [undefined, null, 0, true, 'text']) assert.throws(() => object.checkedAdamicJsonObject(value), TypeError);

// Run the actual inherited-spec arrow, with TypeScript's real path/map helpers.
function transfer(root, ownRaw, extendsRaw) {
    const source = parse(root, commandFile);
    const declaration = nodes(source, n => ts.isVariableDeclaration(n) && n.name.getText(source) === 'setPropertyInResultIfNotUndefined')[0];
    assert(declaration);
    let code = declaration.initializer.getText(source);
    if (mutant === 'inherited' && root === after) code = code.replace('!isArray(extendedSpecs)', 'isArray(extendedSpecs)');
    return Function('ts', 'ownRaw', 'extendsRaw', `
        const {map, isArray, isString, isRootedDiskPath, combinePaths,
            convertToRelativePath, getDirectoryPath, createGetCanonicalFileName, startsWith} = ts;
        ${js(nodes(source, n => ts.isVariableDeclaration(n) && n.name.getText(source) === 'configDirTemplate')[0].parent.parent.getText(source) + '\n' + fn(root, commandFile, 'startsWithConfigDirTemplate'))}
        const ownConfig = {raw: ownRaw}, result = {}, host = {useCaseSensitiveFileNames: true};
        const basePath = '/p', extendedConfigPath = '/p/base.json'; let relativeDifference;
        ${js('const apply = ' + code + ';')}
        apply('files'); return result;
    `)(ts, ownRaw, extendsRaw);
}
for (const [own, extended] of [[{}, {files: ['/p/a.ts', '/p/b.ts']}], [{}, {files: []}], [{files: ['/own.ts']}, {files: ['/other.ts']}], [{}, {}]]) {
    assert.deepEqual(transfer(after, own, extended), transfer(before, own, extended));
}

// The two actual base-Request dispatch bodies, not a reimplementation of their checks.
let locationCode = fn(after, 'src/server/session.ts', 'isAdamicFileLocation') + '\n' + fn(after, 'src/server/session.ts', 'checkedAdamicFileLocation');
if (mutant === 'location') locationCode = locationCode.replace('typeof value.file === "string"', 'typeof value.file === "number"');
const isLocation = Function('isAdamicJsonObject', js(locationCode) + '; return checkedAdamicFileLocation;')(object.isAdamicJsonObject);
function implementation(root, flag) {
    const source = parse(root, 'src/server/session.ts');
    const arrows = nodes(source, n => ts.isArrowFunction(n) && n.body.getText(source).includes('this.getImplementation(') && n.body.getText(source).includes(`/*simplifiedResult*/ ${flag}`));
    assert.equal(arrows.length, 1);
    return Function('checkedAdamicFileLocation', js('const run = function(request) ' + arrows[0].body.getText(source)) + '; return run;')(isLocation);
}
for (const flag of [true, false]) for (const args of [{file: '/a.ts', line: 1, offset: 1}, {file: '/a.ts', line: 2, offset: 3, projectFileName: '/p/tsconfig.json', position: 4}]) {
    const receiver = {requiredResponse: value => value, getImplementation: (value, simplified) => ({value, simplified})};
    const original = implementation(before, flag).call(receiver, {arguments: args});
    const checked = implementation(after, flag).call(receiver, {arguments: args});
    assert.deepEqual(checked, original); assert.equal(checked.value, args);
}

const positionOnly = {file: '/a.ts', position: 3};
const receiverByPosition = {requiredResponse: value => value, getImplementation: value => value.position};
assert.equal(implementation(after, true).call(receiverByPosition, {arguments: positionOnly}),
    implementation(before, true).call(receiverByPosition, {arguments: positionOnly}));

// Produce real response values with the stock language service, then run the changed read fragments.
const file = '/sample.ts', text = 'let value = 1; value;';
const service = ts.createLanguageService({getCompilationSettings: () => ({noLib: true}), getScriptFileNames: () => [file], getScriptVersion: () => '0', getScriptSnapshot: name => name === file ? ts.ScriptSnapshot.fromString(text) : undefined, getCurrentDirectory: () => '/', getDefaultLibFileName: () => '/lib.d.ts', fileExists: name => name === file, readFile: name => name === file ? text : undefined});
const details = service.getCompletionEntryDetails(file, text.length - 1, 'value', undefined, undefined, undefined, undefined);
const references = service.findReferences(file, 4);
assert(details && references);
const clientSource = parse(after, 'src/harness/client.ts');
const declarations = nodes(clientSource, n => ts.isFunctionDeclaration(n) && /^adamic(ArrayOf|Is)/.test(n.name?.text || ''));
let validation = declarations.map(n => n.getText(clientSource)).join('\n');
if (mutant === 'full-reply-string') validation = validation.replace('typeof value.text === "string" || ', '');
if (mutant === 'full-reply-invalid') validation = validation.replace('(value.text === undefined || typeof value.text === "string" ||', '(value.text === undefined || true ||');
if (mutant === 'response') validation = validation.replace('Array.isArray(value) &&', '!Array.isArray(value) &&');
const checks = Function('ts', js(validation) + '; return {adamicArrayOf, adamicIsCompletionEntryDetails, adamicIsReferencedSymbol};')({...ts, isAdamicJsonObject: object.isAdamicJsonObject});
const debug = {assert(condition) { assert(condition); }};
for (const rule of rules.filter(r => r.file === 'src/harness/client.ts' && ['getCompletionEntryDetails', 'findReferences'].includes(r.owner) && r.before.includes('return'))) {
    const response = {body: rule.owner === 'findReferences' ? references : [details]};
    const run = code => Function('response', 'Debug', ...Object.keys(checks), code)(response, debug, ...Object.values(checks));
    assert.equal(run(rule.after), run(rule.before), 'checked response retains array/item identity');
}

// completionEntryDetails-full keeps service span edits and accepts protocol tag text.
const detailsRule = rules.find(r => r.file === 'src/harness/client.ts' && r.owner === 'getCompletionEntryDetails' && r.before.includes('Debug.assert'));
for (const tagText of [undefined, '', 'p1', [], [{text: 'p1', kind: 'text'}]]) {
    const tags = [{name: 'param', text: tagText}];
    const span = {start: 0, length: 0};
    const codeActions = [{description: 'insert', changes: [{fileName: file, textChanges: [{span, newText: 'x'}]}]}];
    const full = {...details, tags, codeActions};
    const response = {body: [full]};
    const run = code => Function('response', 'Debug', ...Object.keys(checks), code)(response, debug, ...Object.values(checks));
    assert.equal(run(detailsRule.before), full);
    const checked = run(detailsRule.after);
    assert.equal(checked, full);
    assert.equal(checked.tags, tags);
    assert.equal(checked.tags[0].text, tagText);
    assert.equal(checked.codeActions, codeActions);
    assert.equal(checked.codeActions[0].changes[0].textChanges[0].span, span);
}
for (const invalid of [null, 1, true, {}, [1], [{text: 'p1', kind: 1}]]) {
    const response = {body: [{...details, tags: [{name: 'param', text: invalid}]}]};
    assert.throws(() => Function('response', 'Debug', ...Object.keys(checks), detailsRule.after)(response, debug, ...Object.values(checks)), TypeError);
}

for (const rule of rules.filter(r => r.file === 'src/harness/harnessLanguageService.ts' && r.before.includes('info.config.message'))) {
    const run = code => {
        const parts = {displayParts: []};
        return Function('ts', 'info', 'parts', 'let customMessage = ""; ' + code + '; return {customMessage, parts};')(
            {isAdamicJsonObject: object.isAdamicJsonObject}, {config: {message: 'configured'}}, parts);
    };
    const checked = mutant === 'plugin' ? rule.after.replace('!== "string"', '!== "number"') : rule.after;
    assert.deepEqual(run(checked), run(rule.before));
}

let localized = fn(after, 'src/services/services.ts', 'isAdamicLocalizedMessages');
if (mutant === 'localized') localized = localized.replace('!== "string"', '!== "number"');
const isMessages = Function('isAdamicJsonObject', js(localized) + '; return isAdamicLocalizedMessages;')(object.isAdamicJsonObject);
const localeRule = rules.find(r => r.owner === 'createLanguageService');
function locale(code) {
    const messages = {key: 'initial'}; let calls = 0, callback;
    const host = {getLocalizedDiagnosticMessages() { assert.equal(this, host); calls++; return messages; }};
    Function('host', 'maybeSetLocalizedDiagnosticMessages', 'isAdamicLocalizedMessages', js(code))(host, value => {callback = value;}, isMessages);
    assert.equal(calls, 0, 'hook stays lazy');
    const result = callback(); assert.equal(calls, 1); assert.equal(result, messages, 'retain dictionary identity');
    messages.key = 'updated'; assert.equal(result.key, 'updated');
    return result;
}
assert.deepEqual(locale(localeRule.after), locale(localeRule.before));

const echoRule = rules.find(r => r.file.includes('src/testRunner/'));
function echo(code, value) {
    let handler;
    Function('ts', js('function register() {' + code + '}') + '; return register;')({}).call({addProtocolHandler(name, callback) {assert.equal(name, 'echo'); handler = callback;}});
    return handler({arguments: value});
}
for (const value of [undefined, {}, [], 0, '', false, 'echo']) {
    const checked = mutant === 'echo' ? echoRule.after.replace('req.arguments === null', 'req.arguments !== null') : echoRule.after;
    assert.deepEqual(echo(checked, value), echo(echoRule.before, value));
}
console.log(JSON.stringify({classes: ['object', 'inherited', 'location', 'response', 'plugin', 'localized', 'echo', 'full-reply-string', 'full-reply-invalid'], beforeAfter: 'pass', mutant: mutant || null}));
