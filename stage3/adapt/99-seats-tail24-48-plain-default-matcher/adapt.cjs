"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const { nodes } = require("../47-host-errors/adapt.cjs");

const file = "src/compiler/utilities.ts";
const binding = "{ singleAsteriskRegexFragment, doubleAsteriskRegexFragment, replaceWildcardCharacter }";
const initializer = ": WildcardMatcher = wildcardMatchers[usage]";
const originalParameter = binding + initializer;
const plainParameter = "matcher" + initializer;
const destructure = "const " + binding + " = matcher;";
const firstStatement = 'let subpattern = "";';

function signature(parameter, newline) {
    return [
        "export function getSubPatternFromSpec(",
        "    spec: string,",
        "    basePath: string,",
        '    usage: "files" | "directories" | "exclude",',
        "    " + parameter + ",",
        "): string | undefined ",
    ].join(newline);
}

function findOwner(text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, "utilities.ts has a syntax error");
    const declarations = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === "getSubPatternFromSpec");
    assert.equal(declarations.length, 1, "getSubPatternFromSpec owner drift");
    const [owner] = declarations;
    assert(owner.body && ts.isInternalDeclaration(owner), "expected the internal implementation");
    return { source, owner };
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, "utf8");
    const { source, owner } = findOwner(text);
    assert.equal(owner.parameters.length, 4, "expected the matcher to be the last parameter");
    const parameter = owner.parameters[3];
    const parameterText = parameter.getText(source);
    const alreadyApplied = parameterText === plainParameter;
    assert(alreadyApplied || parameterText === originalParameter, "unreviewed matcher parameter");
    const bodyStart = owner.body.getStart(source);
    const newline = text.includes("\r\n") ? "\r\n" : "\n";
    assert.equal(text.slice(owner.getStart(source), bodyStart), signature(parameterText, newline), "signature drift");

    const insertion = newline + "    " + destructure;
    const firstOriginal = owner.body.statements[alreadyApplied ? 1 : 0];
    assert.equal(firstOriginal?.getText(source), firstStatement, "first statement drift");
    const expectedOpening = "{" + (alreadyApplied ? insertion : "") + newline + "    ";
    assert.equal(text.slice(bodyStart, firstOriginal.getStart(source)), expectedOpening, "body opening drift");

    const names = new Set(["matcher", "arguments", "eval"]);
    const uses = nodes(ts, owner.body, node => ts.isIdentifier(node) && names.has(node.text));
    assert.deepEqual(uses.map(node => node.text), alreadyApplied ? ["matcher"] : [], "unreviewed name in function body");
    if (alreadyApplied) return 0;

    const start = parameter.getStart(source);
    const after = text.slice(0, start) + plainParameter +
        text.slice(parameter.end, bodyStart + 1) + insertion + text.slice(bodyStart + 1);
    fs.writeFileSync(fileName, after);
    return 1;
}

module.exports = { adapt, file, findOwner };
if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
    console.log(JSON.stringify({ files: adapt(process.argv[2]), javascriptChanging: true }));
}
