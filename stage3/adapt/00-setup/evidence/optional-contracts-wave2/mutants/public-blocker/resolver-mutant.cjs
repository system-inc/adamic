'use strict';
// Additional structural resolution runs only after the original resolver settles.
module.exports = function internalResolver(ts, checker, owned, select, options, program) {
    const fs = require('node:fs');
    const path = require('node:path');
    const forbidden = ts.TypeFlags.Any | ts.TypeFlags.Unknown | ts.TypeFlags.TypeParameter | ts.TypeFlags.IndexedAccess;
    const parts = type => type.isUnion() ? type.types : [type];
    const isUndefined = type => !!(type.flags & ts.TypeFlags.Undefined);
    function ownerPath(node) {
        const names = [];
        for (let current = node; current; current = current.parent) {
            if (current.name && !(ts.isModuleDeclaration(current) && current.name.getText() === 'ts')) names.unshift(current.name.getText());
        }
        return JSON.stringify(names);
    }
    // The installed stock compiler's own declaration build decides API membership.
    const apiFile = path.join(path.dirname(ts.getDefaultLibFilePath(options)), 'typescript.d.ts');
    const api = ts.createSourceFile(apiFile, fs.readFileSync(apiFile, 'utf8'), ts.ScriptTarget.Latest, true);
    const publicPaths = new Set();
    function collect(node) {
        if (ts.isPropertySignature(node) || ts.isPropertyDeclaration(node) || ts.isMethodSignature(node)) publicPaths.add(ownerPath(node));
        ts.forEachChild(node, collect);
    }
    collect(api);
    const isPublic = owner => publicPaths.has(ownerPath(owner));
    let blockedOwners;
    function blockers() {
        if (blockedOwners) return blockedOwners;
        blockedOwners = new Map();
        function inspect(source, target, expression, depth = 0, seen = new Set()) {
            if (!source || !target || depth > 5 || (source.flags & forbidden) || (target.flags & forbidden)) return;
            for (const present of parts(source).filter(type => !isUndefined(type))) {
                const arm = objectArm(present, target, expression);
                if (!arm || !(arm.flags & ts.TypeFlags.Object) || !(present.flags & ts.TypeFlags.Object)) continue;
                const key = present.id + ':' + arm.id;
                if (seen.has(key)) continue;
                seen.add(key);
                if (checker.getSignaturesOfType(arm, ts.SignatureKind.Call).length) continue;
                const pairs = checker.getPropertiesOfType(arm).flatMap(property => {
                    const actual = checker.getPropertyOfType(present, property.name);
                    if (!actual) return [];
                    const value = declaredValue(actual, expression), expected = declaredValue(property, expression);
                    return value && expected ? [{ property, value, expected }] : [];
                });
                const publicBlockers = pairs.filter(({ property, value }) => property.declarations?.some(owner =>
                    isPublic(owner) && owner.questionToken && owner.type && parts(value).some(isUndefined) &&
                    !checker.isTypeAssignableTo(checker.getUndefinedType(), checker.getTypeFromTypeNode(owner.type))))
                    .flatMap(({ property }) => property.declarations.filter(isPublic).map(owner => JSON.parse(ownerPath(owner))));
                if (publicBlockers.length) for (const { property, value } of pairs) {
                    if (!parts(value).some(isUndefined)) continue;
                    for (const owner of property.declarations || []) if (!isPublic(owner) && owner.questionToken && owner.type &&
                        (ts.isPropertySignature(owner) || ts.isPropertyDeclaration(owner)) &&
                        !checker.isTypeAssignableTo(checker.getUndefinedType(), checker.getTypeFromTypeNode(owner.type))) blockedOwners.set(owner, publicBlockers);
                }
                for (const { value, expected } of pairs) if (!checker.isTypeAssignableTo(value, expected)) inspect(value, expected, expression, depth + 1, seen);
            }
        }
        for (const file of program.getSourceFiles()) {
            if (!owned.has(path.resolve(file.fileName))) continue;
            function visit(node) {
                if (ts.isObjectLiteralExpression(node)) {
                    const target = checker.getContextualType(node);
                    if (target) inspect(checker.getTypeAtLocation(node), target, node);
                }
                ts.forEachChild(node, visit);
            }
            visit(file);
        }
        return blockedOwners;
    }
    const isBlocked = owner => false;
    function deferred() {
        return [...blockers()].map(([owner, publicBlockers]) => {
            const file = owner.getSourceFile(), point = file.getLineAndCharacterOfPosition(owner.getStart(file));
            return { file: path.relative(path.resolve(directory, '../..'), file.fileName), line: point.line + 1,
                name: owner.name.getText(file), owner_path: JSON.parse(ownerPath(owner)), public_blockers: publicBlockers,
                reason: 'A public optional slot in this construction remains incompatible; a partial internal edit would expose a new diagnostic chain' };
        });
    }
    function internal(property, value, finding) {
        if (!property?.declarations?.length || property.declarations.some(owner => isPublic(owner) || isBlocked(owner) ||
            !(ts.isPropertySignature(owner) || ts.isPropertyDeclaration(owner)) || !owned.has(path.resolve(owner.getSourceFile().fileName)))) return false;
        if (select(property, value, finding)) return true;
        return genericObjectOwner(property, value, finding);
    }
    function genericObjectOwner(property, value, finding) {
        let parameterized = false;
        function visit(node) {
            if (ts.isTypeNode(node) && (checker.getTypeFromTypeNode(node).flags & ts.TypeFlags.TypeParameter)) parameterized = true;
            ts.forEachChild(node, visit);
        }
        for (const owner of property.declarations) if (owner.type) visit(owner.type);
        if (!parameterized) return false;
        return select(property, value, { ...finding, resolution_group: 'generic-owner' }, true);
    }
    function declaredValue(property, expression) {
        const declaration = property.declarations?.[0];
        if (property.flags & ts.SymbolFlags.Optional) {
            if (property.declarations?.length !== 1 || !declaration.type) return;
            return checker.getTypeFromTypeNode(declaration.type);
        }
        return checker.getTypeOfSymbolAtLocation(property, expression);
    }
    function unproven(node) {
        if (!node) return false;
        let indexed = ts.isElementAccessExpression(node);
        ts.forEachChild(node, child => { if (unproven(child)) indexed = true; });
        if (!indexed) return false;
        return !(ts.isConditionalExpression(node) && [node.whenTrue, node.whenFalse].some(branch => isUndefined(checker.getTypeAtLocation(branch))));
    }
    function objectArm(source, target, expression) {
        const candidates = parts(target).filter(type => !isUndefined(type));
        if (candidates.length === 1) return candidates[0];
        // Select a unique union arm only through an existing literal discriminant.
        const compatible = candidates.filter(candidate => checker.getPropertiesOfType(candidate).some(property => {
            const actual = checker.getPropertyOfType(source, property.name);
            if (!actual) return false;
            const value = checker.getTypeOfSymbolAtLocation(actual, expression);
            const expected = checker.getTypeOfSymbolAtLocation(property, expression);
            return (value.flags & (ts.TypeFlags.StringLiteral | ts.TypeFlags.NumberLiteral | ts.TypeFlags.EnumLiteral)) &&
                (expected.flags & (ts.TypeFlags.StringLiteral | ts.TypeFlags.NumberLiteral | ts.TypeFlags.EnumLiteral)) &&
                checker.isTypeAssignableTo(value, expected);
        }));
        if (compatible.length === 1) return compatible[0];
    }
    function resolve(source, target, expression, finding, depth = 0, seen = new Set()) {
        if (!source || !target || depth > 5 || (source.flags & forbidden) || (target.flags & forbidden)) return false;
        let resolved = false;
        for (const present of parts(source).filter(type => !isUndefined(type))) {
            if (present.flags & forbidden) continue;
            const arm = objectArm(present, target, expression);
            if (!arm || !(arm.flags & ts.TypeFlags.Object) || !(present.flags & ts.TypeFlags.Object)) continue;
            const key = present.id + ':' + arm.id;
            if (seen.has(key)) continue;
            seen.add(key);
            // Callable variance and generic specializations are not structural data slots.
            if (checker.getSignaturesOfType(arm, ts.SignatureKind.Call).length || checker.getSignaturesOfType(arm, ts.SignatureKind.Construct).length) continue;
            for (const property of checker.getPropertiesOfType(arm)) {
                const actual = checker.getPropertyOfType(present, property.name);
                if (!actual) continue;
                const declaration = actual.declarations?.[0];
                if (declaration && ts.isPropertyAssignment(declaration) && unproven(declaration.initializer)) continue;
                const value = declaredValue(actual, expression);
                if (!value) continue;
                resolved = internal(property, value, finding) || resolved;
                const expected = declaredValue(property, expression);
                if (expected && !checker.isTypeAssignableTo(value, expected)) {
                    resolved = resolve(value, expected, expression, finding, depth + 1, seen) || resolved;
                }
            }
        }
        return resolved;
    }
    function at(node, position) {
        let result = node;
        ts.forEachChild(node, child => { if (child.pos <= position && position < child.end) result = at(child, position); });
        return result;
    }
    function contextual(diagnostic, finding, group) {
        const node = at(diagnostic.file, diagnostic.start);
        for (let current = node; current; current = current.parent) {
            if (ts.isShorthandPropertyAssignment(current)) return false;
            let expression = current;
            if (ts.isPropertyAssignment(current)) expression = current.initializer;
            if (ts.isReturnStatement(current)) expression = current.expression;
            if (ts.isVariableDeclaration(current)) expression = current.initializer;
            if (ts.isBinaryExpression(current) && current.operatorToken.kind === ts.SyntaxKind.EqualsToken) expression = current.right;
            if (expression && ts.isBinaryExpression(expression) && expression.operatorToken.kind === ts.SyntaxKind.QuestionQuestionEqualsToken) expression = expression.right;
            if (!expression || !ts.isExpressionNode(expression)) continue;
            const target = checker.getContextualType(expression);
            if (!target || (target.flags & forbidden)) continue;
            const source = checker.getTypeAtLocation(expression);
            const arms = parts(target).filter(type => !isUndefined(type));
            if (!arms.some(type => checker.getPropertiesOfType(type).length)) continue;
            return resolve(source, target, expression, { ...finding, resolution_group: group });
        }
        return false;
    }
    function nullableArgument(diagnostic, finding) {
        return contextual(diagnostic, finding, 'nullable-argument');
    }
    function nullableAssignment(diagnostic, finding) {
        return contextual(diagnostic, finding, 'nullable-assignment');
    }
    function otherCode(diagnostic, finding) {
        return contextual(diagnostic, finding, 'unsupported-code');
    }
    function nestedInitializer(diagnostic, finding) {
        return contextual(diagnostic, finding, 'nested-expression');
    }
    function nestedResult(diagnostic, finding) {
        return contextual(diagnostic, finding, 'nested-generic-result');
    }
    function diagnostic(diagnostic, finding) {
        if (!diagnostic.file || diagnostic.start === undefined || !finding.message.includes("Type 'undefined' is not assignable to type")) return;
        if (diagnostic.code === 2379) return nullableArgument(diagnostic, finding);
        if ([2322, 2345].includes(diagnostic.code)) return otherCode(diagnostic, finding);
        if (diagnostic.code !== 2375) return;
        const node = at(diagnostic.file, diagnostic.start);
        if (ts.isIdentifier(node) && ts.isPropertyAssignment(node.parent) && node.parent.name === node) return nestedInitializer(diagnostic, finding);
        for (let current = node; current; current = current.parent) {
            if (ts.isBinaryExpression(current)) {
                if (current.operatorToken.kind === ts.SyntaxKind.EqualsToken) return nullableAssignment(diagnostic, finding);
                if (current.operatorToken.kind === ts.SyntaxKind.QuestionQuestionEqualsToken) return nestedResult(diagnostic, finding);
            }
        }
        return contextual(diagnostic, finding, 'nested-generic-result');
    }
    function maskedConstructions(program, diagnostics) {
        // A masked construction has no diagnostic on its own contextual
        // expression. Do not let this pass replace a declined diagnostic group.
        const diagnosed = [];
        for (const diagnostic of diagnostics) {
            if (!diagnostic.file || diagnostic.start === undefined) continue;
            for (let current = at(diagnostic.file, diagnostic.start); current; current = current.parent) {
                if (ts.isShorthandPropertyAssignment(current)) break;
                let expression = current;
                if (ts.isPropertyAssignment(current)) expression = current.initializer;
                if (ts.isReturnStatement(current)) expression = current.expression;
                if (ts.isVariableDeclaration(current)) expression = current.initializer;
                if (ts.isBinaryExpression(current) && current.operatorToken.kind === ts.SyntaxKind.EqualsToken) expression = current.right;
                if (!expression || !ts.isExpressionNode(expression)) continue;
                const target = checker.getContextualType(expression);
                if (target && !(target.flags & forbidden) && parts(target).some(type => checker.getPropertiesOfType(type).length)) {
                    diagnosed.push(expression);
                    break;
                }
            }
        }
        let result = false;
        for (const file of program.getSourceFiles()) {
            if (!owned.has(path.resolve(file.fileName))) continue;
            function visit(node) {
                if (ts.isObjectLiteralExpression(node) && !diagnosed.some(expression => expression.getSourceFile() === file && expression.pos <= node.pos && node.end <= expression.end)) {
                    const target = checker.getContextualType(node), source = checker.getTypeAtLocation(node);
                    if (target && !checker.isTypeAssignableTo(source, target)) {
                        const point = file.getLineAndCharacterOfPosition(node.getStart(file));
                        const finding = { file: path.relative(path.resolve(directory, "../.."), file.fileName), line: point.line + 1,
                            column: point.character + 1, code: null, source: node.getText(file),
                            message: 'Existing object construction stores undefined; another incompatible member can mask its optional-property diagnostic',
                            resolution_group: 'late-exposure' };
                        result = resolve(source, target, node, finding) || result;
                    }
                }
                ts.forEachChild(node, visit);
            }
            visit(file);
        }
        return result;
    }
    const directory = path.dirname([...owned][0]);
    return { diagnostic, maskedConstructions, isPublic, isBlocked, deferred };
};
