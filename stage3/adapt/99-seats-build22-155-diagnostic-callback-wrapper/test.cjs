"use strict";

const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const { createHash } = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const ts = require("typescript");

assert.equal(process.argv.length, 3, "usage: node test.cjs <applied TypeScript tree>");
assert.equal(ts.version, "6.0.3");
const applied = path.resolve(process.argv[2]);
const snapshot = "refs/stage3/99-seats-build22-155-diagnostic-callback-wrapper-before";
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-family155-"));
const scannerFile = path.join(scratch, "src/compiler/scanner.ts");
const adapter = path.join(__dirname, "adapt.cjs");
const markers = ["<<<<<<<", "=======", ">>>>>>>", "|||||||"];

function run(command, args) {
    const result = spawnSync(command, args, { encoding: "utf8" });
    if (result.error) throw result.error;
    assert.equal(result.signal, null, command + " terminated by " + result.signal);
    return result;
}

function checked(command, args) {
    const result = run(command, args);
    assert.equal(result.status, 0, result.stdout + result.stderr);
}

function apply() {
    return run(process.execPath, [adapter, scratch]);
}

function emitCompiler() {
    const configFile = path.join(scratch, "src/compiler/tsconfig.json");
    const outDir = path.join(scratch, "emit");
    const typeRoots = [path.resolve(require.resolve("@types/node/package.json"), "../..")];
    const options = {
        outDir, typeRoots, composite: false, incremental: false, declaration: false,
        declarationMap: false, sourceMap: false, emitDeclarationOnly: false, isolatedDeclarations: false,
    };
    const format = errors => ts.formatDiagnosticsWithColorAndContext(errors, {
        getCurrentDirectory: () => scratch,
        getCanonicalFileName: name => name,
        getNewLine: () => "\n",
    });
    const config = ts.getParsedCommandLineOfConfigFile(configFile, options, {
        ...ts.sys,
        onUnRecoverableConfigFileDiagnostic: error => { throw new Error(format([error])); },
    });
    const program = ts.createProgram(config.fileNames, config.options);
    const errors = [...config.errors, ...ts.getPreEmitDiagnostics(program)];
    assert.equal(errors.length, 0, format(errors));
    const emitted = program.emit();
    assert.equal(emitted.emitSkipped, false, format(emitted.diagnostics));
    assert.equal(emitted.diagnostics.length, 0, format(emitted.diagnostics));
    console.log("PASS stock TypeScript", ts.version, "compiler diagnostics", errors.length);
    return outDir;
}

function checkScanner(dir) {
    const { ScriptTarget, LanguageVariant, SyntaxKind } = require(path.join(dir, "compiler/_namespaces/ts.js"));
    const { createScanner, skipTrivia } = require(path.join(dir, "compiler/scanner.js"));
    function checkMarker(marker, skips, jsx) {
        const input = (jsx ? "a\n" : " \n") + marker + (marker === "=======" ? "\n" : " branch\n");
        const events = [];
        const scanner = createScanner(ScriptTarget.Latest, skips,
            jsx ? LanguageVariant.JSX : LanguageVariant.Standard, input,
            (message, length, arg0) => events.push([message.code, scanner.getTokenEnd(), length, arg0]),
            jsx ? 0 : 2);
        const token = jsx ? scanner.scanJsxToken() : scanner.scan();
        assert.deepEqual(events, [[1185, 2, 7, undefined]], marker + " diagnostic");
        assert.equal(token, skips ? SyntaxKind.EndOfFileToken : SyntaxKind.ConflictMarkerTrivia, marker + " token");
        const end = skips || marker === "=======" || marker === "|||||||" ? input.length : input.length - 1;
        assert.equal(scanner.getTokenEnd(), end, marker + " end");
        console.log("PASS", JSON.stringify(marker), "skip", skips, "jsx", jsx,
            "diagnostic", JSON.stringify(events), "token", SyntaxKind[token], "end", end);
    }
    for (const marker of markers) {
        checkMarker(marker, false, false);
        checkMarker(marker, true, false);
    }
    checkMarker("<<<<<<<", false, true);
    const text = "<<<<<<< branch\n";
    assert.equal(skipTrivia(text, 0), text.length);
    const silent = createScanner(ScriptTarget.Latest, false, LanguageVariant.Standard, text);
    assert.equal(silent.scan(), SyntaxKind.ConflictMarkerTrivia);
    assert.equal(silent.getTokenEnd(), text.length - 1);
    console.log("PASS omitted helper callback and absent external error handler");
}

function replaceOnce(text, old, replacement) {
    assert.equal(text.split(old).length, 2, "control anchor " + old);
    return text.replace(old, replacement);
}

function digest(bytes) {
    return createHash("sha256").update(bytes).digest("hex");
}

function reject(label, planted, diagnostic) {
    fs.writeFileSync(scannerFile, planted);
    const before = fs.readFileSync(scannerFile);
    const result = apply();
    assert.notEqual(result.status, 0, label + " accepted");
    assert.match(result.stderr, diagnostic, label + " failed at the wrong guard");
    const after = fs.readFileSync(scannerFile);
    assert.deepEqual(after, before, label + " wrote scanner.ts");
    console.log("PASS", label, "exit", result.status, "before", digest(before), "after", digest(after));
}

function checkGuards(text, state) {
    const helper = /^function scanConflictMarkerTrivia[^\r\n]*\{\r\n/m;
    const header = text.match(helper)[0];
    reject("identity drift " + state, replaceOnce(text, header,
        header + '    if (arguments[2]?.name === "error") return text.length;\r\n'),
    /reviewed helper declaration/);

    const scanHeader = "    function scan(): SyntaxKind {\r\n";
    const shadow = [
        "    function scanConflictMarkerTrivia(text: string, pos: number, error?: (diag: DiagnosticMessage, pos: number, len: number) => void): number {",
        '        if (arguments[2]?.name === "error") return text.length;',
        "        if (error) error(Diagnostics.Merge_conflict_marker_encountered, pos, mergeConflictMarkerLength);",
        "        return text.length;",
        "    }",
        "",
        "",
    ].join("\r\n");
    reject("shadow helper " + state, replaceOnce(text, scanHeader, shadow + scanHeader),
        /reviewed helper binding inventory/);
    const alias = "    const conflictMarkerTrivia = scanConflictMarkerTrivia;\r\n";
    reject("helper alias " + state, replaceOnce(text, scanHeader, alias + scanHeader),
        /reviewed helper binding inventory/);

    const caseStart = "                case CharacterCodes.lessThan:\r\n" +
        "                    if (isConflictMarkerTrivia(text!, pos)) {\r\n";
    const selected = text.slice(text.indexOf(caseStart) + caseStart.length).split("\r\n")[0];
    assert.match(selected, /^ {24}pos = scanConflictMarkerTrivia\(/);
    const slot = caseStart + selected + "\r\n";
    const statement = selected.trimStart();
    reject("extra call " + state, replaceOnce(text, scanHeader,
        scanHeader + "        if (false) pos = scanConflictMarkerTrivia(text!, pos, error);\r\n"),
    /helper call inventory/);
    reject("nested call " + state, replaceOnce(text, slot, caseStart +
        "                        if (true) {\r\n    " + selected + "\r\n                        }\r\n"),
    /first conflict-marker statement/);
    reject("duplicate call " + state, replaceOnce(text, slot, slot + selected + "\r\n"),
        /helper call inventory/);
    reject("call moved out of case " + state, replaceOnce(replaceOnce(text, slot, caseStart), scanHeader,
        scanHeader + "        " + statement + "\r\n"), /first conflict-marker statement/);
}

try {
    const archive = path.join(scratch, "src.tar");
    checked("git", ["-C", applied, "archive", "--format=tar", "--output=" + archive, snapshot, "src"]);
    checked("tar", ["-xf", archive, "-C", scratch]);
    fs.unlinkSync(archive);
    const base = fs.readFileSync(scannerFile, "utf8");
    const first = apply();
    assert.equal(first.status, 0, first.stdout + first.stderr);
    const adapted = fs.readFileSync(scannerFile);
    checkScanner(emitCompiler());
    const repeat = apply();
    assert.equal(repeat.status, 0, repeat.stdout + repeat.stderr);
    assert.deepEqual(fs.readFileSync(scannerFile), adapted, "repeat changed scanner.ts");
    console.log("PASS clean repeat", digest(adapted));
    checkGuards(base, "first apply");
    checkGuards(adapted.toString("utf8"), "second apply");
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
