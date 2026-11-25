"use strict";

const assert = require("node:assert/strict");
const {execFileSync} = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const ts = require("typescript");
const {sites, utilities, findUtility, readCalls, callKey, parse, computedStatement, checkComputedBlock, rewrite, apply} = require("./adapt.cjs");

assert.equal(process.argv.length, 3, "usage: node test.cjs <applied tree>");
const tree = process.argv[2];
const ref = "refs/stage3/99-seats-build22-91-chain-array-parameter";

function snapshot(side, file) {
    const text = execFileSync("git", ["-C", tree, "show", `${ref}-${side}:${file}`],
        {encoding: "utf8", maxBuffer: 8 * 1024 * 1024});
    return parse(file, text);
}

function transpile(text) {
    return ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2024}}).outputText;
}

function compileUtility(source) {
    const code = transpile(findUtility(source).getText(source).replace(/^export /, ""));
    return new Function("getLocaleSpecificMessage", "some", "formatStringFromArgs",
        `${code}\nreturn chainDiagnosticMessages;`)(
        message => message.message, items => items.length !== 0,
        (text, args) => text.replace("{0}", String(args[0])),
    );
}

function expectedArgument(source, row, entries) {
    const {call, operation, text} = row;
    if (operation === "keep") return text;
    if (operation === "optional") return `${text} === undefined ? undefined : [${text}]`;
    if (operation === "computed") {
        return "elaboratedErrorInfo === undefined ? undefined : [elaboratedErrorInfo]";
    }
    if (operation === "branch") {
        const last = call.arguments[0].whenFalse;
        assert.equal(last.getText(source), "fileIncludeReasonDetails");
        return text.slice(0, last.getStart(source) - call.arguments[0].getStart(source)) +
            "fileIncludeReasonDetails === undefined ? undefined : [fileIncludeReasonDetails]";
    }
    assert.equal(operation, "wrap");
    const argument = call.arguments[0];
    const start = argument.getStart(source);
    const nested = entries.filter(entry =>
        entry.call !== call && entry.operation !== "keep" &&
        entry.call.getStart(source) >= start && entry.call.end <= argument.end);
    const brackets = nested.flatMap(entry => {
        assert.equal(entry.operation, "wrap", "nested operation");
        return [[entry.call.arguments[0].getStart(source) - start, "["],
            [entry.call.arguments[0].end - start, "]"]];
    });
    let expected = text;
    for (const [at, bracket] of brackets.sort((a, b) => b[0] - a[0])) {
        expected = expected.slice(0, at) + bracket + expected.slice(at);
    }
    return `[${expected}]`;
}

const beforeUtility = compileUtility(snapshot("before", utilities));
const afterUtility = compileUtility(snapshot("after", utilities));
const message = {message: "head {0}", category: 1, code: 100};
const single = {messageText: "single", category: 2, code: 101, next: undefined};
const other = {messageText: "other", category: 3, code: 102, next: undefined};
const existing = [single, other];

function sameChain(beforeDetails, afterDetails) {
    const result = afterUtility(afterDetails, message, "argument");
    assert.deepStrictEqual(result, beforeUtility(beforeDetails, message, "argument"));
    return result;
}

sameChain(undefined, undefined);
sameChain(single, [single]);
assert.strictEqual(sameChain(existing, existing).next, existing, "incoming array identity");

const forms = new Map();
let computed;
let changed = 0;
let unchanged = 0;
for (const [file, rows] of Object.entries(sites)) {
    const before = snapshot("before", file);
    const after = snapshot("after", file);
    const oldCalls = readCalls(before);
    const newCalls = readCalls(after);
    assert.equal(oldCalls.size, rows.length, file);
    assert.equal(newCalls.size, rows.length, file);
    const entries = rows.map(([owner, ordinal, text, operation]) => {
        const key = callKey(owner, ordinal);
        const call = oldCalls.get(key);
        assert(call, `missing before: ${file}: ${key}`);
        assert.equal(call.arguments[0].getText(before), text, key);
        return {key, call, text, operation};
    });
    for (const entry of entries) {
        const {key, call, operation} = entry;
        const location = `${file}: ${key}`;
        const afterCall = newCalls.get(key);
        assert(afterCall, `missing after: ${location}`);
        if (operation === "keep") {
            assert.equal(afterCall.getText(after), call.getText(before), location);
            unchanged++;
        } else {
            const expected = expectedArgument(before, entry, entries);
            assert.equal(afterCall.arguments[0].getText(after), expected, location);
            assert.equal(afterCall.expression.getText(after), call.expression.getText(before), location);
            assert.equal(afterCall.arguments.length, call.arguments.length, location);
            assert.deepStrictEqual(afterCall.arguments.slice(1).map(argument => argument.getText(after)),
                call.arguments.slice(1).map(argument => argument.getText(before)), location);
            const oldArgument = call.arguments[0];
            const newArgument = afterCall.arguments[0];
            assert.equal(after.text.slice(afterCall.getStart(after), newArgument.getStart(after)),
                before.text.slice(call.getStart(before), oldArgument.getStart(before)), location);
            assert.equal(after.text.slice(newArgument.end, afterCall.end),
                before.text.slice(oldArgument.end, call.end), location);
            changed++;
        }
        if (operation === "computed") {
            assert.equal(computed, undefined);
            const oldStatement = computedStatement(call, before);
            const newStatement = computedStatement(afterCall, after);
            const statements = newStatement.parent.statements;
            const declaration = statements[statements.indexOf(newStatement) - 1];
            assert(ts.isVariableStatement(declaration), location);
            assert.equal(declaration.declarationList.declarations.length, 1, location);
            const temporary = declaration.declarationList.declarations[0];
            assert.equal(temporary.name.getText(after), "elaboratedErrorInfo", location);
            assert.equal(temporary.initializer.getText(after), entry.text, location);
            computed = [oldStatement.getText(before), `${declaration.getText(after)}\n${newStatement.getText(after)}`];
        }
        if (operation === "branch") forms.set("branch", afterCall.arguments[0].getText(after));
        if (operation === "optional" && entry.text === "errorInfo") {
            forms.set("optional", afterCall.arguments[0].getText(after));
        }
    }
}

const optional = new Function("errorInfo", `return ${forms.get("optional")};`);
sameChain(single, optional(single));
sameChain(undefined, optional(undefined));
const branch = new Function("redirectInfo", "fileIncludeReasonDetails", `return ${forms.get("branch")};`);
sameChain(undefined, branch(undefined, undefined));
sameChain(single, branch(undefined, single));
assert.strictEqual(sameChain(existing, branch(existing, undefined)).next, existing, "redirect array identity");
sameChain([single, ...existing], branch(existing, single));

assert(computed, "missing computed argument test");
function runComputed(text, utility, produced) {
    const code = transpile(`let errorInfo = input;\n${text}\nreturn errorInfo;`);
    const run = new Function("input", "containingType", "diagnostic", "missingProperty", "container",
        "elaborateNeverIntersection", "chainDiagnosticMessages", code);
    const containingType = {};
    let count = 0;
    const result = run(single, containingType, message, "argument", "container", (info, type) => {
        assert.strictEqual(info, single);
        assert.strictEqual(type, containingType);
        count++;
        return produced;
    }, utility);
    assert.equal(count, 1, "computed chain evaluates once");
    return result;
}
for (const produced of [undefined, single]) {
    assert.deepStrictEqual(runComputed(computed[1], afterUtility, produced),
        runComputed(computed[0], beforeUtility, produced));
}
const programFile = "src/compiler/program.ts";
const programBefore = snapshot("before", programFile).text;
assert.equal(rewrite(programFile, programBefore, sites[programFile]), snapshot("after", programFile).text);

const messageName = "Diagnostics.Visit_https_Colon_Slash_Slashaka_ms_Slashts6_for_migration_information";
for (const [body, kind] of [
    [`const alias: any = chainDiagnosticMessages; alias(undefined, ${messageName});`, "Identifier"],
    [`(chainDiagnosticMessages)(undefined, ${messageName});`, "Identifier"],
    [`ts["chainDiagnosticMessages"](undefined, ${messageName});`, "StringLiteral"],
]) {
    const witness = `${programBefore}\nfunction unlistedChainUse() { ${body} }\n`;
    assert.throws(() => rewrite(programFile, witness, sites[programFile]), error =>
        error.message.startsWith(programFile + ":") && error.message.endsWith(`(${kind})`));
}
assert.throws(() => rewrite(programFile, programBefore + '\nconst embedded = "prefix chainDiagnosticMessages";\n',
    sites[programFile]), /unclassified textual chainDiagnosticMessages/);
const comment = "\n// chainDiagnosticMessages\n/** {@link chainDiagnosticMessages} */\nconst commentControl = 0;\n";
assert.equal(rewrite(programFile, programBefore + comment, sites[programFile]),
    snapshot("after", programFile).text + comment);

const unbraced = parse("unbraced.ts",
    "function witness() { if (x) errorInfo = chainDiagnosticMessages(undefined, diagnostic); }");
const [[key, unbracedCall]] = readCalls(unbraced);
assert.throws(() => checkComputedBlock(computedStatement(unbracedCall, unbraced), key),
    /witness #1, call 1: computed assignment must be directly inside a block/);

const closure = `const elaboratedErrorInfo = "outer";
function witness() {
    let errorInfo = "prior";
    const readOuter = () => elaboratedErrorInfo;
    errorInfo = chainDiagnosticMessages(elaborateNeverIntersection(errorInfo, containingType), diagnostic);
    return readOuter();
}
`;
const closureRows = [["witness #1", 1, "elaborateNeverIntersection(errorInfo, containingType)", "computed"]];
assert.throws(() => rewrite("closure.ts", closure, closureRows),
    /witness #1, call 1: elaboratedErrorInfo already occurs at closure\.ts:1/);

const newline = programBefore.includes("\r\n") ? "\r\n" : "\n";
const importLine = "    chainDiagnosticMessages," + newline;
const sibling = "    harmlessSibling," + newline;
const programAfter = snapshot("after", programFile).text;
assert(programBefore.includes(importLine) && programAfter.includes(importLine), programFile);
assert.equal(rewrite(programFile, programBefore.replace(importLine, sibling + importLine), sites[programFile]),
    programAfter.replace(importLine, sibling + importLine));
const withoutImport = programBefore.replace(importLine, "");
for (const altered of [
    withoutImport + newline + 'import { chainDiagnosticMessages } from "./utilities.js";' + newline,
    programBefore.replace(importLine, "    chainDiagnosticMessages as chainMessages," + newline),
    programBefore.replace(importLine, "    type chainDiagnosticMessages," + newline),
    programBefore + newline + 'import { chainDiagnosticMessages } from "./_namespaces/ts.js";' + newline,
]) {
    assert.throws(() => rewrite(programFile, altered, sites[programFile]),
        {message: /^src\/compiler\/program\.ts:\d+:\d+: unlisted chainDiagnosticMessages reference \(Identifier\)$/});
}
assert.throws(() => rewrite(programFile, withoutImport, sites[programFile]),
    /src\/compiler\/program\.ts: missing pinned chainDiagnosticMessages reference/);

const escapedFile = "src/escapedReference.ts";
const escapedImport = String.raw`import { chainDiagnostic\u004dessages as fn } from "./compiler/utilities.js";
const invoke: any = fn;
export const chain = invoke({ messageText: "child", category: 1, code: 7 }, { message: "head", category: 1, code: 8 });
`;
assert(!escapedImport.includes("chainDiagnosticMessages"));
for (const withWitness of [true, false]) {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "diagnostic-chain-"));
    try {
        const files = new Map(Object.keys(sites).map(file => [file, snapshot("before", file).text]));
        if (withWitness) files.set(escapedFile, escapedImport);
        for (const [file, contents] of files) {
            const target = path.join(scratch, file);
            fs.mkdirSync(path.dirname(target), {recursive: true});
            fs.writeFileSync(target, contents);
        }
        if (withWitness) {
            assert.throws(() => apply(scratch),
                /src\/escapedReference\.ts:\d+:\d+: unlisted chainDiagnosticMessages reference \(Identifier\)/);
        } else {
            apply(scratch);
        }
        for (const [file, contents] of files) {
            const expected = withWitness ? contents : snapshot("after", file).text;
            assert.deepStrictEqual(fs.readFileSync(path.join(scratch, file)), Buffer.from(expected), file);
        }
    } finally {
        fs.rmSync(scratch, {recursive: true});
    }
}

const escapedFresh = String.raw`function witness() {
    let errorInfo = "prior";
    const readOuter = () => elaborated\u0045rrorInfo;
    errorInfo = chainDiagnosticMessages(elaborateNeverIntersection(errorInfo, containingType), diagnostic);
    return readOuter();
}
`;
assert(!escapedFresh.includes("elaboratedErrorInfo"));
assert.throws(() => rewrite("escapedFresh.ts", escapedFresh, closureRows),
    {message: "escapedFresh.ts: witness #1, call 1: elaboratedErrorInfo already occurs at escapedFresh.ts:3"});

const stringFresh = `function witness() {
    let errorInfo = "prior";
    const readOuter = () => eval("elaboratedErrorInfo");
    errorInfo = chainDiagnosticMessages(elaborateNeverIntersection(errorInfo, containingType), diagnostic);
    return readOuter();
}
`;
function hasFreshNameIdentifier(node) {
    return (ts.isIdentifier(node) && node.text === "elaboratedErrorInfo") ||
        ts.forEachChild(node, hasFreshNameIdentifier);
}
assert(!hasFreshNameIdentifier(parse("stringFresh.ts", stringFresh)));
assert.throws(() => rewrite("stringFresh.ts", stringFresh, closureRows),
    {message: "stringFresh.ts: witness #1, call 1: elaboratedErrorInfo already occurs at stringFresh.ts:3"});

console.log(`${changed} edited and ${unchanged} unchanged calls verified in adapter snapshots`);
console.log("Diagnostic chains match, input arrays retain identity, and computed chains evaluate once");
console.log("Reference guard and computed block fixtures verified");
console.log("Escaped import refuses before writes, computed name stays fresh, and import pins ignore sibling specifiers");
console.log("Both computed-name freshness checks reject their independent fixtures");
