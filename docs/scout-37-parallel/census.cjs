// Syntactic call inventory, not a resolved call graph. Declarations are separate.
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
const ts = require(path.resolve(process.argv[2]));
const root = path.resolve(process.argv[3]);
const pin = '050880ce59e30b356b686bd3144efe24f875ebc8';
if (ts.version !== '6.0.3' || cp.execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim() !== pin) {
    throw new Error('TypeScript reader/corpus pin differs');
}
const names = ['createSourceFile', 'parseSourceFile', 'parseList', 'internIdentifier', 'createTypeChecker', 'getDiagnosticsHelper', 'checkSourceFileWithEagerDiagnostics', 'sortAndDeduplicateDiagnostics'];
const patterns = Object.fromEntries(names.map(name => [name, { definitions: [], calls: [] }]));
const sources = {};
const parserState = [];
const checkerState = [];
function files(directory) {
    return fs.readdirSync(directory, { withFileTypes: true }).flatMap(entry => {
        const file = path.join(directory, entry.name);
        return entry.isDirectory() ? files(file) : file.endsWith('.ts') ? [file] : [];
    }).sort();
}
for (const file of files(path.join(root, 'src/compiler'))) {
    const relative = path.relative(root, file).split(path.sep).join('/');
    const bytes = fs.readFileSync(file);
    const source = ts.createSourceFile(file, bytes.toString('utf8'), ts.ScriptTarget.Latest, true);
    sources[relative] = crypto.createHash('sha256').update(bytes).digest('hex');
    const site = node => {
        const pos = source.getLineAndCharacterOfPosition(node.getStart(source));
        return { file: relative, line: pos.line + 1, column: pos.character + 1 };
    };
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name && Object.hasOwn(patterns, node.name.text)) {
            patterns[node.name.text].definitions.push(site(node));
        }
        if (ts.isCallExpression(node)) {
            const callee = node.expression;
            const name = ts.isIdentifier(callee) ? callee.text : ts.isPropertyAccessExpression(callee) ? callee.name.text : '';
            if (Object.hasOwn(patterns, name)) patterns[name].calls.push(site(node));
        }
        if (relative === 'src/compiler/parser.ts' && ts.isModuleDeclaration(node) && node.name.text === 'Parser' && ts.isModuleBlock(node.body)) {
            for (const statement of node.body.statements) {
                if (ts.isVariableStatement(statement)) {
                    for (const declaration of statement.declarationList.declarations) {
                        parserState.push({ ...site(declaration), name: declaration.name.getText(source), binding: statement.declarationList.flags & ts.NodeFlags.Const ? 'const' : 'mutable' });
                    }
                }
            }
        }
        if (relative === 'src/compiler/checker.ts' && ts.isFunctionDeclaration(node) && node.name?.text === 'createTypeChecker') {
            for (const statement of node.body.statements) {
                if (ts.isVariableStatement(statement)) {
                    for (const declaration of statement.declarationList.declarations) {
                        checkerState.push({ ...site(declaration), name: declaration.name.getText(source), binding: statement.declarationList.flags & ts.NodeFlags.Const ? 'const' : 'mutable' });
                    }
                }
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
}
if (Object.keys(sources).length !== 77) throw new Error('corpus file count differs');
console.log(JSON.stringify({ pin, reader: ts.version, files: 77, sources, patterns, parserState, checkerState }, null, 2));
