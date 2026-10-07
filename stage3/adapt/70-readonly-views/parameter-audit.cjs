'use strict';
// Follow whole-object arguments through checker-resolved function bodies.
// Unresolved calls, aliases, casts and returned receivers are conservative declines.
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const path = require('node:path');
module.exports = function auditParameters(program, checker, tree, specifications) {
    if (ts.version !== '6.0.3') throw new Error('expected stock TypeScript 6.0.3');
    const graph = new Map();
    function where(node) {
        const sf = node.getSourceFile(), p = sf.getLineAndCharacterOfPosition(node.getStart(sf));
        return `${path.relative(tree, sf.fileName)}:${p.line + 1}:${p.character + 1}`;
    }
    function assigned(node) {
        let current = node;
        while (current.parent) {
            const p = current.parent;
            if (ts.isParenthesizedExpression(p) || ts.isArrayLiteralExpression(p) || ts.isObjectLiteralExpression(p) ||
                ts.isPropertyAssignment(p) || ts.isSpreadAssignment(p) || ts.isSpreadElement(p)) { current = p; continue; }
            if (ts.isBinaryExpression(p) && p.left === current && p.operatorToken.kind >= ts.SyntaxKind.FirstAssignment &&
                p.operatorToken.kind <= ts.SyntaxKind.LastAssignment) return true;
            if ((ts.isPrefixUnaryExpression(p) || ts.isPostfixUnaryExpression(p)) &&
                [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken].includes(p.operator)) return true;
            if (ts.isDeleteExpression(p) || ((ts.isForOfStatement(p) || ts.isForInStatement(p)) && p.initializer === current)) return true;
            return false;
        }
        return false;
    }
    function inspect(parameter) {
        if (graph.has(parameter)) return;
        const fn = parameter.parent;
        const record = { node: parameter, where: where(parameter), reads: [], writes: [], escapes: [], calls: [] };
        graph.set(parameter, record);
        if (!ts.isIdentifier(parameter.name) || !fn.body) {
            record.escapes.push({ where: where(parameter), reason: 'no uniquely named parameter body' });
            return;
        }
        const symbol = checker.getSymbolAtLocation(parameter.name);
        function visit(node) {
            if (ts.isIdentifier(node) && node !== parameter.name && checker.getSymbolAtLocation(node) === symbol) {
                let current = node;
                while (current.parent && (ts.isParenthesizedExpression(current.parent) || ts.isNonNullExpression(current.parent)))
                    current = current.parent;
                const p = current.parent;
                if (ts.isPropertyAccessExpression(p) && p.expression === current) {
                    let nested = p;
                    while ((ts.isPropertyAccessExpression(nested.parent) || ts.isElementAccessExpression(nested.parent)) &&
                        nested.parent.expression === nested) nested = nested.parent;
                    const called = ts.isCallExpression(nested.parent) && nested.parent.expression === nested;
                    const nestedSymbol = ts.isPropertyAccessExpression(nested) ? checker.getSymbolAtLocation(nested.name) : undefined;
                    const stockMutation = called && nestedSymbol?.declarations?.some(d => ts.isMethodSignature(d) &&
                        d.getSourceFile().isDeclarationFile && path.basename(d.getSourceFile().fileName).startsWith('lib.')) &&
                        ['push', 'pop', 'shift', 'unshift', 'splice', 'sort', 'reverse', 'copyWithin', 'fill', 'set', 'add', 'delete', 'clear'].includes(nestedSymbol.name);
                    if (assigned(nested) || stockMutation) record.writes.push(where(stockMutation ? nested.parent : nested));
                    else {
                        const property = checker.getSymbolAtLocation(p.name);
                        // Accessors may have effects. TextRange.pos/end are plain numeric fields.
                        const method = ts.isCallExpression(p.parent) && p.parent.expression === p;
                        const stockMethod = property?.declarations?.some(d => ts.isMethodSignature(d) &&
                            d.getSourceFile().isDeclarationFile && path.basename(d.getSourceFile().fileName).startsWith('lib.'));
                        const mutations = new Set(['push', 'pop', 'shift', 'unshift', 'splice', 'sort', 'reverse',
                            'copyWithin', 'fill', 'set', 'add', 'delete', 'clear']);
                        if (method && stockMethod && mutations.has(property.name)) record.writes.push(where(p.parent));
                        else if (!property?.declarations?.every(d => ts.isPropertySignature(d) || ts.isPropertyDeclaration(d)))
                            record.escapes.push({ where: where(p), reason: 'accessor, method or unresolved property' });
                        else {
                            const value = checker.getTypeAtLocation(p);
                            const primitive = t => t.isUnion() ? t.types.every(primitive) :
                                !!(t.flags & (ts.TypeFlags.StringLike | ts.TypeFlags.NumberLike | ts.TypeFlags.BooleanLike |
                                    ts.TypeFlags.BigIntLike | ts.TypeFlags.ESSymbolLike | ts.TypeFlags.Null | ts.TypeFlags.Undefined | ts.TypeFlags.Never));
                            if (!primitive(value)) record.escapes.push({ where: where(p), reason: 'nested object requires its own mutation and alias proof' });
                            else record.reads.push(where(p));
                        }
                    }
                } else if (ts.isElementAccessExpression(p) && p.expression === current) {
                    if (assigned(p)) record.writes.push(where(p));
                    else record.escapes.push({ where: where(p), reason: 'element alias requires a separate container proof' });
                } else if (ts.isCallExpression(p) && p.arguments.includes(current)) {
                    const signature = checker.getResolvedSignature(p);
                    const index = p.arguments.indexOf(current);
                    const target = signature?.parameters[index]?.valueDeclaration;
                    if (target && ts.isParameter(target) && target.parent.body &&
                        path.resolve(target.getSourceFile().fileName).startsWith(tree + path.sep + 'src' + path.sep)) {
                        record.calls.push(target);
                        inspect(target);
                    } else record.escapes.push({ where: where(node), reason: 'unresolved or external consumer' });
                } else if ((ts.isIfStatement(p) && p.expression === current) ||
                    (ts.isConditionalExpression(p) && p.condition === current) ||
                    (ts.isPrefixUnaryExpression(p) && p.operator === ts.SyntaxKind.ExclamationToken) ||
                    (ts.isBinaryExpression(p) && [ts.SyntaxKind.EqualsEqualsToken, ts.SyntaxKind.EqualsEqualsEqualsToken,
                        ts.SyntaxKind.ExclamationEqualsToken, ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(p.operatorToken.kind))) {
                    record.reads.push(where(node));
                } else record.escapes.push({ where: where(node), reason: ts.SyntaxKind[p.kind] });
            }
            ts.forEachChild(node, visit);
        }
        visit(fn);
    }
    const selected = specifications.map(spec => {
        const sf = program.getSourceFile(path.join(tree, spec.file));
        const matches = [];
        function visit(n) {
            if (ts.isFunctionDeclaration(n) && n.name?.text === spec.function && n.body) matches.push(n);
            ts.forEachChild(n, visit);
        }
        visit(sf);
        if (matches.length !== 1) throw new Error(`expected one body for ${spec.function}`);
        const fn = matches[0], parameter = fn.parameters.find(p => ts.isIdentifier(p.name) && p.name.text === spec.parameter);
        if (!parameter?.type) throw new Error(`missing parameter ${spec.function}.${spec.parameter}`);
        const exported = !!(ts.getCombinedModifierFlags(fn) & ts.ModifierFlags.Export);
        const internal = ts.getJSDocTags(fn).some(t => t.tagName.text === 'internal');
        if (exported && !internal && !spec.public) throw new Error(`public parameter excluded: ${spec.function}.${spec.parameter}`);
        inspect(parameter);
        const reachable = new Set();
        function collect(p) {
            if (reachable.has(p)) return;
            reachable.add(p);
            graph.get(p).calls.forEach(collect);
        }
        collect(parameter);
        const dependencies = [...reachable].map(p => graph.get(p));
        return { ...spec, node: parameter, where: where(parameter),
            writes: dependencies.flatMap(d => d.writes), escapes: dependencies.flatMap(d => d.escapes),
            dependencies: dependencies.map(({ node, calls, ...d }) => ({ ...d, calls: calls.map(where) })) };
    });
    return selected;
};
