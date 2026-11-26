"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const value = "PragmaPseudoMap[keyof PragmaPseudoMap] | PragmaPseudoMap[keyof PragmaPseudoMap][]";
const original = name => `forEach(action: <TKey extends keyof PragmaPseudoMap>(value: PragmaPseudoMap[TKey] | PragmaPseudoMap[TKey][], key: TKey, map: ${name}) => void): void;`;
const replacement = `forEach(callbackfn: (value: ${value}, key: string, map: Map<string, ${value}>) => void, thisArg?: any): void;`;
const trivia = "\r\n    ";

function region(file, name) {
    const interfaces = file.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === name);
    assert.equal(interfaces.length, 1, `expected one ${name} interface`);
    const [declaration] = interfaces;
    const members = declaration.members.filter(node => {
        const key = node.name;
        if (!key) return false;
        assert(!ts.isComputedPropertyName(key), `computed member name in ${name}`);
        return (ts.isIdentifier(key) || ts.isStringLiteral(key)) && key.text === "forEach";
    });
    assert(members.length <= 1, `duplicate ${name}.forEach`);
    const [member] = members;
    if (member) {
        assert(ts.isMethodSignature(member), `expected method ${name}.forEach`);
        assert.equal(file.text.slice(member.getFullStart(), member.getStart(file)), trivia, `unexpected trivia on ${name}.forEach`);
    }
    return { declaration, member };
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const filename = path.join(tree, "src/compiler/types.ts");
    const text = fs.readFileSync(filename, "utf8");
    const file = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
    assert.deepEqual(file.parseDiagnostics, []);
    const readonly = region(file, "ReadonlyPragmaMap");
    const map = region(file, "PragmaMap");
    const clauses = map.declaration.heritageClauses;
    assert.equal(clauses?.length, 1, "expected one PragmaMap extends clause");
    assert.equal(clauses[0].getText(file), `extends Map<string, ${value}>, ReadonlyPragmaMap`, "unexpected PragmaMap extends clause");

    if (!readonly.member && map.member?.getText(file) === replacement) return 0;
    assert.equal(readonly.member?.getText(file), original("ReadonlyPragmaMap"), "unexpected ReadonlyPragmaMap.forEach");
    assert.equal(map.member?.getText(file), original("PragmaMap"), "unexpected PragmaMap.forEach");

    const edits = [
        { start: readonly.member.getFullStart(), end: readonly.member.end, text: "" },
        { start: map.member.getStart(file), end: map.member.end, text: replacement },
    ];
    let after = text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) {
        after = after.slice(0, edit.start) + edit.text + after.slice(edit.end);
    }
    fs.writeFileSync(filename, after);
    return edits.length;
}

assert.equal(process.argv.length, 3, "usage: node adapt.cjs <typescript tree>");
const edits = adapt(process.argv[2]);
console.log(JSON.stringify({ files: edits ? 1 : 0, edits }));
