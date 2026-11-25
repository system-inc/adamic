"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

function declaredKinds(file) {
    const declarations = new Map();
    for (const item of file.statements) {
        if (!ts.isInterfaceDeclaration(item) && !ts.isTypeAliasDeclaration(item)) continue;
        const name = item.name.text;
        const previous = declarations.get(name) || [];
        previous.push(item);
        declarations.set(name, previous);
    }
    function declaration(name) {
        const found = declarations.get(name) || [];
        assert.equal(found.length, 1, `Expected one declaration of ${name}`);
        return found[0];
    }
    function namedType(node) {
        assert.ok(ts.isTypeReferenceNode(node) && ts.isIdentifier(node.typeName), node.getText(file));
        assert.ok(!node.typeArguments?.length, node.getText(file));
        return node.typeName.text;
    }
    function members(name, parents = []) {
        assert(!parents.includes(name), `Cyclic union alias ${name}`);
        const item = declaration(name);
        if (ts.isInterfaceDeclaration(item)) return [item];
        const nodes = ts.isUnionTypeNode(item.type) ? item.type.types : [item.type];
        return nodes.flatMap(node => members(namedType(node), [...parents, name]));
    }
    function property(item, name, parents = []) {
        assert.ok(!parents.includes(item), `Cyclic heritage ${item.name.text}`);
        const owned = item.members.filter(member => {
            const key = member.name;
            if (!key) return false;
            if (ts.isComputedPropertyName(key)) {
                assert(ts.isStringLiteralLike(key.expression),
                    `Unresolved computed property name on ${item.name.text}: ${key.getText(file)}`);
                return key.expression.text === name;
            }
            return (ts.isIdentifier(key) || ts.isStringLiteral(key)) && key.text === name;
        });
        assert(owned.length <= 1, `Duplicate ${name} on ${item.name.text}`);
        if (owned.length) {
            assert(ts.isPropertySignature(owned[0]), `${name} on ${item.name.text} is not a property signature`);
            return owned[0];
        }
        const inherited = new Set();
        for (const clause of item.heritageClauses || []) {
            assert.equal(clause.token, ts.SyntaxKind.ExtendsKeyword);
            for (const base of clause.types) {
                assert(ts.isIdentifier(base.expression), base.getText(file));
                const direct = declarations.get(base.expression.text) || [];
                const bases = direct.length && direct.every(ts.isInterfaceDeclaration) ? direct : members(base.expression.text);
                for (const ancestor of bases) {
                    const found = property(ancestor, name, [...parents, item]);
                    if (found) inherited.add(found);
                }
            }
        }
        assert(inherited.size <= 1, `Ambiguous inherited ${name} on ${item.name.text}`);
        return [...inherited][0];
    }
    const syntaxKind = one(file.statements.filter(item => ts.isEnumDeclaration(item) && item.name.text === "SyntaxKind"),
        "Expected one SyntaxKind enum");
    const enumMembers = new Set(syntaxKind.members.map(item => item.name.getText(file)));
    function kind(item) {
        const type = property(item, "kind")?.type;
        assert(type && ts.isTypeReferenceNode(type) && ts.isQualifiedName(type.typeName)
            && ts.isIdentifier(type.typeName.left) && type.typeName.left.text === "SyntaxKind"
            && !type.typeArguments, `Expected a single SyntaxKind literal for ${item.name.text}`);
        const name = type.typeName.right.text;
        assert(enumMembers.has(name), `Unknown SyntaxKind.${name}`);
        return name;
    }
    const signatures = members("SignatureDeclaration");
    const functionLikes = members("FunctionLikeDeclaration");
    const bodyMembers = signatures.filter(item => property(item, "body"));
    assert.deepEqual(bodyMembers.map(item => item.name.text).sort(),
        functionLikes.map(item => item.name.text).sort(),
        "Body-bearing SignatureDeclaration members differ from FunctionLikeDeclaration");
    const allKinds = signatures.map(kind);
    assert.equal(new Set(allKinds).size, allKinds.length, "Duplicate signature kind");
    return bodyMembers.map(kind);
}

function parse(fileName, text) {
    const file = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(file.parseDiagnostics.length, 0, fileName);
    return file;
}

function collect(root, predicate) {
    const found = [];
    function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    }
    visit(root);
    return found;
}

function one(items, label) {
    assert.equal(items.length, 1, label);
    return items[0];
}

function owner(file, name) {
    return one(collect(file, node => ts.isFunctionDeclaration(node) && node.name?.text === name),
        `Expected one function named ${name}`);
}

function assertCoverage(switchNode, kinds, label) {
    const actual = switchNode.caseBlock.clauses.filter(ts.isCaseClause).map(clause => {
        const expression = clause.expression;
        assert(ts.isPropertyAccessExpression(expression) && ts.isIdentifier(expression.expression)
            && expression.expression.text === "SyntaxKind", label);
        return expression.name.text;
    });
    assert.deepEqual(actual.sort(), [...kinds].sort(), `${label} case coverage`);
    assert.equal(switchNode.caseBlock.clauses.filter(ts.isDefaultClause).length, 1, `${label} default`);
}

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node adapt.cjs <adapted-tree>");
const tree = process.argv[2];
const read = relative => {
    const fileName = path.join(tree, "src/compiler", relative);
    return parse(fileName, fs.readFileSync(fileName, "utf8"));
};
const utilities = read("utilities.ts");
const serializer = read("transformers/typeSerializer.ts");
const kinds = declaredKinds(read("types.ts"));
const helperName = "bodyOfSignatureForSerialization";
const eol = file => file.text.includes("\r\n") ? "\r\n" : "\n";
const flagsKinds = kinds;
const flagsBefore = [
    "if (!(node as FunctionLikeDeclaration).body) {",
    "        flags |= FunctionFlags.Invalid;",
    "    }",
].join(eol(utilities));
const flagsAfter = [
    "switch (node.kind) {",
    ...flagsKinds.flatMap(kind => [
        `        case SyntaxKind.${kind}:`,
        "            if (!node.body) {",
        "                flags |= FunctionFlags.Invalid;",
        "            }",
        "            break;",
    ]),
    "        default:",
    "            flags |= FunctionFlags.Invalid;",
    "    }",
].join(eol(utilities));
const helperText = [
    `function ${helperName}(node: SignatureDeclaration) {`,
    "    switch (node.kind) {",
    ...kinds.flatMap(kind => [`        case SyntaxKind.${kind}:`, "            return node.body;"]),
    "        default:",
    "            return undefined;",
    "    }",
    "}",
].join(eol(serializer));
const helperPlan = owner(parse("helper.ts", helperText), helperName);
assertCoverage(helperPlan.body.statements[0], kinds, helperName);
const flagsPlan = one(parse("switch.ts", flagsAfter).statements, "Expected one generated flags switch");
assertCoverage(flagsPlan, kinds, "getFunctionFlags");

const flagsOwner = owner(utilities, "getFunctionFlags");
assert.equal(flagsOwner.parent, utilities, "getFunctionFlags must remain at file scope");
const flagsSite = one(flagsOwner.body.statements.filter(statement =>
    (ts.isIfStatement(statement) || ts.isSwitchStatement(statement)) &&
    collect(statement, node => ts.isPropertyAccessExpression(node) && node.name.text === "body").length),
    "getFunctionFlags body-test statement drifted or was duplicated");

const serializerOwner = owner(serializer, "createRuntimeTypeSerializer");
assert.equal(serializerOwner.parent, serializer, "createRuntimeTypeSerializer must remain at file scope");
const inner = owner(serializer, "serializeParameterTypesOfNode");
assert.equal(inner.parent, serializerOwner.body, "serializeParameterTypesOfNode owner drifted");
const namedCall = (node, name) => ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === name;
const serializerSite = one(collect(inner.body, node =>
    ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.AmpersandAmpersandToken
    && namedCall(node.left, "isFunctionLike") && namedCall(node.right, "nodeIsPresent")),
    "serializeParameterTypesOfNode body test drifted or was duplicated");
const expressionBefore = "isFunctionLike(node) && nodeIsPresent((node as FunctionLikeDeclaration).body)";
const expressionAfter = `isFunctionLike(node) && nodeIsPresent(${helperName}(node))`;
for (const required of ["SignatureDeclaration", "SyntaxKind"]) {
    one(collect(serializer, node => ts.isImportSpecifier(node) && node.name.text === required),
        `Missing or duplicate serializer import for ${required}`);
}
const helpers = collect(serializer, node => ts.isFunctionDeclaration(node) && node.name?.text === helperName);

function state(actual, before, after, label) {
    if (actual === before) return "before";
    assert.equal(actual, after, label);
    return "after";
}

const flagsState = state(flagsSite.getText(utilities), flagsBefore, flagsAfter, "getFunctionFlags exact statement");
const serializerState = state(serializerSite.getText(serializer), expressionBefore, expressionAfter, "serializer exact condition");
assert.equal(flagsState, serializerState, "Mixed before and after states");
if (flagsState === "after") {
    const helper = one(helpers, "Expected one already adapted serializer helper");
    assert.equal(helper.parent, serializer);
    assert.equal(helper.getText(serializer), helperText);
} else {
    assert.equal(helpers.length, 0, "Serializer helper already declared");
    const flagsText = utilities.text.slice(0, flagsSite.getStart(utilities)) + flagsAfter + utilities.text.slice(flagsSite.end);
    const end = serializerOwner.end;
    const serializerText = serializer.text.slice(0, serializerSite.getStart(serializer)) + expressionAfter
        + serializer.text.slice(serializerSite.end, end) + eol(serializer).repeat(2) + helperText + serializer.text.slice(end);
    parse(utilities.fileName, flagsText);
    parse(serializer.fileName, serializerText);
    fs.writeFileSync(utilities.fileName, flagsText);
    fs.writeFileSync(serializer.fileName, serializerText);
}
