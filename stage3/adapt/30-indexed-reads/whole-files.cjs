"use strict";
const entries = require("./whole-sites.json");
function planWhole(ts, file, text, check = false) {
    if (!entries.some(entry => entry.file === file)) return { text, edits: 0 };
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const found = [];
    function visit(node) {
        if (ts.isVariableDeclaration(node) && node.name.getText(source) === "assertionCache") found.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    if (file !== "debug.ts" || found.length !== 1) throw new Error("unknown whole-file closure: " + file);
    const variable = found[0], partial = variable.type;
    if (!ts.isTypeReferenceNode(partial) || partial.typeName.getText(source) !== "Partial" || partial.typeArguments.length !== 1) throw new Error("assertion cache Partial drift");
    const record = partial.typeArguments[0];
    if (!ts.isTypeReferenceNode(record) || record.typeName.getText(source) !== "Record" || record.typeArguments.length !== 2 || record.typeArguments[0].getText(source) !== "AssertionKeys") throw new Error("assertion cache Record drift");
    const value = record.typeArguments[1];
    const adapted = ts.isUnionTypeNode(value) && value.types.length === 2 && value.types[1].kind === ts.SyntaxKind.UndefinedKeyword;
    const object = adapted ? value.types[0] : value;
    if (!ts.isTypeLiteralNode(object) || object.members.length !== 2 || object.members.map(member => member.getText(source)).join(" ") !== "level: AssertionLevel; assertion: AnyFunction;") throw new Error("assertion cache entry drift");
    if (adapted) return { text, edits: 0 };
    if (check) throw new Error("truthful cache value union missing: debug.ts:155");
    return { text: text.slice(0, object.end) + " | undefined" + text.slice(object.end), edits: 1 };
}
module.exports = { planWhole };
