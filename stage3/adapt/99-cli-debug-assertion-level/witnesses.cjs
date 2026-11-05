#!/usr/bin/env node
"use strict";
// debug.ts's own statements (LogLevel, LoggingHost and namespace Debug), extracted whole from the tree before 99
// and the tree after it, each transpiled by stock TypeScript 6.0.3 and run under Node with debug.ts's imports
// bound to stock 6.0.3's own. Each run starts a fresh namespace and sets the level from the environment with
// sys.ts's own `Debug.setAssertionLevel(...)` call, extracted from the same tree, before any assertion runs.
// Then each of the six assertion functions is called from a tsc-style call site, `Debug.assertNode(...)`:
// at None, failing; at Normal (`NODE_ENV=development`), passing and failing; and at None, then raised to
// Normal by setAssertionLevel, failing and passing (upstream's restore bug, LEDGER.md entry 8).
// With --mutant, the after tree's shouldAssertFunction writes `(Debug as any)[name] = noop` again.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { shouldBody } = require("./adapt.cjs");

const mutation = [
    "{",
    "        if (!shouldAssert(level)) {",
    "            (Debug as any)[name] = noop;",
    "            return false;",
    "        }",
    "        return true;",
    "    }",
].join("\r\n");

// Each assertion with arguments it passes and arguments it fails, built from stock 6.0.3's factory and guards.
const identifier = ts.factory.createIdentifier("x");
const assertions = {
    assertEachNode: { passing: [[identifier], ts.isIdentifier], failing: [[identifier], ts.isStringLiteral] },
    assertNode: { passing: [identifier, ts.isIdentifier], failing: [identifier, ts.isStringLiteral] },
    assertNotNode: { passing: [identifier, ts.isStringLiteral], failing: [identifier, ts.isIdentifier] },
    assertOptionalNode: { passing: [undefined, ts.isStringLiteral], failing: [identifier, ts.isStringLiteral] },
    assertOptionalToken: { passing: [identifier, ts.SyntaxKind.Identifier], failing: [identifier, ts.SyntaxKind.StringLiteral] },
    assertMissingNode: { passing: [undefined], failing: [identifier] },
};
const development = { NODE_ENV: "development" };

// The ruled rows. raised means the level starts at None, one failing call runs, and then
// Debug.setAssertionLevel(AssertionLevel.Normal) runs before the call observed.
const scenarios = [
    { name: "None, failing", environment: {}, raised: false, call: "failing", ruled: "returns" },
    { name: "Normal, passing", environment: development, raised: false, call: "passing", ruled: "returns" },
    { name: "Normal, failing", environment: development, raised: false, call: "failing", ruled: "fails" },
    { name: "None then Normal, failing", environment: {}, raised: true, call: "failing", ruled: "bug, then fails" },
    { name: "None then Normal, passing", environment: {}, raised: true, call: "passing", ruled: "bug, then returns" },
];

function parse(fileName) {
    return ts.createSourceFile(fileName, fs.readFileSync(fileName, "utf8"), ts.ScriptTarget.Latest, true);
}

// debug.ts without its imports, the value names it imports, and sys.ts's setAssertionLevel call.
function sources(tree, mutant) {
    const debug = parse(path.join(tree, "src/compiler/debug.ts"));
    const imports = debug.statements.filter(ts.isImportDeclaration);
    assert.equal(imports.length, 2, "debug.ts's imports moved");
    const names = imports[1].importClause.namedBindings.elements.filter(element => !element.isTypeOnly).map(element => element.name.text);
    let text = debug.text.slice(imports[1].end);
    if (mutant) {
        assert.equal(text.split(shouldBody.adapted).length - 1, 1, "mutant site drift");
        text = text.replace(shouldBody.adapted, mutation);
        names.push("noop");
    }
    const sys = parse(path.join(tree, "src/compiler/sys.ts"));
    const calls = sys.statements.filter(ts.isIfStatement).flatMap(statement => ts.isBlock(statement.thenStatement) ? statement.thenStatement.statements : [])
        .filter(statement => ts.isExpressionStatement(statement) && ts.isCallExpression(statement.expression) && statement.expression.expression.getText(sys) === "Debug.setAssertionLevel");
    assert.equal(calls.length, 1, "sys.ts's setAssertionLevel call moved");
    return { text, names, startup: calls[0].getText(sys), patched: text.includes("(Debug as any)[") };
}

// A fresh namespace: debug.ts's statements, sys.ts's startup call over a sys whose environment is given, and
// one call site per assertion, `Debug.<name>(...)` as tsc's own modules write them.
function instantiate(source) {
    const sites = Object.keys(assertions).map(name => `    ${name}: (...args: any[]) => (Debug as any).${name}(...args),`).join("\n");
    const module = `${source.text}\nexport function startup(sys: { getEnvironmentVariable(name: string): string }) {\n    ${source.startup}\n}\nexport const sites = {\n${sites}\n};\n`;
    const output = ts.transpileModule(module, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
    const globals = { ts };
    for (const name of source.names) {
        assert.notEqual(ts[name], undefined, `stock 6.0.3 has no ${name}`);
        globals[name] = ts[name];
    }
    const exports = {};
    new Function("exports", ...Object.keys(globals), output)(exports, ...Object.values(globals));
    return exports;
}

function call(instance, name, args) {
    try {
        return { returned: instance.sites[name](...args) };
    }
    catch (error) {
        return { threw: error.constructor.name, message: error.message };
    }
}

function observe(source, name, scenario) {
    const instance = instantiate(source);
    const environment = scenario.environment;
    // Node's sys reads `process.env[name] || ""` (sys.ts's getEnvironmentVariable).
    instance.startup({ getEnvironmentVariable: variable => environment[variable] || "" });
    const assertion = instance.Debug[name];
    if (scenario.raised) {
        call(instance, name, assertions[name].failing);
        instance.Debug.setAssertionLevel(ts.AssertionLevel.Normal);
    }
    const result = call(instance, name, assertions[name][scenario.call]);
    // Whether the namespace member is still the assertion function it started as.
    const member = instance.Debug[name];
    result.member = member === assertion ? "the assertion" : typeof member === "function" ? `another function, ${member.name || "anonymous"}` : `an object, keys ${Object.keys(member).join(", ")}`;
    return result;
}

const debugFailure = observation => observation.threw === "Error" && observation.message.startsWith("Debug Failure. False expression: Unexpected node.");

// A row holds when both sides are the ruled ones. A Debug Failure must be the same message on both sides, and
// the raised row's after must be the same failure that side gives at Normal.
function asRuled(result, normal) {
    const { original, adapted } = result;
    const returned = observation => !observation.threw && observation.returned === undefined;
    const notAFunction = `Debug.${result.assertion} is not a function`;
    switch (result.ruled) {
        case "returns":
            return returned(original) && returned(adapted);
        case "fails":
            return debugFailure(original) && debugFailure(adapted) && original.message === adapted.message;
        case "bug, then fails":
            return original.threw === "TypeError" && original.message === notAFunction
                && debugFailure(adapted) && adapted.message === normal.adapted.message;
        case "bug, then returns":
            return original.threw === "TypeError" && original.message === notAFunction && returned(adapted);
    }
    throw new Error(`unruled row ${result.ruled}`);
}

function main() {
    const args = process.argv.slice(2);
    const mutant = args[0] === "--mutant";
    if (mutant) args.shift();
    assert.equal(args.length, 3, "usage: node witnesses.cjs [--mutant] <tree before 99> <tree after 99> <output.json>");
    assert.equal(ts.version, "6.0.3");
    const [before, after, output] = args;
    const sides = { original: sources(before, false), adapted: sources(after, mutant) };
    assert(sides.original.patched, "the before tree is already adapted");
    assert.equal(sides.adapted.patched, mutant, "the after tree is not adapted");
    const results = [];
    for (const name of Object.keys(assertions)) {
        const normal = {};
        for (const scenario of scenarios) {
            const result = { assertion: name, scenario: scenario.name, ruled: scenario.ruled, original: observe(sides.original, name, scenario), adapted: observe(sides.adapted, name, scenario) };
            if (scenario.name === "Normal, failing") Object.assign(normal, result);
            result.asRuled = asRuled(result, normal);
            results.push(result);
        }
    }
    const failed = results.filter(result => !result.asRuled);
    const moved = results.filter(result => JSON.stringify(result.original) !== JSON.stringify(result.adapted)).map(result => `${result.assertion}, ${result.scenario}`);
    fs.writeFileSync(output, JSON.stringify({ node: process.version, typescript: ts.version, mutant, cases: results.length, failed: failed.map(result => `${result.assertion}, ${result.scenario}`), moved, results }, null, 2) + "\n");
    for (const result of results) {
        console.log(`${result.asRuled ? "as ruled" : "DIFFER  "}  ${result.assertion}, ${result.scenario}: original ${JSON.stringify(result.original)}; adapted ${JSON.stringify(result.adapted)}`);
    }
    for (const line of moved) console.log(`moves  ${line}`);
    console.log(`${results.length - failed.length} of ${results.length} as ruled, ${moved.length} move`);
    if (failed.length) process.exitCode = 1;
}
main();
