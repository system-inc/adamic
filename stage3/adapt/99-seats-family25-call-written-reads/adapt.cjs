#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');

// Keep both property reads when tsc tests and then consumes the value.
// These edits change only the flow reference that the checker sees.
const edits = {
    'checker.ts': [
        ['tracked symbols', 'trackedSymbols: context.trackedSymbols,', 'trackedSymbols: (context as typeof context).trackedSymbols,'],
        ['contextual variable type', 'if (!links.type && !isParameterOfContextSensitiveSignature(symbol))', 'if (!(links as typeof links).type && !isParameterOfContextSensitiveSignature(symbol))'],
        ['enum member type', 'const enumType = getDeclaredTypeOfEnum(getParentOfSymbol(symbol)!);\n            if (!links.declaredType) {\n                links.declaredType = enumType;\n            }',
         'const enumType = getDeclaredTypeOfEnum(getParentOfSymbol(symbol)!);\n            const currentLinks = links;\n            if (!currentLinks.declaredType) {\n                currentLinks.declaredType = enumType;\n            }\n            return currentLinks.declaredType;'],
        ['assignment position', 'return isSymbolAssigned(symbol) && symbol.lastAssignmentPos !== undefined && symbol.lastAssignmentPos < 0;',
         'if (!isSymbolAssigned(symbol)) return false;\n        const currentSymbol = symbol;\n        return currentSymbol.lastAssignmentPos !== undefined && currentSymbol.lastAssignmentPos < 0;'],
        ['contextual return type', 'if (!signature.resolvedReturnType) {\n                        signature.resolvedReturnType = returnType;',
         'if (!(signature as typeof signature).resolvedReturnType) {\n                        signature.resolvedReturnType = returnType;'],
        ['resolved symbol', 'if (!links.resolvedSymbol) {\n                        links.resolvedSymbol = getApplicableIndexSymbol(',
         'if (!(links as typeof links).resolvedSymbol) {\n                        links.resolvedSymbol = getApplicableIndexSymbol('],
    ],
    'utilitiesPublic.ts': [
        ['illegal type parameters', 'if (canHaveIllegalTypeParameters(node) && node.typeParameters) {\n        return node.typeParameters;\n    }',
         'if (canHaveIllegalTypeParameters(node)) {\n        const currentNode = node;\n        if (currentNode.typeParameters) return currentNode.typeParameters;\n    }'],
    ],
    'transformers/declarations.ts': [
        ['side-effect import clause', 'decl.modifiers,\n                decl.importClause,\n                rewriteModuleSpecifier(decl, decl.moduleSpecifier),',
         'decl.modifiers,\n                (decl as typeof decl).importClause,\n                rewriteModuleSpecifier(decl, decl.moduleSpecifier),'],
    ],
};

function plan(tree) {
    return Object.entries(edits).map(([file, changes]) => {
        const fileName = path.join(tree, 'src/compiler', file);
        const original = fs.readFileSync(fileName, 'utf8');
        const newline = original.includes('\r\n') ? '\r\n' : '\n';
        let result = original;
        const outcomes = [];

        for (const [label, before, after] of changes) {
            const from = before.replaceAll('\n', newline);
            const to = after.replaceAll('\n', newline);
            const beforeCount = result.split(from).length - 1;
            const afterCount = result.split(to).length - 1;
            if (beforeCount === 1 && afterCount === 0) {
                result = result.replace(from, to);
                outcomes.push([label, 'edited']);
            } else if (beforeCount === 0 && afterCount === 1) {
                outcomes.push([label, 'already adapted']);
            } else {
                throw new Error(`${file}: unexpected anchors for ${label} (${beforeCount} before, ${afterCount} after)`);
            }
        }

        return { fileName, original, result, outcomes };
    });
}

if (require.main === module) {
    if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
    const plans = plan(path.resolve(process.argv[2]));
    for (const { fileName, original } of plans) {
        if (fs.readFileSync(fileName, 'utf8') !== original) throw new Error(`source changed during planning: ${fileName}`);
    }
    for (const { fileName, original, result, outcomes } of plans) {
        if (result !== original) fs.writeFileSync(fileName, result);
        for (const [label, status] of outcomes) console.log(`${label}: ${status}`);
    }
}

module.exports = { edits };
