'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const plans = {
    "src/compiler/transformers/module/system.ts": [
        {
            "owner": "transformSystemModule",
            "edits": [
                {
                    "kind": "VariableDeclaration",
                    "before": "contextObjectMap: Identifier[] = []",
                    "after": "contextObjectMap = new Map<number, Identifier>()",
                    "sequence": false
                },
                {
                    "kind": "ExpressionStatement",
                    "before": "contextObject = contextObjectMap[id] = factory.createUniqueName(\"context\");",
                    "after": "const collectedContextObject = factory.createUniqueName(\"context\");\n        contextObjectMap.set(id, collectedContextObject);\n        contextObject = collectedContextObject;",
                    "sequence": true,
                    "afterKind": "VariableStatement"
                },
                {
                    "kind": "ExpressionStatement",
                    "before": "contextObject = contextObjectMap[id];",
                    "after": "contextObject = contextObjectMap.get(id);",
                    "sequence": false
                }
            ],
            "names": [
                "contextObjectMap",
                3
            ]
        }
    ]
};

function planFile(root, relative, groups) {
    const file = path.join(root, relative);
    const text = fs.readFileSync(file, 'utf8');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, relative);
    const newline = text.includes('\r\n') ? '\r\n' : '\n';
    const edits = [];
    const snapshots = [];
    for (const group of groups) {
        const owners = [];
        function find(node) {
            if ((ts.isFunctionDeclaration(node) || ts.isInterfaceDeclaration(node)) &&
                node.name?.text === group.owner) owners.push(node);
            ts.forEachChild(node, find);
        }
        find(source);
        assert.equal(owners.length, 1, group.owner);
        const owner = owners[0];
        if (group.names) {
            let count = 0;
            function countNames(node) {
                if (ts.isIdentifier(node) && node.text === group.names[0]) count++;
                ts.forEachChild(node, countNames);
            }
            countNames(owner);
            assert.equal(count, group.names[1], 'unreviewed table reference');
        }
        for (const edit of group.edits) {
            const candidates = [];
            function scan(node) {
                if ((node.kind === ts.SyntaxKind[edit.kind] || node.kind === ts.SyntaxKind[edit.afterKind])) candidates.push(node);
                ts.forEachChild(node, scan);
            }
            scan(owner);
            const before = edit.before.replaceAll('\n', newline);
            const after = edit.after.replaceAll('\n', newline);
            const positions = [];
            for (const [spelling, applied] of [[before, false], [after, true]]) {
                for (const node of candidates) {
                    const start = node.getStart(source);
                    if (!text.startsWith(spelling, start)) continue;
                    const end = start + spelling.length;
                    if (node.end > end) continue;
                    if (node.end !== end && !edit.sequence) continue;
                    if (edit.sequence) {
                        const siblings = Array.from(node.parent.statements ?? []);
                        const boundary = siblings.findIndex(sibling => sibling.end === end);
                        if (boundary < siblings.indexOf(node)) continue;
                    }
                    positions.push({start, end, applied});
                }
            }
            assert.equal(positions.length, 1, `drifted or duplicate edit in ${group.owner}: ${edit.before}`);
            const match = positions[0];
            snapshots.push(text.slice(match.start, match.end));
            if (!match.applied) edits.push({...match, after});
        }
    }
    let result = text;
    for (const edit of edits.sort((a,b) => b.start-a.start)) {
        result = result.slice(0, edit.start) + edit.after + result.slice(edit.end);
    }
    return {file, text, result, snapshots};
}

function prepare(root) {
    return Object.entries(plans).map(([file, groups]) => planFile(root, file, groups));
}

if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs ADAPTED_TREE');
    const changes = prepare(process.argv[2]);
    let files = 0;
    for (const {file, text, result} of changes) {
        if (text !== result) {
            fs.writeFileSync(file, result);
            files++;
        }
    }
    console.log(JSON.stringify({files}));
}
module.exports = {prepare, files: Object.keys(plans)};
