// Compare retained helper bodies with the independent pinned upstream source.
const fs = require('fs');
const path = require('path');
const ts = require(process.argv[3]);
const directory = process.argv[4] || __dirname;
let count = 0;
for (const file of fs.readdirSync(directory).filter(file => /^\d\d_.*\.a$/.test(file))) {
    const text = fs.readFileSync(path.join(directory, file), 'utf8');
    const fixture = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const functions = [];
    function walk(node) {
        if (ts.isFunctionDeclaration(node) && node.body) functions.push(node);
        ts.forEachChild(node, walk);
    }
    walk(fixture);
    for (const [, source, line] of text.matchAll(/^\/\/ From TypeScript 6\.0\.3, src\/compiler\/([^:]+):(\d+)$/gm)) {
        const upstream = ts.createSourceFile(source, fs.readFileSync(path.join(process.argv[2], 'src/compiler', source), 'utf8'), ts.ScriptTarget.Latest, true);
        let target;
        function find(node) {
            if (ts.isFunctionDeclaration(node) && node.body && upstream.getLineAndCharacterOfPosition(node.getStart(upstream)).line + 1 === +line) target = node;
            ts.forEachChild(node, find);
        }
        find(upstream);
        if (!target || target.name.text === 'createScanner') continue;
        const retained = functions.find(node => node.name.text === target.name.text);
        if (!retained) throw new Error(`${file}: missing ${target.name.text}`);
        if (retained.body.getText(fixture).replace(/\s+/g, '') !== target.body.getText(upstream).replace(/\s+/g, '')) {
            throw new Error(`${file}: changed ${target.name.text}`);
        }
        count++;
    }
}
console.log(`unchanged upstream helper bodies: ${count}`);
