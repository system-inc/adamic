"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const sites = [
    ["emitter.ts", "notImplementedResolver", "EmitResolver", "getConstantValue",
        "getConstantValue: notImplemented",
        "getConstantValue: (): string | number | undefined => notImplemented()",
        "getConstantValue: () => notImplemented()"],
    ["transformer.ts", "nullTransformationContext", "TransformationContext", "setLexicalEnvironmentFlags",
        "setLexicalEnvironmentFlags: noop",
        "setLexicalEnvironmentFlags: (_flags: number, _value: boolean = false): void => {}",
        "setLexicalEnvironmentFlags: (_flags, _value = false) => { }"],
    ["transformer.ts", "nullTransformationContext", "TransformationContext", "enableSubstitution",
        "enableSubstitution: noop",
        "enableSubstitution: (_kind: SyntaxKind): void => {}",
        "enableSubstitution: (_kind) => { }"],
];

const noops = [
    ["startLexicalEnvironment", "_unused?: unknown", "_unused"],
    ["resumeLexicalEnvironment", "_unused?: unknown", "_unused"],
    ["suspendLexicalEnvironment", "_unused?: unknown", "_unused"],
    ["hoistVariableDeclaration", "_node: Identifier", "_node"],
    ["hoistFunctionDeclaration", "_node: FunctionDeclaration", "_node"],
    ["addInitializationStatement", "_node: Statement", "_node"],
    ["startBlockScope", "_unused?: unknown", "_unused"],
    ["addBlockScopedVariable", "_node: Identifier", "_node"],
    ["requestEmitHelper", "_helper: EmitHelper", "_helper"],
    ["enableEmitNotification", "_kind: SyntaxKind", "_kind"],
    ["addDiagnostic", "_diag: DiagnosticWithLocation", "_diag"],
];
const throws = [
    ["getEmitResolver", "EmitResolver", "", ""],
    ["getEmitHost", "EmitHost", "", ""],
    ["getEmitHelperFactory", 'import("./factory/emitHelpers.js").EmitHelperFactory', "", ""],
    ["readEmitHelpers", "EmitHelper[] | undefined", "", ""],
    ["isSubstitutionEnabled", "boolean", "_node: Node | undefined = undefined", "_node = undefined"],
    ["isEmitNotificationEnabled", "boolean", "_node: Node | undefined = undefined", "_node = undefined"],
];

for (const [property, parameter, emitted] of noops) {
    sites.push(["transformer.ts", "nullTransformationContext", "TransformationContext", property,
        property + ": noop", property + ": (" + parameter + "): void => {}", property + ": (" + emitted + ") => { }"]);
}
for (const [property, result, parameter, emitted] of throws) {
    sites.push(["transformer.ts", "nullTransformationContext", "TransformationContext", property,
        property + ": notImplemented", property + ": (" + parameter + "): " + result + " => notImplemented()",
        property + ": (" + emitted + ") => notImplemented()"]);
}

for (const [file, owner, typeName, property, original, adapted, emitted] of sites) {
    const fileName = path.join(process.argv[2], "src/compiler", file);
    const before = fs.readFileSync(fileName, "utf8");

    function locate(text) {
        const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, fileName);
        const declarations = source.statements.filter(ts.isVariableStatement)
            .flatMap(node => [...node.declarationList.declarations])
            .filter(node => node.name.getText(source) === owner);
        assert.equal(declarations.length, 1, owner);
        const { initializer, type } = declarations[0];
        assert.equal(type?.getText(source), typeName);
        assert(initializer && ts.isObjectLiteralExpression(initializer));
        const slots = initializer.properties.filter(node => node.name?.getText(source) === property);
        assert.equal(slots.length, 1, property);
        const slot = slots[0];
        assert(ts.isPropertyAssignment(slot));
        const snapshot = slot.getText(source);
        assert(snapshot === original || snapshot === adapted, "unreviewed snapshot " + snapshot);
        return { start: slot.getStart(source), end: slot.end, snapshot };
    }

    const { start, end, snapshot } = locate(before);
    if (snapshot === original) {
        const after = before.slice(0, start) + adapted + before.slice(end);
        assert.equal(locate(after).snapshot, adapted);
        const options = { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext };
        const js = text => ts.transpileModule(text, { compilerOptions: options, fileName }).outputText;
        const oldLine = "    " + original + ",\n";
        const newLine = "    " + emitted + ",\n";
        const beforeJs = js(before);
        assert.equal(beforeJs.split(oldLine).length, 2, "original JavaScript snapshot");
        let expectedJs = beforeJs.replace(oldLine, newLine);
        if (file === "transformer.ts" && original.endsWith(": noop") && !expectedJs.includes(": noop,")) {
            expectedJs = expectedJs.replace(", noop, notImplemented,", ", notImplemented,");
        }
        assert.equal(js(after), expectedJs, "JavaScript outside the stub slot and unused import changed");
        const declaration = text => ts.transpileDeclaration(text, {
            compilerOptions: { stripInternal: true }, fileName,
        }).outputText;
        assert.equal(declaration(after), declaration(before), "public declaration output changed");
        fs.writeFileSync(fileName, after);
    }
}
