'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
const {expectedApi} = require('./check-api.cjs');

// Compiles every upstream test case that reads the built API (/.ts/typescript.d.ts) against stock 6.0.3's typescript.d.ts
// and against the same file with this head's 19 sanctions applied, and requires the same diagnostics from both (README.md).
// usage: node api-samples.cjs <pinned TypeScript checkout> <stock typescript.d.ts>
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 4, 'usage: node api-samples.cjs <pinned TypeScript checkout> <stock typescript.d.ts>');
const [checkout, stockFile] = process.argv.slice(2);
assert.equal(execFileSync('git', ['-C', checkout, 'rev-parse', 'HEAD'], {encoding: 'utf8'}).trim(), '050880ce59e30b356b686bd3144efe24f875ebc8', 'not the pinned TypeScript commit');
const stockText = fs.readFileSync(stockFile, 'utf8');
const libraryDirectory = path.dirname(stockFile);
const apis = {stock: stockText, adapted: expectedApi(stockText).expected};
// The harness mounts the built API at /.ts. The internal bundles are built from the full tree and are not in the stock package.
const built = side => ({
    '/.ts/typescript.d.ts': apis[side],
    '/.ts/tsserverlibrary.d.ts': fs.readFileSync(path.join(libraryDirectory, 'tsserverlibrary.d.ts'), 'utf8'),
});
// Harness-only directives, which are not compiler options.
const harnessOnly = new Set(['filename', 'noTypesAndSymbols', 'includeBuiltFile', 'libFiles', 'noImplicitReferences', 'currentDirectory']);

// The harness's test format: `// @name: value` directives, and each @filename starts a unit.
function units(fileName) {
    const text = fs.readFileSync(fileName, 'utf8').replace(/^﻿/, '');
    const options = {};
    const files = [];
    let current;
    for (const line of text.split(/\r?\n/)) {
        const directive = /^\/\/\s*@(\w+)\s*:\s*(.*?)\s*$/.exec(line);
        if (directive && directive[1].toLowerCase() === 'filename') {
            current = {name: path.posix.join('/', directive[2]), lines: []};
            files.push(current);
        } else if (directive && !current) {
            options[directive[1]] = directive[2];
        } else {
            if (!current) {
                current = {name: '/' + path.basename(fileName), lines: []};
                files.push(current);
            }
            current.lines.push(line);
        }
    }
    return {options, files: files.map(f => ({name: f.name, text: f.lines.join('\n')}))};
}

function compilerOptions(directives) {
    const json = {};
    for (const [name, value] of Object.entries(directives)) {
        if (harnessOnly.has(name)) continue;
        json[name] = value === 'true' ? true : value === 'false' ? false : name === 'lib' ? value.split(',').map(v => v.trim()) : value;
    }
    const converted = ts.convertCompilerOptionsFromJson(json, '/');
    assert.equal(converted.errors.length, 0, ts.flattenDiagnosticMessageText(converted.errors[0]?.messageText ?? '', ' '));
    return {...converted.options, types: []};
}

function compile(test, side) {
    const virtual = new Map([...test.files.map(f => [f.name, f.text]), ...Object.entries(built(side))]);
    const options = compilerOptions(test.options);
    const host = ts.createCompilerHost(options);
    const library = name => path.join(libraryDirectory, path.basename(name));
    const inLibrary = name => name.startsWith(libraryDirectory) || /^\/?lib\.[\w.]+\.d\.ts$/.test(path.basename(name)) && fs.existsSync(library(name));
    host.getDefaultLibLocation = () => libraryDirectory;
    host.getCurrentDirectory = () => '/';
    host.fileExists = name => virtual.has(name) || inLibrary(name) && fs.existsSync(library(name));
    host.readFile = name => virtual.get(name) ?? (inLibrary(name) ? fs.readFileSync(library(name), 'utf8') : undefined);
    host.directoryExists = name => [...virtual.keys()].some(key => key.startsWith(name.endsWith('/') ? name : name + '/'));
    host.getDirectories = () => [];
    host.realpath = name => name;
    host.getSourceFile = (name, languageVersion) => {
        const text = host.readFile(name);
        return text === undefined ? undefined : ts.createSourceFile(name, text, languageVersion, true);
    };
    const roots = test.files.filter(f => /\.(ts|tsx)$/.test(f.name)).map(f => f.name);
    const program = ts.createProgram(roots, options, host);
    return ts.getPreEmitDiagnostics(program).map(d => `${d.file ? d.file.fileName + ':' + (d.file.getLineAndCharacterOfPosition(d.start).line + 1) : 'global'} TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ')}`);
}

const cases = execFileSync('git', ['-C', checkout, 'grep', '-l', '-F', '/.ts/typescript.d.ts', '--', 'tests/cases'], {encoding: 'utf8'}).trim().split('\n').sort();
const results = cases.map(relative => {
    const test = units(path.join(checkout, relative));
    const stock = compile(test, 'stock');
    const adapted = compile(test, 'adapted');
    return {case: relative, stock, adapted, same: JSON.stringify(stock) === JSON.stringify(adapted)};
});
console.log(JSON.stringify(results, null, 2));
for (const result of results) assert(result.same, `${result.case}: diagnostics differ with the adapted API`);
console.log(`${results.length} cases, the same diagnostics against both APIs`);
