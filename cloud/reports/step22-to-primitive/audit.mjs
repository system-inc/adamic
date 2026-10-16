// Audit conversion operands with the pinned upstream compiler's own checker.
// Arguments: TypeScript checkout, TypeScript API module, @types directory, output JSON.
import fs from 'node:fs';
import path from 'node:path';
import {createHash} from 'node:crypto';
import {pathToFileURL} from 'node:url';

const {default: ts} = await import(pathToFileURL(path.resolve(process.argv[3])).href);
const root = path.resolve(process.argv[2]);
const config = ts.readConfigFile(root + '/src/compiler/tsconfig.json', ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, root + '/src/compiler');
parsed.options.typeRoots = [path.resolve(process.argv[4])];
parsed.options.ignoreDeprecations = '6.0';
const program = ts.createProgram(parsed.fileNames, parsed.options);
const checker = program.getTypeChecker();
const files = program.getSourceFiles().filter(file =>
    file.fileName.startsWith(root + '/src/compiler/') &&
    !file.isDeclarationFile && !file.fileName.endsWith('.generated.ts'));
const rows = [], classes = [], literalMethods = [];
let operands = 0, sites = 0, templates = 0, additions = 0, compounds = 0;

function category(type, seen = new Set()) {
    if (seen.has(type)) return 'unresolved';
    seen.add(type);
    if (type.flags & ts.TypeFlags.Any) return 'any';
    if (type.flags & ts.TypeFlags.Unknown) return 'unknown';
    // A primitive intersected with a branding object is still a primitive value.
    if (type.isIntersection()) {
        const categories = type.types.map(member => category(member, new Set(seen)));
        if (categories.includes('primitive')) return 'primitive';
        // Y & {} proves presence, not that Y is an ECMAScript Object.
        if (categories.includes('unconstrained')) return 'unconstrained';
    }
    if (type.isUnionOrIntersection()) {
        const categories = type.types.map(member => category(member, new Set(seen)));
        return categories.includes('object') ? 'object' :
            categories.find(value => value !== 'primitive') ?? 'primitive';
    }
    if (type.flags & ts.TypeFlags.TypeParameter) {
        const constraint = checker.getBaseConstraintOfType(type);
        return constraint ? category(constraint, seen) : 'unconstrained';
    }
    return type.flags & ts.TypeFlags.Object ? 'object' : 'primitive';
}

for (const file of files) {
    function add(expression, kind) {
        operands++;
        const type = checker.getTypeAtLocation(expression);
        const classification = category(type);
        if (classification === 'primitive') return;
        const position = file.getLineAndCharacterOfPosition(expression.getStart(file));
        rows.push({
            file: path.relative(root, file.fileName),
            line: position.line + 1,
            column: position.character + 1,
            kind,
            category: classification,
            type: checker.typeToString(type, expression, ts.TypeFormatFlags.NoTruncation),
            expression: expression.getText(file).replace(/\s+/g, ' '),
        });
    }
    function visit(node) {
        if (ts.isTemplateExpression(node) && !ts.isTaggedTemplateExpression(node.parent)) {
            sites++;
            templates++;
            for (const span of node.templateSpans) add(span.expression, 'template');
        }
        if (ts.isBinaryExpression(node) &&
            [ts.SyntaxKind.PlusToken, ts.SyntaxKind.PlusEqualsToken].includes(node.operatorToken.kind)) {
            sites++;
            if (node.operatorToken.kind === ts.SyntaxKind.PlusToken) additions++;
            else compounds++;
            const operator = ts.tokenToString(node.operatorToken.kind);
            add(node.left, operator + ' left');
            add(node.right, operator + ' right');
        }
        if (ts.isClassDeclaration(node) || ts.isClassExpression(node)) {
            for (const member of node.members) {
                if (!member.name || !['toString', 'valueOf'].includes(member.name.getText(file))) continue;
                const position = file.getLineAndCharacterOfPosition(member.getStart(file));
                classes.push({
                    file: path.relative(root, file.fileName),
                    line: position.line + 1,
                    class: node.name?.text ?? '<anonymous>',
                    method: member.name.getText(file),
                });
            }
        }
        if (ts.isPropertyAssignment(node) && node.name.getText(file) === 'toString' &&
            ts.isObjectLiteralExpression(node.parent)) {
            const position = file.getLineAndCharacterOfPosition(node.getStart(file));
            literalMethods.push({file: path.relative(root, file.fileName), line: position.line + 1});
        }
        ts.forEachChild(node, visit);
    }
    visit(file);
}
const diagnostics = ts.getPreEmitDiagnostics(program);
const counts = Object.fromEntries(['object', 'any', 'unknown', 'unconstrained', 'unresolved'].map(
    classification => [classification, rows.filter(row => row.category === classification).length]));
const result = {
    version: ts.version, files: files.length, sites, templates, additions, compounds, operands,
    classes, literalMethods,
    manifest: files.map(file => ({
        file: path.relative(root, file.fileName),
        sha256: createHash('sha256').update(fs.readFileSync(file.fileName)).digest('hex'),
    })),
    counts, rows,
    diagnostics: diagnostics.map(diagnostic => ({
        file: diagnostic.file ? path.relative(root, diagnostic.file.fileName) : undefined,
        line: diagnostic.file && diagnostic.start !== undefined ?
            diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start).line + 1 : 0,
        code: diagnostic.code,
        message: ts.flattenDiagnosticMessageText(diagnostic.messageText, ' '),
    })),
};
fs.writeFileSync(process.argv[5], JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify({files: files.length, sites, operands, classes, counts, diagnostics: diagnostics.length}));
