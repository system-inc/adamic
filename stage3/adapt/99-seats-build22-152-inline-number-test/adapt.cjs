'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const sites = [
    {
        argument: 'oldOptionsOrEmitKind',
        declaration: 'oldOptionsOrEmitKind: CompilerOptions | BuilderFileEmit | undefined',
        before: 'const oldEmitKind = oldOptionsOrEmitKind && (isNumber(oldOptionsOrEmitKind) ? oldOptionsOrEmitKind : getBuilderFileEmit(oldOptionsOrEmitKind));',
    },
    {
        argument: 'optionsOrEmitKind',
        declaration: 'optionsOrEmitKind: CompilerOptions | BuilderFileEmit',
        before: 'const emitKind = isNumber(optionsOrEmitKind) ? optionsOrEmitKind : getBuilderFileEmit(optionsOrEmitKind);',
    },
];

function compilerProgram(root) {
    const configPath = path.resolve(root, 'src/compiler/tsconfig.json');
    const config = ts.getParsedCommandLineOfConfigFile(configPath, {}, {
        ...ts.sys,
        onUnRecoverableConfigFileDiagnostic(diagnostic) {
            throw Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'));
        },
    });
    if (!config || config.errors.length) throw Error('Cannot read compiler tsconfig.json');
    return ts.createProgram({
        rootNames: config.fileNames,
        options: config.options,
        projectReferences: config.projectReferences,
    });
}

function read(program, root, file) {
    const name = path.resolve(root, 'src/compiler', file);
    const source = program.getSourceFile(name);
    if (!source || source.parseDiagnostics.length) throw Error('Cannot parse ' + file);
    return {name, text: source.text, source};
}

function requirePredicateBinding(checker, symbol, predicate) {
    while (symbol && symbol.flags & ts.SymbolFlags.Alias) {
        symbol = checker.getAliasedSymbol(symbol);
    }
    if (symbol?.declarations?.length !== 1 || symbol.declarations[0] !== predicate) {
        throw Error('isNumber does not resolve to the pinned core.ts predicate');
    }
}

function owner(source, name) {
    const matches = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
    if (matches.length !== 1 || !matches[0].body) throw Error('Expected one definition of ' + name);
    return matches[0];
}

function main(root) {
    if (!root || ts.version !== '6.0.3') throw Error('Expected a tree and TypeScript 6.0.3');
    const program = compilerProgram(root);
    const core = read(program, root, 'core.ts');
    const predicate = owner(core.source, 'isNumber');
    if (predicate.getText(core.source).replaceAll('\r\n', '\n') !==
        'export function isNumber(x: unknown): x is number {\n    return typeof x === "number";\n}') {
        throw Error('Changed isNumber definition');
    }

    const builder = read(program, root, 'builder.ts');
    const checker = program.getTypeChecker();
    const pending = owner(builder.source, 'getPendingEmitKind');
    if (pending.parameters.length !== sites.length ||
        [...pending.parameters].reverse().some((parameter, index) =>
            !ts.isIdentifier(parameter.name) || parameter.getText(builder.source) !== sites[index].declaration)) {
        throw Error('Changed getPendingEmitKind parameters');
    }

    const edits = [];
    const calls = [];
    function visit(node) {
        if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'isNumber') {
            calls.push(node);
        }
        ts.forEachChild(node, visit);
    }
    visit(pending.body);

    for (const [index, site] of sites.entries()) {
        const statement = pending.body.statements[index];
        const call = 'isNumber(' + site.argument + ')';
        const replacement = 'typeof ' + site.argument + ' === "number"';
        const snapshot = statement?.getText(builder.source);
        // Replay still needs the same binding in scope at the original call site.
        if (snapshot === site.before.replace(call, replacement)) {
            requirePredicateBinding(checker,
                checker.resolveName('isNumber', statement, ts.SymbolFlags.Value, false), predicate);
            continue;
        }
        const node = calls.find(candidate => candidate.getText(builder.source) === call &&
            candidate.getStart(builder.source) >= statement?.getStart(builder.source) && candidate.end <= statement.end);
        if (snapshot !== site.before || !node || node.arguments.length !== 1 ||
            !ts.isIdentifier(node.arguments[0]) || node.arguments[0].text !== site.argument) {
            throw Error('Changed call snapshot for ' + site.argument);
        }
        requirePredicateBinding(checker, checker.getSymbolAtLocation(node.expression), predicate);
        edits.push({start: node.getStart(builder.source), end: node.end, replacement});
    }
    if (calls.length !== edits.length || (edits.length && edits.length !== sites.length)) {
        throw Error('Unexpected or half-adapted isNumber calls in getPendingEmitKind');
    }

    let next = builder.text;
    for (const {start, end, replacement} of edits.reverse()) {
        next = next.slice(0, start) + replacement + next.slice(end);
    }
    if (edits.length) fs.writeFileSync(builder.name, next);
    console.log(JSON.stringify({files: edits.length ? 1 : 0, edits: edits.length, javascriptChanging: true}));
}

main(process.argv[2]);
