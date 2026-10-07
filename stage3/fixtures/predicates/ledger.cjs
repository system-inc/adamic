// Reconcile the pinned census with bodies, using stock TypeScript's parser and symbols.
// Usage: PREDICATES_TYPESCRIPT=/path/to/typescript.js node ledger.cjs upstream sites.json output.json
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.PREDICATES_TYPESCRIPT);
if (ts.version !== '6.0.3') throw Error('Expected TypeScript 6.0.3');
const root = path.resolve(process.argv[2]);
const sites = JSON.parse(fs.readFileSync(process.argv[3], 'utf8')).filter(s => s.reason === 'a type predicate');
const files = [...new Set(sites.map(s => path.join(root, s.file)))];
const program = ts.createProgram(files, { target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.NodeNext, moduleResolution: ts.ModuleResolutionKind.NodeNext });
const checker = program.getTypeChecker();
const rows = [];
function unwrapped(n) {
    while (n && (ts.isParenthesizedExpression(n) || ts.isAsExpression(n) || ts.isNonNullExpression(n))) n = n.expression;
    return n;
}
function kindOperand(n) {
    n = unwrapped(n);
    return n && (ts.isPropertyAccessExpression(n) && n.name.text === 'kind' || ts.isIdentifier(n) && n.text === 'kind');
}
for (const site of sites) {
    const sf = program.getSourceFile(path.join(root, site.file));
    let predicate;
    function find(n) {
        if (ts.isTypePredicateNode(n) && n.getStart(sf) === site.start) predicate = n;
        ts.forEachChild(n, find);
    }
    find(sf);
    if (!predicate || predicate.end !== site.end) throw Error('Census mismatch: ' + JSON.stringify(site));
    const declaration = predicate.parent;
    let implementation = declaration;
    if (!implementation.body && ts.isFunctionDeclaration(declaration) && declaration.name) {
        const symbol = checker.getSymbolAtLocation(declaration.name);
        implementation = symbol?.declarations?.find(d => ts.isFunctionDeclaration(d) && d.body) || declaration;
    }
    const body = implementation.body;
    const features = new Set();
    const calls = [];
    function visit(n) {
        if (n !== body && ts.isFunctionLike(n)) return;
        if (ts.isTypeOfExpression(n)) features.add('typeof');
        if (ts.isBinaryExpression(n)) {
            if (n.operatorToken.kind === ts.SyntaxKind.InstanceOfKeyword) features.add('instanceof');
            if (n.operatorToken.kind === ts.SyntaxKind.EqualsEqualsEqualsToken && (kindOperand(n.left) || kindOperand(n.right))) features.add('kind ===');
            if (n.operatorToken.kind === ts.SyntaxKind.AmpersandToken) features.add('flags mask');
        }
        if (ts.isSwitchStatement(n) && kindOperand(n.expression)) features.add('kind switch');
        if (ts.isCallExpression(n)) {
            const signature = checker.getResolvedSignature(n);
            const calleePredicate = signature && checker.getTypePredicateOfSignature(signature);
            if (calleePredicate) {
                features.add('predicate call');
                calls.push({ expression: n.expression.getText(sf), signature: checker.signatureToString(signature) });
            }
        }
        ts.forEachChild(n, visit);
    }
    if (body) visit(body);
    let shape = 'something else';
    if (!body) shape = 'no body';
    else if (predicate.assertsModifier && site.file === 'src/compiler/debug.ts') shape = 'Debug.assert-style asserts';
    else if (features.has('instanceof')) shape = 'instanceof';
    else if (features.has('typeof')) shape = 'typeof';
    else if (features.has('flags mask')) shape = 'flags mask';
    else if (features.has('kind ===')) shape = 'kind comparison';
    else if (features.has('predicate call')) shape = 'predicate call';
    const start = implementation.getStart(sf);
    const parameter = predicate.parameterName.getText(sf);
    function directKind(n) {
        n = unwrapped(n);
        if (!n || !ts.isBinaryExpression(n)) return false;
        if (n.operatorToken.kind === ts.SyntaxKind.BarBarToken) return directKind(n.left) && directKind(n.right);
        const left = unwrapped(n.left), right = unwrapped(n.right);
        return n.operatorToken.kind === ts.SyntaxKind.EqualsEqualsEqualsToken
            && ts.isPropertyAccessExpression(left) && left.name.text === 'kind'
            && left.expression.getText(sf) === parameter
            && ts.isPropertyAccessExpression(right) && right.expression.getText(sf) === 'SyntaxKind';
    }
    const direct = body && ts.isBlock(body) && body.statements.length === 1 && ts.isReturnStatement(body.statements[0]) && directKind(body.statements[0].expression);
    rows.push({ file: site.file, line: site.line, column: site.column, start: site.start, end: site.end,
        name: declaration.name?.getText(sf) || '<callback or function type>', predicate: predicate.getText(sf),
        class: shape, features: [...features].sort(), calls,
        implementation_line: body ? sf.getLineAndCharacterOfPosition(start).line + 1 : null,
        direct_kind_chain: !!direct, body: body?.getText(sf) || null });
}
if (rows.length !== 651 || new Set(rows.map(r => r.file)).size !== 31) throw Error('Wrong census population');
fs.writeFileSync(process.argv[4], JSON.stringify(rows, null, 2) + '\n');
const counts = {};
for (const r of rows) counts[r.class] = (counts[r.class] || 0) + 1;
console.log(JSON.stringify({ nodes: rows.length, files: 31, counts, direct_kind_chains: rows.filter(r => r.direct_kind_chain).length }, null, 2));
