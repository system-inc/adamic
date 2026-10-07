'use strict';
// Read-only symbol tracing for the original 434 pending TS2345 rows.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error('expected stock TypeScript 6.0.3');
const [tree, ledgerPath, output] = process.argv.slice(2);
const ledger = JSON.parse(fs.readFileSync(ledgerPath, 'utf8'));
const rows = ledger.rows.filter(r => r.scope === 'pending-index-origin-review' || r.origin_review_cohort === 'original-434');
if (rows.length !== 434) throw new Error('expected original pending population of 434');
const program = ts.createProgram(ts.sys.readDirectory(path.join(tree, 'src/compiler'), ['.ts']), {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    noImplicitReturns: true, noFallthroughCasesInSwitch: true, erasableSyntaxOnly: true,
    verbatimModuleSyntax: true, allowImportingTsExtensions: true, noEmit: true,
    module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [],
});
const checker = program.getTypeChecker();
const assignments = new Map();
function symbol(node) {
    let s = checker.getSymbolAtLocation(node);
    if (s?.flags & ts.SymbolFlags.Alias) s = checker.getAliasedSymbol(s);
    return s;
}
function scan(n) {
    if (ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isIdentifier(n.left)) {
        const s = symbol(n.left);
        if (s) {
            if (!assignments.has(s)) assignments.set(s, []);
            assignments.get(s).push(n.right);
        }
    }
    ts.forEachChild(n, scan);
}
for (const file of program.getSourceFiles()) if (file.fileName.startsWith(tree + '/src/compiler/')) scan(file);
function loc(node) {
    const file = node.getSourceFile();
    const p = file.getLineAndCharacterOfPosition(node.getStart());
    return {file: path.relative(tree, file.fileName), line: p.line + 1, column: p.character + 1,
        kind: ts.SyntaxKind[node.kind], text: node.getText().slice(0, 1800)};
}
function at(node, position) {
    let result = node;
    ts.forEachChild(node, child => {
        if (child.getStart() <= position && position < child.end) result = at(child, position);
    });
    return result;
}
function returnValues(declaration) {
    const found = [];
    if (!declaration?.body) return found;
    if (!ts.isBlock(declaration.body)) return [declaration.body];
    function walk(n) {
        if (ts.isFunctionLike(n)) return;
        if (ts.isReturnStatement(n) && n.expression) found.push(n.expression);
        ts.forEachChild(n, walk);
    }
    walk(declaration.body);
    return found;
}
function trace(node, steps, terminals, active = new Set(), depth = 0, via = 'argument') {
    if (!node) return;
    const identity = node.getSourceFile().fileName + ':' + node.pos;
    if (active.has(identity) || depth > 12) {
        terminals.push({reason: 'cycle-or-depth-limit', ...loc(node)});
        return;
    }
    const nextActive = new Set(active); nextActive.add(identity);
    steps.push({via, ...loc(node)});
    const go = (n, reason) => trace(n, steps, terminals, nextActive, depth + 1, reason);
    if (ts.isElementAccessExpression(node)) {
        terminals.push({reason: 'indexed-read', ...loc(node)}); return;
    }
    if (ts.isParenthesizedExpression(node) || ts.isAsExpression(node) || ts.isNonNullExpression(node)) {
        go(node.expression, 'transparent-expression'); return;
    }
    if (ts.isConditionalExpression(node)) {
        go(node.whenTrue, 'conditional-true-value'); go(node.whenFalse, 'conditional-false-value'); return;
    }
    if (ts.isBinaryExpression(node)) {
        const k = node.operatorToken.kind;
        if ([ts.SyntaxKind.EqualsToken, ts.SyntaxKind.BarBarToken, ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.QuestionQuestionToken].includes(k)) {
            if (k !== ts.SyntaxKind.EqualsToken) go(node.left, 'value-branch');
            go(node.right, 'assigned-or-branch-value'); return;
        }
    }
    if (ts.isArrayLiteralExpression(node)) {
        for (const e of node.elements) go(e, 'array-element'); return;
    }
    if (ts.isIdentifier(node)) {
        const s = symbol(node);
        const declarations = s?.declarations || [];
        let followed = false;
        for (const d of declarations) {
            steps.push({via: 'symbol-declaration', ...loc(d)});
            if (ts.isVariableDeclaration(d) && d.initializer) {go(d.initializer, 'variable-initializer'); followed = true;}
            if (ts.isBindingElement(d)) {
                const pattern = d.parent, owner = pattern.parent;
                if (ts.isVariableDeclaration(owner) && owner.initializer) {
                    terminals.push({reason: ts.isArrayBindingPattern(pattern) ? 'destructured-index' : 'destructured-property', ...loc(d)});
                    go(owner.initializer, 'destructured-source'); followed = true;
                } else if (ts.isParameter(owner)) {
                    terminals.push({reason: 'callback-destructured-parameter', ...loc(owner)}); followed = true;
                } else if (ts.isVariableDeclaration(owner)) {
                    const loop = owner.parent.parent;
                    if (ts.isForOfStatement(loop)) {go(loop.expression, 'for-of-collection');followed=true;}
                }
            }
            if (ts.isVariableDeclaration(d) && !d.initializer) {
                const loop = d.parent.parent;
                if (ts.isForOfStatement(loop)) {go(loop.expression, 'for-of-collection');followed=true;}
            }
            if (ts.isParameter(d)) {
                const fn = d.parent;
                let container = fn.parent;
                while (container && (ts.isParenthesizedExpression(container) || ts.isAsExpression(container))) container = container.parent;
                terminals.push({reason: ts.isCallExpression(container) ? 'callback-parameter' : 'declared-parameter', ...loc(d),
                    supplied_by: container ? loc(container) : null});
                if (ts.isCallExpression(container)) {
                    if (ts.isPropertyAccessExpression(container.expression)) go(container.expression.expression, 'callback-collection');
                    else if (container.arguments.length) go(container.arguments[0], 'callback-collection');
                }
                followed = true;
            }
            if ((ts.isPropertySignature(d) || ts.isPropertyDeclaration(d)) && d.questionToken) {
                terminals.push({reason:'optional-property',...loc(d)});followed=true;
            }
        }
        for (const rhs of assignments.get(s) || []) {go(rhs, 'assignment-to-same-symbol');followed=true;}
        if (!followed) terminals.push({reason:'unfollowed-symbol',...loc(node)});
        return;
    }
    if (ts.isPropertyAccessExpression(node)) {
        const declarations = symbol(node.name)?.declarations || [];
        let followed = false;
        for (const d of declarations) {
            steps.push({via:'property-declaration',...loc(d)});
            if (d.questionToken) {terminals.push({reason:'optional-property',...loc(d)});followed=true;}
            if (ts.isPropertyAssignment(d)) {go(d.initializer,'property-initializer');followed=true;}
        }
        if (!followed) {terminals.push({reason:'property-value',...loc(node)});go(node.expression,'receiver-origin');}
        return;
    }
    if (ts.isCallExpression(node)) {
        const callee = node.expression.getText();
        if (ts.isPropertyAccessExpression(node.expression)) {
            const method = node.expression.name.text;
            if (['keys', 'values', 'entries'].includes(method)) {
                terminals.push({reason:'collection-iterator',...loc(node), stock_type:checker.typeToString(checker.getTypeAtLocation(node))});return;
            }
            if (['sort','slice','filter'].includes(method)) {go(node.expression.expression,'collection-preserving-method');return;}
            if (method === 'map' && node.arguments[0] && ts.isFunctionLike(node.arguments[0])) {
                for (const value of returnValues(node.arguments[0])) go(value,'mapped-element-result');return;
            }
        }
        if (['arrayFrom','filter','concatenate','sortAndDeduplicate'].includes(callee)) {
            for (const argument of node.arguments.slice(0, callee==='concatenate'?2:1)) go(argument,'generic-collection-input');return;
        }
        if (callee === 'resolveSymbol') {
            go(node.arguments[0],'resolveSymbol-input');return;
        }
        const signature = checker.getResolvedSignature(node);
        let declaration = signature?.declaration;
        if (declaration && !declaration.body) {
            const implementation = symbol(ts.isPropertyAccessExpression(node.expression) ? node.expression.name : node.expression)?.declarations?.find(d => d.body);
            if (implementation) declaration = implementation;
        }
        if (declaration) steps.push({via:'resolved-callee',...loc(declaration)});
        const returns = returnValues(declaration);
        if (returns.length) {
            for (const value of returns) go(value, 'callee-return-value');
        } else terminals.push({reason:'call-result',...loc(node), declaration: declaration ? loc(declaration) : null});
        return;
    }
    if (ts.isObjectLiteralExpression(node)) {
        terminals.push({reason:'object-shape',...loc(node)});return;
    }
    terminals.push({reason:'value-expression',...loc(node)});
}
const result = [];
for (const row of rows) {
    const file = program.getSourceFile(path.join(tree, row.argument.file));
    const position = file.getPositionOfLineAndCharacter(row.argument.line - 1, row.argument.column - 1);
    let argument = at(file, position);
    while (argument.parent && argument.end - argument.getStart() < row.argument.text.length) argument = argument.parent;
    const steps = [], terminals = [];
    trace(argument, steps, terminals);
    result.push({id:row.id, argument:loc(argument), steps, terminals});
}
fs.writeFileSync(output, JSON.stringify(result, null, 2) + '\n');
process.stdout.write(`Traced ${result.length} pending arguments without modifying sources.\n`);
