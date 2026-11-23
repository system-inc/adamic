"use strict";

const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const ts = require("typescript");
const { adapt } = require("./adapt.cjs");

assert.equal(process.argv.length, 3, "usage: node test.cjs <applied tree>");
const tree = process.argv[2];
const adapterRef = "refs/stage3/99-seats-build22-103-spec-diagnostic-arguments";
const parserFile = "src/compiler/commandLineParser.ts";
const diagnosticsFile = "src/compiler/diagnosticInformationMap.generated.ts";

function snapshot(side, file) {
    return execFileSync("git", ["-C", tree, "show", `${adapterRef}-${side}:${file}`]);
}

function parse(file, bytes) {
    const source = ts.createSourceFile(file, bytes.toString("utf8"), ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file} must parse`);
    return source;
}

function named(nodes, name) {
    const matches = nodes.filter(node => ts.isIdentifier(node.name) && node.name.text === name);
    assert.equal(matches.length, 1, `expected one ${name} declaration`);
    return matches[0];
}

function variables(source) {
    return source.statements.filter(ts.isVariableStatement)
        .flatMap(statement => [...statement.declarationList.declarations]);
}

function readValidator(side) {
    const bytes = snapshot(side, parserFile);
    const source = parse(parserFile, bytes);
    const functions = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name);
    const validator = named(functions, "validateSpecs");
    const producer = named(functions, "specToDiagnostic");
    const invalidParent = named(functions, "invalidDotDotAfterRecursiveWildcard");
    const trailingPattern = named(variables(source), "invalidTrailingRecursionPattern");
    return { bytes, source, validator, producer, invalidParent, trailingPattern };
}

const before = readValidator("before");
const after = readValidator("after");
const diagnosticSource = parse(diagnosticsFile, snapshot("before", diagnosticsFile));
const entries = named(variables(diagnosticSource), "Diagnostics").initializer;
assert(ts.isObjectLiteralExpression(entries));
const diagnosticNames = [
    "File_specification_cannot_end_in_a_recursive_directory_wildcard_Asterisk_Asterisk_Colon_0",
    "File_specification_cannot_contain_a_parent_directory_that_appears_after_a_recursive_directory_wildcard_Asterisk_Asterisk_Colon_0",
];
const Diagnostics = Object.fromEntries(diagnosticNames.map(name => {
    const { initializer } = named(entries.properties.filter(ts.isPropertyAssignment), name);
    const call = ts.isAsExpression(initializer) ? initializer.expression : initializer;
    assert(ts.isCallExpression(call));
    assert.equal(call.expression.getText(diagnosticSource), "diag");
    const [code, category, key, message] = call.arguments;
    assert(ts.isNumericLiteral(code) && ts.isPropertyAccessExpression(category));
    assert(ts.isStringLiteral(key) && ts.isStringLiteral(message));
    const categoryValue = ts.DiagnosticCategory[category.name.text];
    assert.equal(typeof categoryValue, "number");
    return [name, { code: Number(code.text), category: categoryValue, key: key.text, message: message.text }];
}));

let trace;
function getTsConfigPropArrayElementValue(sourceFile, specKey, spec) {
    trace.push(["getTsConfigPropArrayElementValue", sourceFile, specKey, spec]);
    return { sourceFile, specKey, spec };
}

function createDiagnosticForNodeInSourceFileOrCompilerDiagnostic(sourceFile, element, message, spec) {
    trace.push(["createDiagnosticForNodeInSourceFileOrCompilerDiagnostic", sourceFile, element, message, spec]);
    return { sourceFile, element, message, spec };
}

const bindings = {
    Diagnostics,
    isString: value => typeof value === "string",
    Debug: { assert },
    startsWith: (text, prefix) => text.startsWith(prefix),
    endsWith: (text, suffix) => text.endsWith(suffix),
    getTsConfigPropArrayElementValue,
    createDiagnosticForNodeInSourceFileOrCompilerDiagnostic,
};

function compileValidator({ source, validator, producer, invalidParent, trailingPattern }) {
    const input = [
        `const ${trailingPattern.getText(source)};`,
        invalidParent.getText(source),
        producer.getText(source),
        validator.getText(source),
    ].join("\n");
    const { outputText } = ts.transpileModule(input, {
        compilerOptions: { target: ts.ScriptTarget.ES2024 },
    });
    // A separate function scope per snapshot and one realm for deep comparison.
    return new Function(...Object.keys(bindings), `"use strict";\n${outputText}\nreturn validateSpecs;`)(
        ...Object.values(bindings),
    );
}

function run(validate, specs, disallowTrailingRecursion, sourceFile, specKey) {
    const errors = [];
    trace = [];
    const kept = validate(specs, errors, disallowTrailingRecursion, sourceFile, specKey);
    return { kept, errors, trace };
}

const original = compileValidator(before);
const adapted = compileValidator(after);
const specs = Object.freeze(["src/**", "**/../lib", "src/**/*.ts", 42]);
const jsonSourceFile = {
    kind: ts.SyntaxKind.SourceFile,
    fileName: "tsconfig.json",
    text: JSON.stringify({ include: specs, exclude: specs }),
};
for (const [specKey, disallowTrailingRecursion] of [["include", true], ["exclude", false]]) {
    for (const sourceFile of [jsonSourceFile, undefined]) {
        const name = `${specKey} / ${sourceFile ? "TsConfigSourceFile" : "undefined"}`;
        const expected = run(original, specs, disallowTrailingRecursion, sourceFile, specKey);
        const actual = run(adapted, specs, disallowTrailingRecursion, sourceFile, specKey);
        assert.deepStrictEqual(actual.kept, expected.kept, `${name}: kept specs`);
        assert.deepStrictEqual(actual.errors, expected.errors, `${name}: diagnostics`);
        assert.deepStrictEqual(actual.trace, expected.trace, `${name}: diagnostic trace`);
        const keptSpecs = disallowTrailingRecursion ? [specs[2]] : [specs[0], specs[2]];
        assert.deepStrictEqual(expected.kept, keptSpecs, `${name}: exercised filter`);
        const diagnosticSpecs = disallowTrailingRecursion ? specs.slice(0, 2) : specs.slice(1, 2);
        assert.deepStrictEqual(expected.errors.map(error => error.spec), diagnosticSpecs, `${name}: exercised rules`);
        for (const [index, error] of actual.errors.entries()) {
            assert.strictEqual(error.message, expected.errors[index].message, `${name}: diagnostic message identity`);
        }
        console.log(`${name}: kept ${JSON.stringify(actual.kept)}, ${actual.errors.length} diagnostics, ${actual.trace.length} trace calls match`);
    }
}

function literalPairs({ producer }) {
    assert(producer.body);
    const returns = [];
    function visit(node) {
        if (ts.isFunctionLike(node)) return;
        if (ts.isReturnStatement(node)) returns.push(node);
        ts.forEachChild(node, visit);
    }
    visit(producer.body);
    assert.equal(returns.length, 2);
    return returns.map(statement => {
        const expression = statement.expression;
        assert(expression && ts.isArrayLiteralExpression(expression), "expected a returned array literal");
        assert.equal(expression.elements.length, 2, "expected a literal pair");
        assert(expression.elements.every(element => !ts.isSpreadElement(element) && !ts.isOmittedExpression(element)));
        return expression;
    });
}

const [firstPair] = literalPairs(before);
literalPairs(after);
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "spec-diagnostic-arguments-"));
try {
    const file = path.join(scratch, parserFile);
    fs.mkdirSync(path.dirname(file), { recursive: true });
    const text = before.source.text;
    const insertAt = firstPair.end - 1;
    fs.writeFileSync(file, text.slice(0, insertAt) + ", spec" + text.slice(insertAt));
    const mutant = fs.readFileSync(file);
    const mtime = fs.statSync(file, { bigint: true }).mtimeNs;
    assert.throws(() => adapt(scratch), /unexpected specToDiagnostic snapshot/);
    assert.deepStrictEqual(fs.readFileSync(file), mutant);
    assert.equal(fs.statSync(file, { bigint: true }).mtimeNs, mtime);

    fs.writeFileSync(file, before.bytes);
    assert.equal(adapt(scratch).edits, 1);
    assert.deepStrictEqual(fs.readFileSync(file), after.bytes);
    const adaptedMtime = fs.statSync(file, { bigint: true }).mtimeNs;
    assert.equal(adapt(scratch).edits, 0);
    assert.deepStrictEqual(fs.readFileSync(file), after.bytes);
    assert.equal(fs.statSync(file, { bigint: true }).mtimeNs, adaptedMtime);
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
console.log("2 literal pairs per snapshot; wider producer refused without a write; exact snapshot replay and no-op");
