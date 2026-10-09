'use strict';
// A narrowly bounded proof for an inferred local receiver. Explicit any,
// assertions, retained aliases and whole-object forwarding remain declines.
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const path = require('node:path');
module.exports = function localAliasAuditor(checker, tree) {
    const cache = new Map();
    function where(n) {
        const sf = n.getSourceFile(), p = sf.getLineAndCharacterOfPosition(n.getStart(sf));
        return `${path.relative(tree, sf.fileName)}:${p.line + 1}:${p.character + 1}`;
    }
    function write(n) {
        let current = n;
        while (current.parent) {
            const p = current.parent;
            if (ts.isParenthesizedExpression(p) || ts.isArrayLiteralExpression(p) || ts.isObjectLiteralExpression(p) ||
                ts.isPropertyAssignment(p) || ts.isSpreadAssignment(p) || ts.isSpreadElement(p)) { current = p; continue; }
            if (ts.isBinaryExpression(p) && p.left === current && p.operatorToken.kind >= ts.SyntaxKind.FirstAssignment &&
                p.operatorToken.kind <= ts.SyntaxKind.LastAssignment) return p;
            if ((ts.isPrefixUnaryExpression(p) || ts.isPostfixUnaryExpression(p)) &&
                [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken].includes(p.operator)) return p;
            if (ts.isDeleteExpression(p) || ((ts.isForOfStatement(p) || ts.isForInStatement(p)) && p.initializer === current)) return p;
            return undefined;
        }
    }
    return function inferredLocal(node, member) {
        let current = node;
        while (current.parent) {
            const p = current.parent;
            if (ts.isParenthesizedExpression(p) || ts.isNonNullExpression(p) ||
                (ts.isBinaryExpression(p) && [ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.BarBarToken,
                    ts.SyntaxKind.QuestionQuestionToken].includes(p.operatorToken.kind)) ||
                (ts.isConditionalExpression(p) && p.condition !== current)) { current = p; continue; }
            break;
        }
        const assignment = current.parent;
        if (!ts.isBinaryExpression(assignment) || assignment.right !== current ||
            assignment.operatorToken.kind !== ts.SyntaxKind.EqualsToken || !ts.isIdentifier(assignment.left)) return undefined;
        const symbol = checker.getSymbolAtLocation(assignment.left), declaration = symbol?.valueDeclaration;
        if (!declaration || !ts.isVariableDeclaration(declaration) || declaration.type || declaration.initializer ||
            !ts.isIdentifier(declaration.name) || !(declaration.parent.flags & ts.NodeFlags.Let)) return undefined;
        let scope = declaration.parent;
        while (scope && !ts.isFunctionLike(scope)) scope = scope.parent;
        if (!scope?.body) return undefined;
        if (!cache.has(symbol)) cache.set(symbol, new Map());
        const members = cache.get(symbol);
        if (members.has(member)) return members.get(member);
        const proof = { declaration: where(declaration), member, reads: [], rebindings: [], writes: [], escapes: [] };
        members.set(member, proof);
        function visit(n) {
            if (ts.isIdentifier(n) && n !== declaration.name && checker.getSymbolAtLocation(n) === symbol) {
                let value = n;
                while (value.parent && (ts.isParenthesizedExpression(value.parent) || ts.isNonNullExpression(value.parent))) value = value.parent;
                const p = value.parent;
                const name = ts.isPropertyAccessExpression(p) && p.expression === value ? p.name.text :
                    ts.isElementAccessExpression(p) && p.expression === value && ts.isStringLiteralLike(p.argumentExpression) ? p.argumentExpression.text : undefined;
                if (name === member) {
                    const writer = write(p);
                    if (writer) proof.writes.push(where(writer));
                    else proof.reads.push(where(p));
                } else if (ts.isBinaryExpression(p) && p.left === value && p.operatorToken.kind === ts.SyntaxKind.EqualsToken) {
                    proof.rebindings.push(where(p));
                } else if ((ts.isIfStatement(p) && p.expression === value) ||
                    (ts.isConditionalExpression(p) && p.condition === value) ||
                    (ts.isPrefixUnaryExpression(p) && p.operator === ts.SyntaxKind.ExclamationToken) || ts.isTypeOfExpression(p) ||
                    (ts.isBinaryExpression(p) && [ts.SyntaxKind.EqualsEqualsEqualsToken, ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(p.operatorToken.kind))) {
                    proof.reads.push(where(n));
                } else proof.escapes.push({ where: where(n), reason: ts.SyntaxKind[p.kind] });
            }
            ts.forEachChild(n, visit);
        }
        visit(scope);
        return proof;
    };
};
