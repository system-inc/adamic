"use strict";

const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const { adapt, file, findOwner } = require("./adapt.cjs");

assert.equal(process.argv.length, 3, "usage: node test.cjs APPLIED_TREE");
assert.equal(ts.version, "6.0.3");
const tree = path.resolve(process.argv[2]);
const ref = "refs/stage3/99-seats-tail24-48-plain-default-matcher";
function snapshot(side) {
    return execFileSync("git", ["-C", tree, "show", ref + "-" + side + ":" + file]);
}

const before = snapshot("before");
const after = snapshot("after");
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "plain-default-matcher-"));
try {
    const target = path.join(scratch, file);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, before);
    assert.equal(adapt(scratch), 1);
    assert.deepEqual(fs.readFileSync(target), after);
    console.log("PASS adapter replay matches the after snapshot byte for byte");
    const time = fs.statSync(target).mtimeMs;
    assert.equal(adapt(scratch), 0);
    assert.deepEqual(fs.readFileSync(target), after);
    assert.equal(fs.statSync(target).mtimeMs, time);
    console.log("PASS adapter replay writes nothing on the after snapshot");

    for (const [label, rewrite, reason] of [
        ["missing", s => s.replace("export function getSubPatternFromSpec(", "export function absent("), /owner drift/],
        ["duplicate", s => s + "\n" + findOwner(s).owner.getText(), /owner drift/],
        ["signature drift", s => s.replace("}: WildcardMatcher = wildcardMatchers[usage],", "}: WildcardMatcher = wildcardMatchers.files,"), /unreviewed matcher parameter/],
        ["binding collision", s => insertAtBody(s, '\n    const matcher = "";'), /first statement drift/],
    ]) {
        const changed = rewrite(before.toString("utf8"));
        assert.notEqual(changed, before.toString("utf8"), label);
        fs.writeFileSync(target, changed);
        assert.throws(() => adapt(scratch), reason, label);
        assert.equal(fs.readFileSync(target, "utf8"), changed, label);
        console.log("PASS " + label + " rejected before writing");
    }
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}

function insertAtBody(text, value) {
    const { source, owner } = findOwner(text);
    const offset = owner.body.getStart(source) + 1;
    return text.slice(0, offset) + value + text.slice(offset);
}

const fields = ["singleAsteriskRegexFragment", "doubleAsteriskRegexFragment", "replaceWildcardCharacter"];
const usages = ["files", "directories", "exclude"];
const specs = ["**", "folder", "*.ts", "src/**/test?.ts"];
function load(bytes) {
    const { source, owner } = findOwner(bytes.toString("utf8"));
    const log = [];
    const receiver = (label, mark) => new Proxy({
        singleAsteriskRegexFragment: "[^/]*" + mark,
        doubleAsteriskRegexFragment: "(/.*)?" + mark,
        replaceWildcardCharacter: char => "[" + char + "]" + mark,
    }, {
        get(target, name) {
            log.push(label + "." + name);
            return Reflect.get(target, name);
        },
    });
    const table = Object.fromEntries(usages.map(usage => [usage, receiver(usage, usage[0])]));
    const explicit = receiver("explicit", "x");
    const globals = {
        wildcardMatchers: new Proxy(table, {
            get(target, name) {
                log.push("table." + name);
                return Reflect.get(target, name);
            },
        }),
        getNormalizedPathComponents(spec, base) {
            log.push("body");
            return ["/", base, ...spec.split("/")];
        },
        last: values => values.at(-1),
        removeTrailingDirectorySeparator: value => value.replace(/\/$/, ""),
        isImplicitGlob: value => !/[.*?]/.test(value),
        directorySeparator: "/",
        CharacterCodes: { asterisk: 42, question: 63 },
        reservedCharacterPattern: /[.*?]/g,
        implicitExcludePathRegexPattern: "(?!node_modules)",
    };
    const code = ts.transpileModule(owner.getText(source).replace(/^export /, ""), {
        compilerOptions: { target: ts.ScriptTarget.ES2022 },
    }).outputText;
    const fn = vm.runInNewContext(code + "\ngetSubPatternFromSpec", globals);
    return {
        length: fn.length,
        run(spec, usage, mode) {
            log.length = 0;
            const extra = { omitted: [], explicit: [explicit], undefined: [undefined], null: [null] }[mode];
            try {
                return { value: fn(spec, "base", usage, ...extra), log: [...log] };
            } catch (error) {
                return { error: error.name, message: error.message, log: [...log] };
            }
        },
    };
}

const old = load(before);
const changed = load(after);
assert.equal(old.length, 3);
assert.equal(changed.length, old.length, "function length");
console.log("PASS function length is 3 on both snapshots");
const messages = new Set();
for (const usage of usages) {
    for (const spec of specs) {
        for (const mode of ["omitted", "explicit", "undefined", "null"]) {
            const a = old.run(spec, usage, mode);
            const b = changed.run(spec, usage, mode);
            if (mode === "null") {
                assert.equal(a.error, "TypeError");
                assert.equal(b.error, "TypeError", "explicit null must throw");
                assert.deepEqual(a.log, []);
                assert.deepEqual(b.log, []);
                messages.add(JSON.stringify([a.message, b.message]));
                continue;
            }
            // Every other mode returns on both sides, so an error thrown the same way on both can't pass the comparison.
            assert.ok("value" in a && !("error" in a), `stock ${usage} ${mode} must return, threw ${a.error}: ${a.message}`);
            assert.ok("value" in b && !("error" in b), `adapted ${usage} ${mode} must return, threw ${b.error}: ${b.message}`);
            const label = mode === "explicit" ? mode : usage;
            const lookup = mode === "explicit" ? [] : ["table." + usage];
            assert.deepEqual(a.log, [...lookup, ...fields.map(name => label + "." + name), "body"]);
            assert.deepEqual(b, a, "result and lookup/field-read order");
        }
    }
    console.log("PASS " + usage + ": 4 specs x 4 input modes");
}
for (const message of messages) console.log("PASS explicit null TypeError messages: " + message);
