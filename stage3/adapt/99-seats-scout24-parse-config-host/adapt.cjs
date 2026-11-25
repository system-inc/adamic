'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const relative = 'src/compiler/watch.ts';
const ownerName = 'parseConfigFileWithSystem';
const newline = '\r\n';
const signature = 'export function parseConfigFileWithSystem(configFileName: string, optionsToExtend: CompilerOptions, extendedConfigCache: Map<string, ExtendedConfigCacheEntry> | undefined, watchOptionsToExtend: WatchOptions | undefined, system: System, reportDiagnostic: DiagnosticReporter): ParsedCommandLine | undefined {';
const before = [
    signature,
    '    const host: ParseConfigFileHost = system as any;',
    '    host.onUnRecoverableConfigFileDiagnostic = diagnostic => reportUnrecoverableDiagnostic(system, reportDiagnostic, diagnostic);',
    '    const result = getParsedCommandLineOfConfigFile(configFileName, optionsToExtend, host, extendedConfigCache, watchOptionsToExtend);',
    '    host.onUnRecoverableConfigFileDiagnostic = undefined!; // TODO: GH#18217',
    '    return result;',
    '}',
].join(newline);
const after = [
    signature,
    '    const host: ParseConfigFileHost = {',
    '        readFile: fileName => system.readFile(fileName),',
    '        fileExists: fileName => system.fileExists(fileName),',
    '        getCurrentDirectory: () => system.getCurrentDirectory(),',
    '        readDirectory: (root, extensions, excludes, includes, depth) => system.readDirectory(root, extensions, excludes, includes, depth),',
    '        useCaseSensitiveFileNames: system.useCaseSensitiveFileNames,',
    '        directoryExists: system.directoryExists?.bind(system),',
    '        realpath: system.realpath?.bind(system),',
    '        onUnRecoverableConfigFileDiagnostic: diagnostic => reportUnrecoverableDiagnostic(system, reportDiagnostic, diagnostic),',
    '    };',
    '    return getParsedCommandLineOfConfigFile(configFileName, optionsToExtend, host, extendedConfigCache, watchOptionsToExtend);',
    '}',
].join(newline);

function plan(text) {
    assert.equal(ts.version, '6.0.3');
    const source = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'watch.ts parse error');
    const owners = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === ownerName);
    assert.equal(owners.length, 1, 'Expected one ' + ownerName);
    const owner = owners[0];
    const actual = owner.getText(source);
    assert(actual === before || actual === after, 'Unreviewed ' + ownerName + ' snapshot');

    const imports = source.statements.filter(node => ts.isImportDeclaration(node) &&
        ts.isStringLiteral(node.moduleSpecifier) && node.moduleSpecifier.text === './_namespaces/ts.js');
    assert.equal(imports.length, 1, 'Expected one compiler namespace import');
    assert.equal(imports[0].moduleSpecifier.getText(source), '"./_namespaces/ts.js"',
        'Unreviewed compiler namespace import spelling');
    const bindings = imports[0].importClause?.namedBindings;
    assert(bindings && ts.isNamedImports(bindings), 'Expected named compiler imports');
    assert(!bindings.elements.some(element => element.name.text === 'maybeBind'),
        'Unexpected maybeBind import');
    if (actual === after) return text;
    return text.slice(0, owner.getStart(source)) + after + text.slice(owner.end);
}

module.exports = { relative, before, after, plan };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    const file = path.join(process.argv[2], relative);
    const text = fs.readFileSync(file, 'utf8');
    const edited = plan(text);
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'Source changed during adaptation');
    if (edited !== text) fs.writeFileSync(file, edited);
    console.log(ownerName + ': ' + (edited === text ? '0 edits' : '1 body edit'));
}
