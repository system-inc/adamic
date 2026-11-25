#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

// The four columns are declaration, one-based call ordinal, exact details text and operation.
const sites = {
    "src/compiler/checker.ts": [
        ["createTypeChecker #1 / resolveExternalModule #1", 1, "diagnosticDetails", "optional"],
        ["createTypeChecker #1 / resolveExternalModule #1", 2, "errorInfo", "wrap"],
        ["createTypeChecker #1 / errorOnImplicitAnyModule #1", 1, "errorInfo", "optional"],
        ["createTypeChecker #1 / resolveBaseTypesOfClass #1 / diagnostic #1", 1, "elaboration", "optional"],
        ["createTypeChecker #1 / elaborateNeverIntersection #1", 1, "errorInfo", "optional"],
        ["createTypeChecker #1 / elaborateNeverIntersection #1", 2, "errorInfo", "optional"],
        ["createTypeChecker #1 / getPropertyTypeForIndexType #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / getPropertyTypeForIndexType #1", 2, "undefined", "keep"],
        ["createTypeChecker #1 / getPropertyTypeForIndexType #1", 3, "undefined", "keep"],
        ["createTypeChecker #1 / getPropertyTypeForIndexType #1", 4, "undefined", "keep"],
        ["createTypeChecker #1 / getPropertyTypeForIndexType #1", 5, "undefined", "keep"],
        ["createTypeChecker #1 / getPropertyTypeForIndexType #1", 6, "errorInfo", "optional"],
        ["createTypeChecker #1 / checkTypeRelatedTo #1 / reportError #1", 1, "errorInfo", "optional"],
        ["createTypeChecker #1 / checkJsxReturnAssignableToAppropriateBound #1 / generateInitialErrorChain #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / checkJsxOpeningLikeElementOrOpeningFragment #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / reportNonexistentProperty #1", 1, "errorInfo", "optional"],
        ["createTypeChecker #1 / reportNonexistentProperty #1", 2, "errorInfo", "optional"],
        ["createTypeChecker #1 / reportNonexistentProperty #1", 3, "errorInfo", "optional"],
        ["createTypeChecker #1 / reportNonexistentProperty #1", 4, "errorInfo", "optional"],
        ["createTypeChecker #1 / reportNonexistentProperty #1", 5, "errorInfo", "optional"],
        ["createTypeChecker #1 / reportNonexistentProperty #1", 6, "elaborateNeverIntersection(errorInfo, containingType)", "computed"],
        ["createTypeChecker #1 / checkTypeArguments #1 / errorInfo #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / getArgumentArityError #1 / chain #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / getArgumentArityError #1", 1, "chain", "wrap"],
        ["createTypeChecker #1 / getArgumentArityError #1 / chain #2", 1, "undefined", "keep"],
        ["createTypeChecker #1 / getArgumentArityError #1", 2, "chain", "wrap"],
        ["createTypeChecker #1 / getArgumentArityError #1 / chain #3", 1, "undefined", "keep"],
        ["createTypeChecker #1 / getArgumentArityError #1", 3, "chain", "wrap"],
        ["createTypeChecker #1 / getTypeArgumentArityError #1 / chain #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / getTypeArgumentArityError #1", 1, "chain", "wrap"],
        ["createTypeChecker #1 / getTypeArgumentArityError #1 / chain #2", 1, "undefined", "keep"],
        ["createTypeChecker #1 / getTypeArgumentArityError #1", 2, "chain", "wrap"],
        ["createTypeChecker #1 / getTypeArgumentArityError #1 / chain #3", 1, "undefined", "keep"],
        ["createTypeChecker #1 / getTypeArgumentArityError #1", 3, "chain", "wrap"],
        ["createTypeChecker #1 / resolveCall #1", 1, "chain", "optional"],
        ["createTypeChecker #1 / resolveCall #1", 2, "chain", "wrap"],
        ["createTypeChecker #1 / resolveCall #1", 3, "chain", "optional"],
        ["createTypeChecker #1 / resolveCall #1 / chain #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / resolveCall #1 / chain #2", 1, "map(diags, createDiagnosticMessageChainFromDiagnostic)", "keep"],
        ["createTypeChecker #1 / resolveCall #1", 4, "chain", "wrap"],
        ["createTypeChecker #1 / invocationErrorDetails #1", 1, "errorInfo", "keep"],
        ["createTypeChecker #1 / invocationErrorDetails #1", 2, "errorInfo", "wrap"],
        ["createTypeChecker #1 / invocationErrorDetails #1", 3, "undefined", "keep"],
        ["createTypeChecker #1 / invocationErrorDetails #1", 4, "errorInfo", "keep"],
        ["createTypeChecker #1 / invocationErrorDetails #1", 5, "errorInfo", "optional"],
        ["createTypeChecker #1 / invocationErrorDetails #1", 6, "errorInfo", "wrap"],
        ["createTypeChecker #1 / resolveDecorator #1 / messageChain #1", 1, "errorDetails.messageChain", "wrap"],
        ["createTypeChecker #1 / checkTypePredicate #1 / leadingError #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / getAwaitedTypeNoAlias #1", 1, "chain", "optional"],
        ["createTypeChecker #1 / getAwaitedTypeNoAlias #1", 2, "chain", "optional"],
        ["createTypeChecker #1 / checkAsyncFunctionReturnType #1 / errorInfo #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / issueMemberSpecificError #1 / rootChain #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / checkInheritedPropertiesAreIdentical #1 / errorInfo #1", 1, "undefined", "keep"],
        ["createTypeChecker #1 / checkInheritedPropertiesAreIdentical #1", 1, "errorInfo", "wrap"],
    ],
    "src/compiler/program.ts": [
        ["createProgram #3 / verifyCompilerOptions #1", 1, "chainDiagnosticMessages(/*details*/ undefined, Diagnostics.Visit_https_Colon_Slash_Slashaka_ms_Slashts6_for_migration_information)", "wrap"],
        ["createProgram #3 / verifyCompilerOptions #1", 2, "undefined", "keep"],
        ["createProgram #3 / verifyCompilerOptions #1 / verifyEmitFilePath #1", 1, "undefined", "keep"],
        ["createProgram #3 / verifyCompilerOptions #1 / verifyEmitFilePath #1", 2, "chain", "optional"],
        ["createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1 / details #1", 1, "undefined", "keep"],
        ["createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1", 1, "details", "wrap"],
        ["createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1 / chain #1", 1, "details", "wrap"],
        ["createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1", 2, "undefined", "keep"],
        ["createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1 / chain #2", 1, "details", "optional"],
    ],
    "src/compiler/programDiagnostics.ts": [
        ["createProgramDiagnostics #1 / createDiagnosticExplainingFile #1", 1, "cachedChain.details", "keep"],
        ["createProgramDiagnostics #1 / createDiagnosticExplainingFile #1", 2, "fileIncludeReasons", "keep"],
        ["createProgramDiagnostics #1 / createDiagnosticExplainingFile #1", 3, "redirectInfo ?\r\n                    fileIncludeReasonDetails ? [fileIncludeReasonDetails, ...redirectInfo] : redirectInfo :\r\n                    fileIncludeReasonDetails", "branch"],
    ],
    "src/compiler/utilities.ts": [
        ["createModuleNotFoundChain #1 / result #1", 1, "undefined", "keep"],
        ["createModuleNotFoundChain #1 / result #1", 2, "undefined", "keep"],
        ["createModuleNotFoundChain #1 / result #1", 3, "undefined", "keep"],
        ["createModuleNotFoundChain #1 / result #1", 4, "undefined", "keep"],
        ["createModuleNotFoundChain #1 / result #1", 5, "undefined", "keep"],
        ["createModeMismatchDetails #1 / result #1", 1, "undefined", "keep"],
        ["createModeMismatchDetails #1 / result #1", 2, "undefined", "keep"],
        ["createModeMismatchDetails #1 / result #1", 3, "undefined", "keep"],
        ["createModeMismatchDetails #1 / result #1", 4, "undefined", "keep"],
    ],
    "src/compiler/watch.ts": [
        ["explainIfFileIsRedirectAndImpliedFormat #1", 1, "undefined", "keep"],
        ["explainIfFileIsRedirectAndImpliedFormat #1", 2, "undefined", "keep"],
        ["explainIfFileIsRedirectAndImpliedFormat #1", 3, "undefined", "keep"],
        ["explainIfFileIsRedirectAndImpliedFormat #1", 4, "undefined", "keep"],
        ["explainIfFileIsRedirectAndImpliedFormat #1", 5, "undefined", "keep"],
        ["fileIncludeReasonToDiagnostics #1", 1, "undefined", "keep"],
        ["fileIncludeReasonToDiagnostics #1", 2, "undefined", "keep"],
        ["fileIncludeReasonToDiagnostics #1", 3, "undefined", "keep"],
        ["fileIncludeReasonToDiagnostics #1", 4, "undefined", "keep"],
        ["fileIncludeReasonToDiagnostics #1", 5, "undefined", "keep"],
        ["fileIncludeReasonToDiagnostics #1", 6, "undefined", "keep"],
        ["fileIncludeReasonToDiagnostics #1", 7, "undefined", "keep"],
        ["fileIncludeReasonToDiagnostics #1", 8, "undefined", "keep"],
        ["fileIncludeReasonToDiagnostics #1", 9, "undefined", "keep"],
    ],
    "src/testRunner/unittests/diagnosticCollection.ts": [
        ["chain #1", 1, "ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, \"x\")", "wrap"],
        ["chain #1", 2, "undefined", "keep"],
        ["chain #2", 1, "ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, \"a\")", "wrap"],
        ["chain #2", 2, "undefined", "keep"],
        ["chainBetter #1", 1, "ts.chainDiagnosticMessages(\r\n                    ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, \"x\"),\r\n                    ts.Diagnostics.Did_you_mean_0,\r\n                    \"b\",\r\n                )", "wrap"],
        ["chainBetter #1", 2, "ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, \"x\")", "wrap"],
        ["chainBetter #1", 3, "undefined", "keep"],
        ["chain #3", 1, "ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, \"x\")", "wrap"],
        ["chain #3", 2, "undefined", "keep"],
        ["chain #4", 1, "ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, \"x\")", "wrap"],
        ["chain #4", 2, "undefined", "keep"],
        ["chain #5", 1, "ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, \"x\")", "wrap"],
        ["chain #5", 2, "undefined", "keep"],
    ],
};

const nonCallReferences = {
    "src/compiler/checker.ts": "./_namespaces/ts.js",
    "src/compiler/program.ts": "./_namespaces/ts.js",
    "src/compiler/programDiagnostics.ts": "./_namespaces/ts.js",
    "src/compiler/watch.ts": "./_namespaces/ts.js",
};

function collectCalls(ts, source) {
    const owners = new Map();
    const calls = [];
    function visit(node, enclosing) {
        let owner = enclosing;
        if ((ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node) ||
             ts.isFunctionExpression(node) || ts.isClassDeclaration(node) ||
             (ts.isVariableDeclaration(node) && node.initializer)) && node.name) {
            const name = node.name.getText(source);
            const key = enclosing ? `${enclosing.key} / ${name}` : name;
            const occurrence = (owners.get(key) ?? 0) + 1;
            owners.set(key, occurrence);
            owner = { key: `${key} #${occurrence}`, count: 0 };
        }
        if (ts.isCallExpression(node)) {
            const callee = node.expression.getText(source);
            if (callee === "chainDiagnosticMessages" || callee === "ts.chainDiagnosticMessages") {
                assert(owner, `unnamed chainDiagnosticMessages caller in ${source.fileName}`);
                assert(node.arguments[0], `missing details in ${source.fileName}, ${owner.key}`);
                calls.push({ owner: owner.key, ordinal: ++owner.count, node });
            }
        }
        ts.forEachChild(node, child => visit(child, owner));
    }
    visit(source, undefined);
    return calls;
}


const utilities = "src/compiler/utilities.ts";

function findUtility(source) {
    const owners = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === "chainDiagnosticMessages");
    assert.equal(owners.length, 1, utilities);
    assert(ts.isInternalDeclaration(owners[0]), "chainDiagnosticMessages must remain internal");
    return owners[0];
}

function callKey(owner, ordinal) {
    return `${owner}, call ${ordinal}`;
}

function readCalls(source) {
    return new Map(collectCalls(ts, source).map(({owner, ordinal, node}) => [callKey(owner, ordinal), node]));
}

function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, file);
    return source;
}

function optionalArray(value) {
    return `${value} === undefined ? undefined : [${value}]`;
}

function computedStatement(call, source) {
    const assignment = call.parent;
    assert(ts.isBinaryExpression(assignment) && assignment.right === call &&
        assignment.operatorToken.kind === ts.SyntaxKind.EqualsToken &&
        assignment.left.getText(source) === "errorInfo" &&
        ts.isExpressionStatement(assignment.parent), "computed chain assignment moved");
    return assignment.parent;
}

function checkComputedBlock(statement, where) {
    assert(ts.isBlock(statement.parent), `${where}: computed assignment must be directly inside a block`);
    const source = statement.getSourceFile();
    const name = "elaboratedErrorInfo";
    // The inserted const binds references throughout its block, including in earlier closures.
    // A name the file never spells cannot capture an existing reference.
    function findIdentifier(node) {
        if (ts.isIdentifier(node) && node.text === name) return node;
        return ts.forEachChild(node, findIdentifier);
    }
    const identifier = findIdentifier(source);
    const position = identifier ? identifier.getStart(source) : source.text.indexOf(name);
    if (position !== -1) {
        const {line} = source.getLineAndCharacterOfPosition(position);
        assert.fail(`${where}: ${name} already occurs at ${source.fileName}:${line + 1}`);
    }
}

function verifyReferences(source, allowed) {
    const name = "chainDiagnosticMessages";
    const file = source.fileName;
    const pinned = nonCallReferences[file];
    let found = false;
    const positions = new Set();
    let classified = 0;

    function visit(node) {
        if ((ts.isIdentifier(node) || ts.isStringLiteralLike(node)) && node.text === name) {
            if (!allowed.has(node)) {
                const specifier = node.parent;
                const declaration = ts.findAncestor(specifier, ts.isImportDeclaration);
                const {line, character} = source.getLineAndCharacterOfPosition(node.getStart(source));
                assert(pinned && !found && ts.isImportSpecifier(specifier) &&
                    specifier.name === node && !specifier.propertyName && !specifier.isTypeOnly &&
                    specifier.getText(source) === name && declaration?.importClause &&
                    !declaration.importClause.isTypeOnly && declaration.moduleSpecifier.text === pinned,
                    `${file}:${line + 1}:${character + 1}: unlisted ${name} reference (${ts.SyntaxKind[node.kind]})`);
                found = true;
            }
            classified++;
            const offset = node.getText(source).indexOf(name);
            if (offset >= 0) positions.add(node.getStart(source) + offset);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert(!pinned || found, `${file}: missing pinned ${name} reference (ImportDeclaration)`);

    // Classify unmatched raw spellings using trivia at their parser token boundary.
    let textual = 0;
    for (const {index} of source.text.matchAll(/chainDiagnosticMessages/g)) {
        if (!positions.has(index)) {
            const token = ts.getTokenAtPosition(source, index);
            if (ts.findAncestor(token, ts.isJSDoc)) continue;
            const start = token.getFullStart();
            const comments = [
                ...(ts.getLeadingCommentRanges(source.text, start) ?? []),
                ...(ts.getTrailingCommentRanges(source.text, start) ?? []),
            ];
            if (index < token.getStart(source) &&
                comments.some(comment => index >= comment.pos && index + name.length <= comment.end)) continue;
            const {line, character} = source.getLineAndCharacterOfPosition(index);
            assert.fail(`${file}:${line + 1}:${character + 1}: unclassified textual ${name} (${ts.SyntaxKind[token.kind]})`);
        }
        textual++;
    }
    assert.equal(textual, classified, `${file}: textual ${name} reference count`);
}

function rewrite(file, input, rows) {
    const source = parse(file, input);
    const calls = readCalls(source);
    const allowed = new Set();
    const edits = [];
    const replace = (node, text) => edits.push({start: node.getStart(source), end: node.end, text});

    if (file === utilities) {
        const owner = findUtility(source);
        allowed.add(owner.name);
        const details = owner.parameters[0];
        assert.equal(details.getText(source), "details: DiagnosticMessageChain | DiagnosticMessageChain[] | undefined");
        replace(details.type, "DiagnosticMessageChain[] | undefined");
        const returns = owner.body.statements.filter(ts.isReturnStatement);
        assert.equal(returns.length, 1, utilities);
        const result = returns[0].expression;
        assert(ts.isObjectLiteralExpression(result), utilities);
        const next = result.properties.find(node =>
            ts.isPropertyAssignment(node) && node.name.getText(source) === "next");
        assert.equal(next?.initializer.getText(source), "details === undefined || Array.isArray(details) ? details : [details]");
        replace(next.initializer, "details");
    }

    for (const [owner, ordinal, expected, operation] of rows) {
        const key = callKey(owner, ordinal);
        const where = `${file}: ${key}`;
        const call = calls.get(key);
        assert(call, `missing or moved chainDiagnosticMessages at ${where}`);
        calls.delete(key);
        allowed.add(ts.isIdentifier(call.expression) ? call.expression : call.expression.name);
        const argument = call.arguments[0];
        assert.equal(argument.getText(source), expected, where);
        if (operation === "keep") continue;
        if (operation === "wrap") {
            // Separate delimiters preserve the edits to nested arguments.
            const start = argument.getStart(source);
            edits.push({start, end: start, text: "["}, {start: argument.end, end: argument.end, text: "]"});
        } else if (operation === "optional") {
            assert(ts.isIdentifier(argument), where);
            replace(argument, optionalArray(expected));
        } else if (operation === "computed") {
            const statement = computedStatement(call, source);
            checkComputedBlock(statement, where);
            const start = statement.getStart(source);
            const indent = input.slice(input.lastIndexOf("\n", start - 1) + 1, start);
            assert.match(indent, /^ +$/, where);
            const newline = input.includes("\r\n") ? "\r\n" : "\n";
            edits.push({start, end: start, text: `const elaboratedErrorInfo = ${expected};${newline}${indent}`});
            replace(argument, optionalArray("elaboratedErrorInfo"));
        } else {
            assert.equal(operation, "branch", where);
            assert(ts.isConditionalExpression(argument), where);
            assert.equal(argument.whenFalse.getText(source), "fileIncludeReasonDetails", where);
            replace(argument.whenFalse, optionalArray("fileIncludeReasonDetails"));
        }
    }
    assert.equal(calls.size, 0, `unlisted chainDiagnosticMessages at ${file}: ${[...calls.keys()].join(", ")}`);
    verifyReferences(source, allowed);
    let output = input;
    let boundary = input.length;
    for (const {start, end, text} of edits.sort((left, right) => right.start - left.start)) {
        assert(end <= boundary, `overlapping edits in ${file}`);
        output = output.slice(0, start) + text + output.slice(end);
        boundary = start;
    }
    return output;
}

function apply(tree) {
    const verified = new Set();
    const writes = new Map();
    const extensions = [".ts", ".tsx", ".mts", ".cts", ".js", ".jsx", ".mjs", ".cjs"];
    // The stock compiler bundles in lib are build inputs. Audit the source roots.
    const files = ["src", "scripts"].flatMap(directory =>
        ts.sys.readDirectory(path.join(tree, directory), extensions));
    for (const target of files) {
        const file = path.relative(tree, target);
        const input = fs.readFileSync(target, "utf8");
        const output = rewrite(file, input, sites[file] ?? []);
        verified.add(file);
        if (input !== output) writes.set(target, output);
    }
    for (const file of Object.keys(sites)) {
        assert(verified.has(file), `missing call file ${file}: ${callKey(...sites[file][0])}`);
    }
    for (const [target, output] of writes) fs.writeFileSync(target, output);
}

if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs <applied tree>");
    apply(path.resolve(process.argv[2]));
}
module.exports = {sites, utilities, findUtility, readCalls, callKey, parse, computedStatement, checkComputedBlock, rewrite, apply};
