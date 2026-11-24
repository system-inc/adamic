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
const { file, inspect } = require("./adapt.cjs");

const ref = "refs/stage3/99-seats-build22-78-unique-name-closures";
const applied = process.argv[2];
const options = { target: ts.ScriptTarget.ES2015 };

function snapshot(side, filename) {
    return execFileSync("git", ["-C", applied, "show", ref + "-" + side + ":" + filename],
        { maxBuffer: 8 * 1024 * 1024 });
}

function transpile(compiler, source, filename) {
    const result = compiler.transpileModule(source, {
        compilerOptions: options,
        fileName: filename,
        reportDiagnostics: true,
    });
    assert.strictEqual(result.diagnostics.length, 0, filename + ": " + result.diagnostics
        .map(diagnostic => compiler.flattenDiagnosticMessageText(diagnostic.messageText, "\n")).join("\n"));
    return result.outputText;
}

function statefulCallback() {
    let calls = 0;
    return {
        count: () => calls,
        uniqueName(name) {
            assert.strictEqual(name, "_superIndex");
            calls++;
            return "_superIndex_" + calls;
        },
    };
}

function surroundingText(text, name) {
    const start = text.indexOf(name);
    assert.notStrictEqual(start, -1, "missing rendered name " + name);
    assert.strictEqual(start, text.lastIndexOf(name), "duplicated rendered name " + name);
    return [text.slice(0, start), text.slice(start + name.length)];
}

function renderHelpers(before, after) {
    const replacements = new Map();
    for (const [index, original] of before.sites.entries()) {
        const adapted = after.sites[index];
        assert.strictEqual(adapted.name, original.name);
        const label = original.name;
        const originalInitializer = before.source.slice(original.start, original.end);
        const adaptedInitializer = after.source.slice(adapted.start, adapted.end);
        const originalJS = transpile(ts, before.tag + "\nconst text = " + originalInitializer + ";", label + "-original.ts");
        const adaptedJS = transpile(ts, "const text = " + adaptedInitializer + ";", label + "-adapted.ts");
        const originalText = vm.runInNewContext(originalJS + "\ntext;");
        const adaptedText = vm.runInNewContext(adaptedJS + "\ntext;");
        const originalCallback = statefulCallback();
        const adaptedCallback = statefulCallback();
        assert.strictEqual(originalCallback.count(), 0, label + " original before rendering");
        assert.strictEqual(adaptedCallback.count(), 0, label + " adapted before rendering");

        let initialParts;
        for (const rendering of [1, 2]) {
            const expected = originalText(originalCallback.uniqueName);
            const actual = adaptedText(adaptedCallback.uniqueName);
            const message = label + " rendering " + rendering;
            assert.strictEqual(actual, expected, message + " text");
            assert.strictEqual(originalCallback.count(), rendering, message + " original call count");
            assert.strictEqual(adaptedCallback.count(), rendering, message + " adapted call count");

            const name = "_superIndex_" + rendering;
            const originalParts = surroundingText(expected, name);
            assert.deepStrictEqual(surroundingText(actual, name), originalParts, message + " head and tail");
            if (rendering === 1) initialParts = originalParts;
            else assert.deepStrictEqual(originalParts, initialParts, message + " changes only the name");
            const data = Buffer.from(actual);
            console.log(message + ": " + data.length + " bytes sha256 " + createHash("sha256").update(data).digest("hex"));
        }

        const parsed = ts.createSourceFile(label + ".js", adaptedJS, ts.ScriptTarget.ES2015, true);
        const [declaration] = parsed.statements.filter(ts.isVariableStatement)
            .flatMap(statement => [...statement.declarationList.declarations]);
        replacements.set(label, declaration.initializer.getText(parsed));
    }
    return replacements;
}

// The existing binding forces the printer to rename both index helpers.
const fixture = [
    "const _superIndex = 17;",
    "class Base {",
    "    get x() { return 0; }",
    "    set x(value: number) {}",
    "}",
    "class ReadOne extends Base {",
    '    async read(key: "x") { return super.x + super[key]; }',
    "}",
    "class ReadTwo extends Base {",
    '    async read(key: "x") { return super.x + super[key]; }',
    "}",
    "class WriteOne extends Base {",
    '    async write(key: "x") { super.x = 1; super[key] = 1; }',
    "}",
    "class WriteTwo extends Base {",
    '    async write(key: "x") { super.x = 1; super[key] = 1; }',
    "}",
    "",
].join("\n");

async function checkPrinter(replacements) {
    const bundle = require.resolve("typescript");
    const source = await fs.promises.readFile(bundle, "utf8");
    const parsed = ts.createSourceFile(bundle, source, ts.ScriptTarget.Latest, true);
    assert.deepEqual(parsed.parseDiagnostics, [], bundle);
    const properties = new Map([...replacements.keys()].map(name => [name, []]));
    function visit(node) {
        if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && properties.has(node.name.text)
            && node.initializer && ts.isObjectLiteralExpression(node.initializer)) {
            for (const property of node.initializer.properties) {
                if (ts.isPropertyAssignment(property) && ts.isIdentifier(property.name) && property.name.text === "text") {
                    properties.get(node.name.text).push(property.initializer);
                }
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(parsed);

    const edits = [];
    for (const [name, values] of properties) {
        if (values.length !== 1) return name + ": stock bundle has " + values.length + " text initializers";
        edits.push({ start: values[0].getStart(parsed), end: values[0].end, text: replacements.get(name) });
    }
    let patched = source;
    for (const edit of edits.sort((left, right) => right.start - left.start)) {
        patched = patched.slice(0, edit.start) + edit.text + patched.slice(edit.end);
    }
    const stock = transpile(ts, fixture, "async-super.ts");
    const readers = (stock.match(/const _superIndex(?:_\d+)? = name => super\[name\];/g) || []).length;
    const writers = (stock.match(/const _superIndex(?:_\d+)? = \(function \(geti, seti\)/g) || []).length;
    console.log("Stock printer fixture:\n" + fixture);
    console.log("Stock printer output:\n" + stock);
    if (readers !== 2 || writers !== 2) {
        return "stock printer rendered " + readers + " read helpers and " + writers + " write helpers";
    }

    const module = { exports: {} };
    vm.runInNewContext(patched, {
        module, exports: module.exports, require: createRequire(bundle),
        __filename: bundle, __dirname: path.dirname(bundle), process, Buffer,
        setTimeout, clearTimeout, setImmediate, clearImmediate,
    }, { filename: bundle });
    assert.strictEqual(module.exports.version, ts.version);
    assert.strictEqual(transpile(module.exports, fixture, "async-super.ts"), stock, "stock printer output");
    console.log("Stock and adapted printer output match (2 read scopes and 2 write scopes)");
    return null;
}

function replaceInitializer(source, site, value) {
    return source.slice(0, site.start) + value + source.slice(site.end);
}

function checkText(before, after, site, tag) {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-family78-"));
    const scratchFile = path.join(scratch, file);
    try {
        fs.mkdirSync(path.dirname(scratchFile), { recursive: true });
        const run = () => spawnSync(process.execPath, [path.join(__dirname, "adapt.cjs"), scratch], { encoding: "utf8" });
        fs.writeFileSync(scratchFile, before);
        for (const label of ["first apply", "second apply"]) {
            const previousMtime = fs.statSync(scratchFile, { bigint: true }).mtimeNs;
            const result = run();
            assert.ifError(result.error);
            assert.strictEqual(result.status, 0, label + ": " + result.stderr);
            assert.deepStrictEqual(fs.readFileSync(scratchFile), after, label);
            if (label === "second apply") {
                assert.strictEqual(fs.statSync(scratchFile, { bigint: true }).mtimeNs, previousMtime, "second apply wrote the file");
            }
            console.log(label + " matches the after snapshot byte for byte");
        }

        const source = before.toString("utf8");
        const substitution = '${"_superIndex"}';
        assert.strictEqual(site.before.split(substitution).length, 2, "one original substitution");
        const returnLine = "        return result;";
        assert.strictEqual(tag.split(returnLine).length, 2, "one return line in helperString");
        const changedTag = tag.replace(returnLine, '        return result + "\\n";');
        const afterSource = after.toString("utf8");
        assert.strictEqual(source.split(tag).length, 2, "one tag in before snapshot");
        assert.strictEqual(afterSource.split(tag).length, 2, "one tag in after snapshot");
        const cases = [
            ["partly adapted input", replaceInitializer(source, site, site.after), /partly adapted emitHelpers\.ts/],
            ["second substitution", replaceInitializer(source, site, site.before.replace(substitution, substitution + substitution)), /expected exactly one substitution/],
            ["before-file tag drift", source.replace(tag, changedTag), /helperString: tag drift/],
            ["after-file tag drift", afterSource.replace(tag, changedTag), /helperString: tag drift/],
        ];
        for (const [label, text, diagnostic] of cases) {
            const input = Buffer.from(text);
            fs.writeFileSync(scratchFile, input);
            const result = run();
            assert.ifError(result.error);
            assert(Number.isInteger(result.status) && result.status !== 0, label + " must exit nonzero");
            assert.match(result.stderr, diagnostic);
            assert.deepStrictEqual(fs.readFileSync(scratchFile), input, String(diagnostic));
            console.log("Refuses " + label + " (" + diagnostic.source + ") with input bytes intact");
        }
    } finally {
        fs.rmSync(scratch, { recursive: true, force: true });
    }
}

async function main() {
    assert.strictEqual(process.argv.length, 3, "usage: node test.cjs <applied tree>");
    const read = side => {
        const bytes = snapshot(side, file);
        const source = bytes.toString("utf8");
        return { bytes, source, ...inspect(source) };
    };
    const before = read("before");
    const after = read("after");
    assert.strictEqual(before.sites.length, 2);
    assert.strictEqual(after.sites.length, 2);
    assert(before.sites.every(site => site.original), "before snapshot must hold both tags");
    assert(after.sites.every(site => !site.original), "after snapshot must hold both closures");

    const replacements = renderHelpers(before, after);
    const printerProblem = await checkPrinter(replacements);
    checkText(before.bytes, after.bytes, before.sites.find(site => site.name === "asyncSuperHelper"), before.tag);
    if (printerProblem) throw new Error(printerProblem);
}

main().catch(error => {
    console.error(error);
    process.exitCode = 1;
});
