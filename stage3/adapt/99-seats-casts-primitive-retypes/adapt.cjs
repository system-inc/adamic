#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

// Exact owner snapshots after the earlier stage3 adapters, including original line endings.
const pins = [
    {
        "name": "getEntityNameFromTypeNode",
        "before": "export function getEntityNameFromTypeNode(node: TypeNode): EntityNameOrEntityNameExpression | undefined {\r\n    switch (node.kind) {\r\n        case SyntaxKind.TypeReference:\r\n            return (node as TypeReferenceNode).typeName;\r\n\r\n        case SyntaxKind.ExpressionWithTypeArguments:\r\n            return isEntityNameExpression((node as ExpressionWithTypeArguments).expression)\r\n                ? (node as ExpressionWithTypeArguments).expression as EntityNameExpression\r\n                : undefined;\r\n\r\n        // TODO(rbuckton): These aren't valid TypeNodes, but we treat them as such because of `isPartOfTypeNode`, which returns `true` for things that aren't `TypeNode`s.\r\n        case SyntaxKind.Identifier as TypeNodeSyntaxKind:\r\n        case SyntaxKind.QualifiedName as TypeNodeSyntaxKind:\r\n            return (node as Node as EntityName);\r\n    }\r\n\r\n    return undefined;\r\n}",
        "after": "export function getEntityNameFromTypeNode(node: TypeNode | EntityName): EntityNameOrEntityNameExpression | undefined {\r\n    switch (node.kind) {\r\n        case SyntaxKind.TypeReference:\r\n            return (node as TypeReferenceNode).typeName;\r\n\r\n        case SyntaxKind.ExpressionWithTypeArguments:\r\n            return isEntityNameExpression((node as ExpressionWithTypeArguments).expression)\r\n                ? (node as ExpressionWithTypeArguments).expression as EntityNameExpression\r\n                : undefined;\r\n\r\n        // TODO(rbuckton): These aren't valid TypeNodes, but we treat them as such because of `isPartOfTypeNode`, which returns `true` for things that aren't `TypeNode`s.\r\n        case SyntaxKind.Identifier:\r\n        case SyntaxKind.QualifiedName:\r\n            return node;\r\n    }\r\n\r\n    return undefined;\r\n}"
    }
];

function refuse(reason) {
    throw new Error(`seats primitive retypes: ${reason}`);
}

if (process.argv.length !== 3) refuse('usage: node adapt.cjs <typescript tree>');
if (ts.version !== '6.0.3') refuse(`expected stock TypeScript 6.0.3, got ${ts.version}`);
const file = path.join(process.argv[2], 'src/compiler/utilities.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
if (source.parseDiagnostics.length) refuse('utilities.ts must parse');

const matches = pins.map(pin => {
    const owners = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === pin.name);
    if (owners.length !== 1 || !owners[0].body) refuse(`expected one implementation of ${pin.name}`);
    const [owner] = owners;
    if (!ts.getJSDocTags(owner).some(tag => tag.tagName.text === 'internal')) {
        refuse(`missing @internal JSDoc tag for ${pin.name}`);
    }
    const actual = owner.getText(source);
    if (actual !== pin.before && actual !== pin.after) refuse(`unexpected exact text for ${pin.name}`);
    return {start: owner.getStart(source), end: owner.end, after: pin.after, adapted: actual === pin.after};
});

const countAdapted = matches.filter(match => match.adapted).length;
if (countAdapted && countAdapted !== matches.length) refuse('partly adapted utilities.ts');
if (!countAdapted) {
    let result = text;
    for (const {start, end, after} of matches.sort((a, b) => b.start - a.start)) {
        result = result.slice(0, start) + after + result.slice(end);
    }
    if (fs.readFileSync(file, 'utf8') !== text) refuse('utilities.ts changed during planning');
    fs.writeFileSync(file, result);
}
console.log(JSON.stringify({files: countAdapted ? 0 : 1, edits: countAdapted ? 0 : matches.length}));
