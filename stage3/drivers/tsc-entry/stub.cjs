// Scratch-only body replacement. Refuse declaration/top-level stopping sites.
const fs = require('node:fs');
const ts = require(process.env.SCANNER_TYPESCRIPT);
const [file, lineText, columnText, output] = process.argv.slice(2);
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const position = source.getPositionOfLineAndCharacter(Number(lineText) - 1, Number(columnText) - 1);
let found;
function visit(node) {
    if (position < node.getStart(source) || position >= node.end) return;
    if (ts.isFunctionLike(node) && node.body && position >= node.body.getStart(source)) found = node;
    node.forEachChild(visit);
}
visit(source);
if (!found) {
    console.error('Stopping site has no enclosing function body; cannot replace a body at this site.');
    process.exit(2);
}
const body = found.body;
const replacement = '{ throw new Error("tsc-entry scratch placeholder"); }';
fs.writeFileSync(file, text.slice(0, body.getStart(source)) + replacement + text.slice(body.end));
fs.writeFileSync(output, JSON.stringify({ file, line: Number(lineText), column: Number(columnText),
    function: found.name?.getText(source) || '<anonymous>', start: body.getStart(source), end: body.end,
    original: body.getText(source), replacement }, null, 2) + '\n');
