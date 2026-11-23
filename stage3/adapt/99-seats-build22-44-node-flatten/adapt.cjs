'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const edits = {
    "core.ts": [
        {
            "kind": "ImportSpecifier",
            "before": "type MapLike",
            "after": "type MapLike,\n    type Node"
        },
        {
            "kind": "FunctionDeclaration",
            "before": "export function flatMap<T, U extends {}>(array: readonly T[] | undefined, mapfn: (x: T, i: number) => U | readonly U[] | undefined): readonly U[] {\n    let result: U[] | undefined;\n    if (array !== undefined) {\n        for (let i = 0; i < array.length; i++) {\n            const v = mapfn(array[i]!, i);\n            if (v) {\n                if (isArray(v)) {\n                    result = addRange(result, v);\n                }\n                else {\n                    result = append<U>(result, v);\n                }\n            }\n        }\n    }\n    return result ?? emptyArray;\n}",
            "after": "export function flatMap<T, U extends {}>(array: readonly T[] | undefined, mapfn: (x: T, i: number) => U | readonly U[] | undefined): readonly U[] {\n    let result: U[] | undefined;\n    if (array !== undefined) {\n        for (let i = 0; i < array.length; i++) {\n            const v = mapfn(array[i]!, i);\n            if (v) {\n                if (isArray(v)) {\n                    result = addRange(result, v);\n                }\n                else {\n                    result = append<U>(result, v);\n                }\n            }\n        }\n    }\n    return result ?? emptyArray;\n}\n\n/** @internal */\nexport function flatMapNodeLists<T, U extends Node>(array: readonly T[] | undefined, mapfn: (x: T, i: number) => readonly U[] | undefined): readonly U[] {\n    let result: U[] | undefined;\n    if (array !== undefined) {\n        for (let i = 0; i < array.length; i++) {\n            const mapped: readonly U[] | undefined = mapfn(array[i]!, i);\n            if (mapped !== undefined) {\n                result = addRange(result, mapped);\n            }\n        }\n    }\n    return result ?? emptyArray;\n}"
        }
    ],
    "utilities.ts": [
        {
            "kind": "ImportSpecifier",
            "before": "flatMap",
            "after": "flatMap,\n    flatMapNodeLists"
        },
        {
            "kind": "CallExpression",
            "before": "flatMap(getJSDocTags(node), tag => isNonTypeAliasTemplate(tag) ? tag.typeParameters : undefined)",
            "after": "flatMapNodeLists(getJSDocTags(node), tag => isNonTypeAliasTemplate(tag) ? tag.typeParameters : undefined)"
        }
    ],
    "utilitiesPublic.ts": [
        {
            "kind": "ImportSpecifier",
            "before": "flatMap",
            "after": "flatMap,\n    flatMapNodeLists"
        },
        {
            "kind": "CallExpression",
            "before": "flatMap(jsDoc.tags, tag => isJSDocTemplateTag(tag) ? tag.typeParameters : undefined)",
            "after": "flatMapNodeLists(jsDoc.tags, tag => isJSDocTemplateTag(tag) ? tag.typeParameters : undefined)"
        },
        {
            "kind": "CallExpression",
            "before": "flatMap(node.parent.tags, tag => isJSDocTemplateTag(tag) ? tag.typeParameters : undefined)",
            "after": "flatMapNodeLists(node.parent.tags, tag => isJSDocTemplateTag(tag) ? tag.typeParameters : undefined)"
        }
    ]
};
const root = process.argv[2];
for (const [name, changes] of Object.entries(edits)) {
    const filename = path.join(root, 'src/compiler', name);
    const original = fs.readFileSync(filename, 'utf8');
    const newline = original.includes('\r\n') ? '\r\n' : '\n';
    const source = ts.createSourceFile(filename, original, ts.ScriptTarget.Latest, true);
    const patches = changes.map(({ kind, before, after }) => {
        const expected = before.replaceAll('\n', newline);
        const matches = [];
        function visit(node) {
            if (node.kind === ts.SyntaxKind[kind] && node.getText(source) === expected) matches.push(node);
            ts.forEachChild(node, visit);
        }
        visit(source);
        assert.equal(matches.length, 1, filename + ': expected one ' + kind + ' matching ' + before);
        const [match] = matches;
        return { start: match.getStart(source), end: match.end, after: after.replaceAll('\n', newline) };
    });
    patches.sort((first, second) => second.start - first.start);
    let text = original;
    let boundary = original.length;
    for (const { start, end, after } of patches) {
        assert(end <= boundary, filename + ': overlapping edits');
        text = text.slice(0, start) + after + text.slice(end);
        boundary = start;
    }
    fs.writeFileSync(filename, text);
}
