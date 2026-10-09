"use strict";
// Edit only a private discovery copy, recording the removed body and replacement.
const fs = require("node:fs");
const ts = require(process.env.SCANNER_TYPESCRIPT);
if (ts.version !== "6.0.3") throw Error("requires stock TypeScript 6.0.3");
const [file, line, column, forceReturn] = process.argv.slice(2);
let text = fs.readFileSync(file, "utf8");
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const position = source.getPositionOfLineAndCharacter(+line - 1, +column - 1);
const candidates = [];
function visit(node) {
    if (node.getStart(source) <= position && position < node.end && ts.isFunctionLike(node) && node.body) candidates.push(node);
    ts.forEachChild(node, visit);
}
visit(source);
candidates.sort((a, b) => Number(ts.isFunctionDeclaration(b)) - Number(ts.isFunctionDeclaration(a)) || (a.end - a.pos) - (b.end - b.pos));
const owner = candidates[0];
if (!owner) throw Error("no enclosing function");
let annotation = "";
if (!owner.type && ts.isFunctionDeclaration(owner)) {
    const program = ts.createProgram([file], {strict: true, allowImportingTsExtensions: true, target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext});
    const checker = program.getTypeChecker();
    let declaration;
    function find(node) {
        if (node.pos === owner.pos && node.kind === owner.kind) declaration = node;
        ts.forEachChild(node, find);
    }
    find(program.getSourceFile(file));
    const signature = checker.getSignatureFromDeclaration(declaration);
    annotation = ": " + checker.typeToString(checker.getReturnTypeOfSignature(signature), declaration, ts.TypeFormatFlags.NoTruncation) + " ";
}
if (forceReturn && !owner.type) annotation = ": " + forceReturn + " ";
const name = owner.name?.getText(source) || "<anonymous>";
const replacement = annotation + '{ throw new Error("discovery placeholder: ' + name + '"); }';
text = text.slice(0, owner.body.getStart(source)) + replacement + text.slice(owner.body.end);
if (forceReturn && owner.type) text = text.slice(0, owner.type.getStart(source)) + forceReturn + text.slice(owner.type.end);
fs.writeFileSync(file, text);
console.log(JSON.stringify({file, line: +line, column: +column, name, removed: owner.body.getText(source), replacement}));
