"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node test.cjs <applied-tree>");
const tree = process.argv[2];
const ref = "refs/stage3/99-seats-scout24-syntactic-result-member";
const relative = "src/compiler/expressionToTypeNode.ts";
const adapter = path.join(__dirname, "adapt.cjs");
function snapshot(state) {
    return execFileSync("git", ["-C", tree, "show", `${ref}-${state}:${relative}`]);
}
const before = snapshot("before");
const after = snapshot("after");
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "syntactic-result-test-"));
const file = path.join(scratch, relative);

function invoke() {
    const result = spawnSync(process.execPath, [adapter, scratch], { encoding: "utf8" });
    if (result.error) throw result.error;
    assert.equal(result.signal, null);
    return result;
}

function reject(label, bytes, diagnostic) {
    fs.writeFileSync(file, bytes);
    const mtime = fs.statSync(file, { bigint: true }).mtimeNs;
    const result = invoke();
    assert.notEqual(result.status, 0, label + " accepted");
    assert.match(result.stderr, diagnostic, label + " failed at the wrong guard");
    assert(fs.readFileSync(file).equals(Buffer.from(bytes)), label + " wrote source bytes");
    assert.equal(fs.statSync(file, { bigint: true }).mtimeNs, mtime, label + " changed mtime");
    console.log("PASS", label);
}

function checkDrifts(bytes, state) {
    const text = bytes.toString("utf8");
    const source = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    function exact(kind, expected) {
        const matches = [];
        function visit(node) {
            if (kind(node) && node.getText(source) === expected) matches.push(node);
            ts.forEachChild(node, visit);
        }
        visit(source);
        assert.equal(matches.length, 1, expected);
        return matches[0];
    }
    function substitute(node, edited) {
        return text.slice(0, node.getStart(source)) + edited + text.slice(node.end);
    }

    const callText = "syntacticResult(factory.createKeywordTypeNode(SyntaxKind.StringKeyword))";
    const call = exact(ts.isCallExpression, callText);
    reject(state + " changes producer argument",
        substitute(call, callText.replace("SyntaxKind.StringKeyword", "SyntaxKind.NumberKeyword")),
        /reviewed syntacticResult statement/);
    reject(state + " Object.assign wraps producer",
        substitute(call, `Object.assign(${callText}, { reportFallback: false })`),
        /reviewed syntacticResult statement/);
    reject(state + " spread wraps producer",
        substitute(call, `{ ...${callText}, reportFallback: false }`),
        /reviewed syntacticResult statement/);

    const guard = "result.type !== undefined ? result.type : inferExpressionType(expr, context, result.reportFallback)";
    const firstRead = exact(ts.isConditionalExpression, guard);
    assert(ts.isReturnStatement(firstRead.parent));
    const statement = firstRead.parent;
    reject(state + " writes result flag after first guard",
        substitute(statement, statement.getText(source) + "\r\n        result.reportFallback = false;"),
        /reportFallback access.*has no reviewed enclosure|reviewed reportFallback guard/);
    const branchReturn = call.parent;
    assert(ts.isReturnStatement(branchReturn) && ts.isBlock(branchReturn.parent));
    reject(state + " adds nineteenth producer",
        substitute(branchReturn, "return syntacticResult(undefined, true);\r\n                    " + branchReturn.getText(source)),
        /reviewed syntacticResult statement/);
    reject(state + " moves read outside guard",
        substitute(statement, "const fallback = result.reportFallback;\r\n        return " +
            guard.replace("result.reportFallback", "fallback") + ";"),
        /reportFallback access.*has no reviewed enclosure|reviewed reportFallback guard/);

    const assignment = exact(ts.isExpressionStatement,
        "resultType = syntacticResult(serializeTypeAnnotationOfDeclaration(declaredType, context, node, symbol));");
    const start = assignment.getStart(source);
    const { line } = source.getLineAndCharacterOfPosition(start);
    const indent = text.slice(source.getLineStarts()[line], start);
    assert.match(indent, /^ +$/);
    reject(state + " writes type of a result binding",
        substitute(assignment, assignment.getText(source) + "\r\n" + indent + "alreadyReported.type = resultType.type;"),
        /reviewed result reference alreadyReported at line [0-9]+: alreadyReported\.type/);

    const originalDeclaration = "result = typeFromExpression(expr, context, /*isConstContext*/ false, addUndefined, preserveLiterals)";
    const original = exact(ts.isVariableStatement, "const " + originalDeclaration + ";");
    const declaration = original.declarationList.declarations;
    assert.equal(declaration.length, 1);
    assert.equal(declaration[0].getText(source), originalDeclaration);
    const originalStart = original.getStart(source);
    const originalLine = source.getLineAndCharacterOfPosition(originalStart).line;
    const indentation = text.slice(source.getLineStarts()[originalLine], originalStart);
    const insertAfter = (...lines) => substitute(original, original.getText(source) +
        lines.map(line => "\r\n" + indentation + line).join(""));
    reject(state + " declares unannotated result binding",
        insertAfter("const extra = typeFromExpression(expr, context);",
            "extra.type = factory.createKeywordTypeNode(SyntaxKind.AnyKeyword);"),
        /reviewed typed result binding extra at line [0-9]+/);
    // Isolate the assignment: the checker does recognize a subsequent property receiver.
    reject(state + " assigns unannotated let from result call",
        insertAfter("let extra;", "extra = typeFromExpression(expr, context);"),
        /reviewed result call statement typeFromExpression at line [0-9]+/);
}

try {
    fs.mkdirSync(path.dirname(file), { recursive: true });
    fs.writeFileSync(file, before);
    const first = invoke();
    assert.equal(first.status, 0, first.stderr);
    assert(fs.readFileSync(file).equals(after), "first application differs from its after snapshot");
    console.log("PASS first application matches after snapshot");
    const mtime = fs.statSync(file, { bigint: true }).mtimeNs;
    const replay = invoke();
    assert.equal(replay.status, 0, replay.stderr);
    assert(fs.readFileSync(file).equals(after), "replay changed bytes");
    assert.equal(fs.statSync(file, { bigint: true }).mtimeNs, mtime, "replay changed mtime");
    console.log("PASS replay preserves bytes and mtime");
    checkDrifts(before, "before");
    checkDrifts(after, "after");
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
