"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

// Exact declarations in the fully adapted baseline, including line endings.
const producerSnapshot = `function specToDiagnostic(spec: CompilerOptionsValue, disallowTrailingRecursion?: boolean): [DiagnosticMessage, string] | undefined {
    Debug.assert(typeof spec === "string");
    if (disallowTrailingRecursion && invalidTrailingRecursionPattern.test(spec)) {
        return [Diagnostics.File_specification_cannot_end_in_a_recursive_directory_wildcard_Asterisk_Asterisk_Colon_0, spec];
    }
    else if (invalidDotDotAfterRecursiveWildcard(spec)) {
        return [Diagnostics.File_specification_cannot_contain_a_parent_directory_that_appears_after_a_recursive_directory_wildcard_Asterisk_Asterisk_Colon_0, spec];
    }
}`.split("\n").join("\r\n");
const validatorSnapshot = `function validateSpecs(specs: readonly string[], errors: Diagnostic[], disallowTrailingRecursion: boolean, jsonSourceFile: TsConfigSourceFile | undefined, specKey: string): readonly string[] {
    return specs.filter(spec => {
        if (!isString(spec)) return false;
        const diag = specToDiagnostic(spec, disallowTrailingRecursion);
        if (diag !== undefined) {
            errors.push(createDiagnostic(...diag));
        }
        return diag === undefined;
    });

    function createDiagnostic(message: DiagnosticMessage, spec: string): Diagnostic {
        const element = getTsConfigPropArrayElementValue(jsonSourceFile, specKey, spec);
        return createDiagnosticForNodeInSourceFileOrCompilerDiagnostic(jsonSourceFile, element, message, spec);
    }
}`.split("\n").join("\r\n");

const fileName = "src/compiler/commandLineParser.ts";
const oldCall = "createDiagnostic(...diag)";
const newCall = "createDiagnostic(diag[0], diag[1])";
const updatedValidator = validatorSnapshot.replace(oldCall, newCall);

function topLevelFunction(source, name) {
    const matches = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
    assert.equal(matches.length, 1, `expected one ${name} declaration`);
    return matches[0];
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const file = path.join(tree, fileName);
    const text = fs.readFileSync(file, "utf8");
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, "commandLineParser.ts must parse");

    const producer = topLevelFunction(source, "specToDiagnostic");
    assert(producer.getText(source) === producerSnapshot, "unexpected specToDiagnostic snapshot");

    const validator = topLevelFunction(source, "validateSpecs");
    const validatorText = validator.getText(source);
    if (validatorText === updatedValidator) return { files: 0, edits: 0 };
    assert(validatorText === validatorSnapshot, "unexpected validateSpecs snapshot");

    const calls = [];
    function visit(node) {
        if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) &&
            node.expression.text === "createDiagnostic" && node.arguments.length === 1) {
            const [argument] = node.arguments;
            if (ts.isSpreadElement(argument) && ts.isIdentifier(argument.expression) &&
                argument.expression.text === "diag") {
                calls.push(node);
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(validator);
    assert.equal(calls.length, 1, "expected one createDiagnostic spread call");
    const [call] = calls;
    assert.equal(call.getText(source), oldCall);
    const updated = text.slice(0, call.getStart(source)) + newCall + text.slice(call.end);
    fs.writeFileSync(file, updated);
    return { files: 1, edits: 1, javascriptChanging: true };
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
    console.log(JSON.stringify(adapt(process.argv[2])));
}
