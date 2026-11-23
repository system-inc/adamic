'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const edits = {
    "builder.ts": [
        {
            "kind": "VariableStatement",
            "before": "const oldSignature = Debug.checkDefined(state.oldSignatures).get(path) || undefined;",
            "after": "const oldSignatures: Map<Path, string | false> | undefined = state.oldSignatures;\n    if (oldSignatures === undefined || oldSignatures === null) Debug.fail();\n    const oldSignature = oldSignatures.get(path) || undefined;"
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
