"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const { createHash } = require("node:crypto");
const fs = require("node:fs");
const { createRequire } = require("node:module");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");

assert.equal(process.argv.length, 3, "usage: node test.cjs APPLIED_TREE");
assert.equal(ts.version, "6.0.3");
const tree = path.resolve(process.argv[2]);
const snapshot = "refs/stage3/99-seats-build22-41-pragma-arg-regexes";
const adapter = path.join(__dirname, "adapt.cjs");
const parser = "src/compiler/parser.ts";
const types = "src/compiler/types.ts";
const names = ["types", "lib", "path", "no-default-lib", "resolution-mode", "preserve", "name"];
function readSnapshot(side, file) {
    return execFileSync("git", ["-C", tree, "show", `${snapshot}-${side}:${file}`]);
}
const input = Object.fromEntries([parser, types].map(file => [file, readSnapshot("before", file)]));
const recordedParser = readSnapshot("after", parser);
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), "pragma-regex-test-"));

const barrel = "src/compiler/_namespaces/ts.ts";
const fixtureSource = path.join(tmp, "before");
let snapshotFiles;

function fixture(label, parserBytes = input[parser], typesBytes = input[types], extras = {}) {
    const root = path.join(tmp, label);
    fs.cpSync(fixtureSource, root, { recursive: true });
    fs.writeFileSync(path.join(root, parser), parserBytes);
    fs.writeFileSync(path.join(root, types), typesBytes);
    for (const [file, bytes] of Object.entries(extras)) fs.writeFileSync(path.join(root, file), bytes);
    return root;
}

function digest(root, file) {
    return createHash("sha256").update(fs.readFileSync(path.join(root, file))).digest("hex");
}

function apply(root, extraFiles = []) {
    const files = new Set([parser, types, ...extraFiles]);
    const digestFiles = () => Array.from(files, file => digest(root, file));
    const before = digestFiles();
    const result = spawnSync(process.execPath, [adapter, root], { encoding: "utf8" });
    if (result.error) throw result.error;
    const after = digestFiles();
    const changes = after.map((hash, index) => hash !== before[index]);
    const details = `exit ${result.status}; stdout ${result.stdout.trim() || "(empty)"}; parser changed ${changes[0]}; types changed ${changes[1]}`;
    return { ...result, changes, details };
}

function accepted(root, edits) {
    const result = apply(root);
    const label = `clean application ${edits ? "1" : "2"}`;
    assert.equal(result.status, 0, `${label}: ${result.details}\n${result.stderr}`);
    assert.deepEqual(JSON.parse(result.stdout), { files: edits, edits });
    assert.deepEqual(result.changes, [Boolean(edits), false], label);
    console.log(`PASS ${label}: ${result.details}`);
}

function rejected(root, label, reason, extraFiles) {
    const result = apply(root, extraFiles);
    assert(result.status > 0, `${label}: expected adapter rejection; ${result.details}`);
    assert.match(result.stderr, reason);
    assert(result.changes.every(changed => !changed), `${label}: ${result.details}`);
    const diagnostic = result.stderr.split("\n").find(line => reason.test(line));
    console.log(`PASS ${label}: ${result.details}; ${diagnostic.trim()}`);
    fs.rmSync(root, { recursive: true, force: true });
}

function replaceOnce(text, before, after) {
    assert.equal(text.split(before).length, 2, `Expected one fixture site for ${before}`);
    return text.replace(before, () => after);
}

function changedCallerGuard(bytes) {
    const text = bytes.toString("utf8");
    const newline = text.includes("\r\n") ? "\r\n" : "\n";
    const condition = "if (!pragma || !(pragma.kind! & PragmaKindFlags.TripleSlashXML)) {" + newline;
    return replaceOnce(text, condition + "            return;", condition + "            if (!pragma) return;");
}

function changedCallerBinding(bytes) {
    return replaceOnce(bytes.toString("utf8"), "const matcher = getNamedArgRegEx(arg.name);",
        "const matcher = ((arg: { name: string }) => getNamedArgRegEx(arg.name))(commentPragmas.jsx.args[0]!);");
}

function shadowedPragmaTable(bytes) {
    const text = bytes.toString("utf8");
    const newline = text.includes("\r\n") ? "\r\n" : "\n";
    const branch = "    if (tripleSlash) {" + newline;
    const changed = replaceOnce(text, branch,
        '    const commentPragmas = { reference: { args: [{ name: "factory" }], kind: PragmaKindFlags.TripleSlashXML } } as const;' +
        newline + branch);
    const syntax = ts.createSourceFile(parser, changed, ts.ScriptTarget.Latest, true);
    assert.equal(syntax.parseDiagnostics.length, 0, "Shadowed pragma table fixture must parse cleanly");
    return changed;
}

function insertAfter(bytes, anchor, ...lines) {
    const text = bytes.toString("utf8");
    const newline = text.includes("\r\n") ? "\r\n" : "\n";
    return replaceOnce(text, anchor + newline, [anchor, ...lines, ""].join(newline));
}

function producerFixtures(parserBytes, application) {
    const typeAnchor = "export type IntroducesNewScopeNode = SignatureDeclaration | JSDocSignature | MappedTypeNode;";
    const redirectedTypes = insertAfter(input[types], typeAnchor, "",
        'export const alternateCommentPragmas: typeof commentPragmas = { ...commentPragmas, reference: { ...commentPragmas.reference, args: [...commentPragmas.reference.args, { name: "factory", optional: true }] as unknown as typeof commentPragmas.reference.args } };');
    const redirectedBarrel = insertAfter(readSnapshot("before", barrel), "export { performance };", "",
        'export { alternateCommentPragmas as commentPragmas } from "../types.js";');
    rejected(fixture("redirect", parserBytes, redirectedTypes, { [barrel]: redirectedBarrel }),
        `redirected pragma table before ${application} apply`, /Unreviewed pragma import target/, [barrel]);
    const writer = insertAfter(input[types], typeAnchor, "",
        'Object.assign(commentPragmas.reference, { args: [...commentPragmas.reference.args, { name: "factory", optional: true }] });');
    rejected(fixture("writer", parserBytes, writer),
        `pragma table writer before ${application} apply`, /Unreviewed pragma table mention/);
    const namespaceReason = /Unreviewed namespace use: src\/compiler\/types\.ts:/;
    const probes = [
        [
            "namespace alias (R1)",
            "const namespaceAlias = ts;",
            "export function writeThroughAlias(key: keyof typeof ts) { Object.assign(namespaceAlias[key] as object, { reference: undefined }); }",
        ],
        [
            "namespace values (R2)",
            "export function appendThroughValues() { for (const entry of Object.values(ts)) if (entry && typeof entry === \"object\" && \"reference\" in entry && \"jsx\" in entry) Object.assign((entry as { reference: object }).reference, { args: [...(entry as { reference: { args: object[] } }).reference.args, { name: \"factory\", optional: true }] }); }",
        ],
        [
            "namespace passed to call (R3)",
            "function reflectOn(target: object, key: string) { Object.assign((target as Record<string, object>)[key], { reference: undefined }); }",
            "export function writeThroughCall(key: string) { reflectOn(ts, key); }",
        ],
        [
            "namespace Reflect.get (R4)",
            "export function writeThroughReflect(key: string) { Object.assign(Reflect.get(ts, key) as object, { reference: undefined }); }",
        ],
        [
            "namespace spread (R5)",
            "export function writeThroughSpread(key: string) { Object.assign(({ ...ts } as Record<string, object>)[key], { reference: undefined }); }",
        ],
        [
            "namespace cast alias (R6)",
            "const viaNames = ts as unknown as Record<string, object>;",
            "export function writeThroughNames() { for (const key of Object.getOwnPropertyNames(viaNames)) if (key.startsWith(\"comment\")) Object.assign(viaNames[key]!, { reference: undefined }); }",
        ],
    ];
    for (const [label, ...lines] of probes) {
        const changed = insertAfter(input[types], typeAnchor, "", 'import * as ts from "./_namespaces/ts.js";', ...lines);
        const syntax = ts.createSourceFile(types, changed, ts.ScriptTarget.Latest, true);
        assert.equal(syntax.parseDiagnostics.length, 0, label);
        assert.equal(changed.split("commentPragmas").length, input[types].toString("utf8").split("commentPragmas").length, label);
        rejected(fixture("namespace-probe", parserBytes, changed),
            `${label} before ${application} apply`, namespaceReason, snapshotFiles);
    }
    const reflected = insertAfter(input[types], typeAnchor, "",
        'import * as ts from "./_namespaces/ts.js";',
        'export function writeNamespaceEntry(key: keyof typeof ts) { Object.assign(ts[key] as object, { reference: undefined }); }');
    rejected(fixture("reflection", parserBytes, reflected),
        `namespace reflection before ${application} apply`, namespaceReason);
    const aliased = insertAfter(parserBytes,
        'function getNamedPragmaArguments(pragma: PragmaDefinition, text: string | undefined): { [index: string]: string | undefined; } | "fail" {',
        '    Object.assign(pragma, { kind: pragma.kind! | PragmaKindFlags.TripleSlashXML });');
    const syntax = ts.createSourceFile(parser, aliased, ts.ScriptTarget.Latest, true);
    assert.equal(syntax.parseDiagnostics.length, 0, "Aliased pragma entry fixture must parse cleanly");
    rejected(fixture("aliased-entry", aliased),
        `aliased pragma entry write before ${application} apply`, /Unreviewed pragma entry callee/, snapshotFiles);
}

function matcher(source, file) {
    const syntax = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true);
    assert.equal(syntax.parseDiagnostics.length, 0);
    const matches = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === "getNamedArgRegEx") matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(syntax);
    assert.equal(matches.length, 1);
    return matches[0];
}

function transpile(text) {
    const result = ts.transpileModule(text, {
        compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext },
        reportDiagnostics: true,
    });
    assert.deepEqual(result.diagnostics, []);
    return result.outputText;
}

function evaluateWithCache(helper) {
    return vm.runInNewContext("const namedArgRegExCache = new Map();\n" + helper +
        "\n({ getNamedArgRegEx, namedArgRegExCache });");
}

function evaluate(helper) {
    return evaluateWithCache(helper).getNamedArgRegEx;
}

function observation(regex, text) {
    const match = regex.exec(text);
    if (!match) return null;
    return [match.index, match.index + match[1].length + 1, match[1], match[2], match[3]];
}

function testCaptures(original, adapted) {
    const base = evaluate(original);
    const head = evaluate(adapted);
    let count = 0;
    for (const name of names) {
        const before = base(name);
        const after = head(name);
        assert.strictEqual(base(name), before, `original cache: ${name}`);
        assert.strictEqual(head(name), after, `adapted cache: ${name}`);
        const mixed = name.replace(/[a-z]/g, (letter, index) => index % 2 ? letter : letter.toUpperCase());
        const rows = [
            [0, " " + name.toUpperCase() + "  =  'a.ts'", "a.ts", undefined],
            [14, "long prefix \u{1f600} " + mixed + ' ="b.ts"', undefined, "b.ts"],
            [6, "prefix\u00a0" + name + "=''", "", undefined],
        ];
        for (const [index, text, single, double] of rows) {
            const label = `TestPragmaRegexCapturesAndOffsets ${name} at index ${index}`;
            const expected = observation(before, text);
            assert.equal(expected?.[0], index, label);
            assert.deepEqual(expected.slice(3), [single, double], label);
            assert.deepEqual(observation(after, text), expected, label);
            count++;
        }
    }
    console.log(`PASS TestPragmaRegexCapturesAndOffsets: ${count} rows, indices 0/14/6, seven cache identities`);
}

function testUnknownNameThrows(adapted) {
    const { getNamedArgRegEx, namedArgRegExCache } = evaluateWithCache(adapted);
    for (const attempt of [1, 2]) {
        const label = `TestPragmaUnknownNameThrows attempt ${attempt}`;
        assert.throws(() => getNamedArgRegEx("factory"), {
            name: "Error",
            message: "Unknown triple-slash pragma argument: factory",
        }, label);
        assert.equal(namedArgRegExCache.has("factory"), false, label);
    }
    console.log("PASS TestPragmaUnknownNameThrows: both calls threw the approved error without caching factory");
}

function testParser(adapted) {
    const bundle = require.resolve("typescript");
    const source = fs.readFileSync(bundle, "utf8");
    const original = matcher(source, "typescript.js");
    const patched = source.slice(0, original.getStart()) + adapted + source.slice(original.end);
    const module = { exports: {} };
    vm.compileFunction(patched, ["exports", "require", "module", "__filename", "__dirname"], { filename: bundle })(
        module.exports, createRequire(bundle), module, bundle, path.dirname(bundle));
    const parse = (api, text) => Object.fromEntries(
        api.createSourceFile("fixture.ts", text, api.ScriptTarget.ESNext, false, api.ScriptKind.TSX).pragmas);
    const comments = [
        "/// <reference TYPES='pkg' lib=\"dom\" path='./unit.ts' no-default-lib='true' resolution-mode='require' preserve='true' />\nlet x = 1;",
        "// leading comment\n/// <reference path=\"./unit.ts\" />\nlet x = 1;",
        "/// <amd-module name=\"sample\" />\n/// <amd-dependency path='dep' name='depname' />\nlet x = 1;",
        "/* @jsx h */\nlet x = 1;",
        "/// <jsx factory=\"h\" />\nlet x = 1;",
    ];
    assert.equal(parse(ts, comments[1]).reference.range.pos, 19);
    for (const [index, text] of comments.entries()) {
        const expected = JSON.stringify(parse(ts, text));
        assert.equal(JSON.stringify(parse(module.exports, text)), expected, `TestPragmaParserObservations comment ${index + 1}`);
        console.log(`PASS TestPragmaParserObservations comment ${index + 1}: ${expected}`);
    }
}

try {
    const archive = path.join(tmp, "before.tar");
    execFileSync("git", ["-C", tree, "archive", "--format=tar", `--output=${archive}`, `${snapshot}-before`, "src"]);
    fs.mkdirSync(fixtureSource);
    execFileSync("tar", ["-xf", archive, "-C", fixtureSource]);
    snapshotFiles = fs.readdirSync(path.join(fixtureSource, "src"), { recursive: true, withFileTypes: true })
        .filter(entry => entry.isFile())
        .map(entry => path.relative(fixtureSource, path.join(entry.parentPath, entry.name)));
    const table = input[types].toString("utf8");
    const newline = table.includes("\r\n") ? "\r\n" : "\n";
    const jsx = ['    "jsx": {', '        args: [{ name: "factory" }],', "        kind: PragmaKindFlags.MultiLine,"].join(newline);
    const combined = replaceOnce(table, jsx, jsx.replace("MultiLine,", "MultiLine | PragmaKindFlags.TripleSlashXML,"));
    const preserve = '            { name: "preserve", optional: true },';
    const extra = replaceOnce(table, preserve, preserve + newline + '            { name: "extra", optional: true },');
    const changedEnum = replaceOnce(table, "MultiLine = 1 << 2,", "MultiLine = (1 << 2) | TripleSlashXML,");
    const enumReason = /Unreviewed PragmaKindFlags declaration/;
    const combinedReason = /Unreviewed pragma kind.*MultiLine \| PragmaKindFlags\.TripleSlashXML/;

    rejected(fixture("combined-first", input[parser], combined), "combined kind before first apply", combinedReason);
    rejected(fixture("enum-first", input[parser], changedEnum), "changed enum before first apply", enumReason);
    const guardReason = /Unreviewed pragma caller guard/;
    rejected(fixture("caller-guard-first", changedCallerGuard(input[parser])), "changed caller guard before first apply", guardReason);
    const blockReason = /Unreviewed pragma caller block/;
    const bindingReason = /Unreviewed pragma caller binding/;
    rejected(fixture("caller-binding-first", changedCallerBinding(input[parser])), "changed caller binding before first apply", blockReason);
    rejected(fixture("shadowed-table-first", shadowedPragmaTable(input[parser])), "shadowed pragma table before first apply", bindingReason);
    producerFixtures(input[parser], "first");
    const clean = fixture("clean");
    accepted(clean, 1);
    const adapted = fs.readFileSync(path.join(clean, parser));
    assert(adapted.equals(recordedParser), "Adapted parser.ts differs from the adapter's -after snapshot");
    console.log("PASS scratch parser.ts matches the adapter's -after snapshot");
    rejected(fixture("caller-guard-second", changedCallerGuard(adapted)), "changed caller guard before second apply", guardReason);
    rejected(fixture("caller-binding-second", changedCallerBinding(adapted)), "changed caller binding before second apply", blockReason);
    rejected(fixture("shadowed-table-second", shadowedPragmaTable(adapted)), "shadowed pragma table before second apply", bindingReason);
    producerFixtures(adapted, "second");
    rejected(fixture("combined-second", adapted, combined), "combined kind before second apply", combinedReason);
    rejected(fixture("enum-second", adapted, changedEnum), "changed enum before second apply", enumReason);
    accepted(clean, 0);
    rejected(fixture("extra-caller", input[parser].toString() + '\ngetNamedArgRegEx("lib");\n'),
        "extra matcher caller", /Expected one matcher and one caller/);
    rejected(fixture("extra-name", input[parser], extra), "eighth XML argument", /Unreviewed triple-slash pragma name set/);

    const original = transpile(matcher(input[parser].toString(), "parser.ts").getText());
    const helper = transpile(matcher(adapted.toString(), "parser.ts").getText());
    testCaptures(original, helper);
    testUnknownNameThrows(helper);
    testParser(helper);
} finally {
    fs.rmSync(tmp, { recursive: true, force: true });
}
