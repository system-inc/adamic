#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const { compiler, nodes } = require("../47-host-errors/adapt.cjs");

const overload = "export function some<T>(array: readonly T[] | undefined): array is readonly T[];";
const implementation = `export function some<T>(array: readonly T[] | undefined, predicate?: (value: T) => boolean): boolean {
    if (array !== undefined) {
        if (predicate !== undefined) {
            for (let i = 0; i < array.length; i++) {
                if (predicate(array[i]!)) {
                    return true;
                }
            }
        }
        else {
            return array.length > 0;
        }
    }
    return false;
}`;
const visitorBefore = "some(declarations)";
const visitorAfter = "declarations !== undefined && declarations.length > 0";
const sites = [
    {
        relative: "transformers/destructuring.ts", owner: "flattenDestructuringAssignment",
        local: "let expressions: Expression[] | undefined;",
        before: "!some(expressions)", after: "expressions === undefined || expressions.length === 0",
    },
    {
        relative: "visitorPublic.ts", owner: "visitFunctionBody",
        local: "const declarations = context.endLexicalEnvironment();",
        before: visitorBefore, after: visitorAfter,
    },
    {
        relative: "visitorPublic.ts", owner: "visitIterationBody",
        local: "const declarations = context.endBlockScope();",
        before: visitorBefore, after: visitorAfter,
    },
];

function one(matches, label) {
    assert.equal(matches.length, 1, label);
    return matches[0];
}

function plan(tree, ts) {
    const program = compiler(tree, ts);
    const checker = program.getTypeChecker();
    const directory = path.join(tree, "src/compiler");
    const changes = new Map();
    for (const site of sites) {
        const label = site.relative + " " + site.owner;
        const file = program.getSourceFile(path.join(directory, site.relative));
        assert(file && !file.parseDiagnostics.length, label + " cannot be parsed");
        const owner = one(file.statements.filter(node => ts.isFunctionDeclaration(node)
            && node.name?.text === site.owner && node.body), label + " function body");
        const condition = one(nodes(ts, owner.body, node => ts.isIfStatement(node)
            && [site.before, site.after].includes(node.expression.getText(file))), label + " reviewed if condition").expression;
        const before = condition.getText(file) === site.before;
        const call = ts.isPrefixUnaryExpression(condition) ? condition.operand : condition;
        let argument;
        if (before) {
            assert(ts.isCallExpression(call) && call.arguments.length === 1, label + " expected a one-argument call");
            argument = call.arguments[0];
            const signature = checker.getResolvedSignature(call)?.declaration;
            assert(signature && ts.isFunctionDeclaration(signature) && !signature.body
                && signature.getSourceFile().fileName === path.join(directory, "core.ts")
                && signature.getText() === overload, label + " unreviewed overload");
            const body = checker.getSymbolAtLocation(signature.name)?.declarations
                ?.find(node => ts.isFunctionDeclaration(node) && node.body);
            assert.equal(body?.getText().replaceAll("\r\n", "\n"), implementation, label + " unreviewed some implementation");
        } else {
            assert(ts.isBinaryExpression(condition) && ts.isBinaryExpression(condition.left), label + " unreviewed inline condition");
            argument = condition.left.left;
        }
        assert(ts.isIdentifier(argument), label + " argument is not a local identifier");
        const local = checker.getSymbolAtLocation(argument)?.valueDeclaration;
        assert(local && ts.isVariableDeclaration(local) && ts.isIdentifier(local.name)
            && ts.isVariableDeclarationList(local.parent) && ts.isVariableStatement(local.parent.parent)
            && local.parent.parent.parent === owner.body
            && local.parent.parent.getText(file) === site.local, label + " unreviewed local declaration");
        if (before) {
            if (!changes.has(file)) changes.set(file, []);
            changes.get(file).push({ start: condition.getStart(file), end: condition.end, text: site.after });
        }
    }
    return changes;
}

function adapt(tree, ts) {
    const changes = plan(tree, ts);
    for (const [file, edits] of changes) {
        assert.equal(fs.readFileSync(file.fileName, "utf8"), file.text, "source changed while planning");
        let text = file.text;
        for (const edit of edits.sort((a, b) => b.start - a.start)) {
            text = text.slice(0, edit.start) + edit.text + text.slice(edit.end);
        }
        fs.writeFileSync(file.fileName, text);
    }
}

module.exports = { adapt, plan };
if (require.main === module) {
    try {
        adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript"));
    } catch (error) {
        console.error(error.message);
        process.exitCode = 1;
    }
}
