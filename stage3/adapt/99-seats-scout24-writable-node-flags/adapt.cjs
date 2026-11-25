#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const writes = {
    "src/compiler/binder.ts": [
        "(node as Mutable<Node>).flags &= ~NodeFlags.Unreachable;",
        "(node as Mutable<Node>).flags |= NodeFlags.Unreachable;",
        "(node.label as Mutable<Node>).flags |= NodeFlags.Unreachable;"
    ],
    "src/compiler/program.ts": [
        "(externalHelpersModuleReference as Mutable<Node>).flags &= ~NodeFlags.Synthesized;",
        "(importDecl as Mutable<Node>).flags &= ~NodeFlags.Synthesized;"
    ],
    "src/compiler/factory/nodeFactory.ts": [
        "(clone as Mutable<T>).flags |= node.flags & ~NodeFlags.Synthesized;",
        "(node as Mutable<Node>).flags |= NodeFlags.Synthesized;"
    ],
    "src/compiler/parser.ts": [
        "(newSourceFile as Mutable<SourceFile>).flags |= sourceFile.flags & NodeFlags.PermanentlySetIncrementalFlags;",
        "(node as Mutable<T>).flags |= NodeFlags.Deprecated;",
        "(node as Mutable<T>).flags |= contextFlags;",
        "(node as Mutable<T>).flags |= NodeFlags.ThisNodeHasError;",
        "(node as Mutable<Node>).flags = type.flags;",
        "(node as Mutable<NonNullExpression>).flags |= NodeFlags.OptionalChain;",
        "(tagExpression as Mutable<Node>).flags |= NodeFlags.OptionalChain;",
        "(m as Mutable<Node>).flags |= NodeFlags.Ambient;",
        "(m as Mutable<Node>).flags |= NodeFlags.Ambient;",
        "(typeNameOrNamespaceName as Mutable<Identifier>).flags |= NodeFlags.IdentifierIsInJSDocNamespace;"
    ],
    "src/compiler/utilities.ts": [
        "(node as Mutable<Node>).flags |= NodeFlags.ThisNodeOrAnySubNodesHasError;",
        "(node as Mutable<Node>).flags |= NodeFlags.HasAggregatedChildData;",
        "(node as Mutable<T>).flags = newFlags;"
    ]
};
// Pin the 20 direct writes by exact text and enclosing function.
const directFunctions = {
    "src/compiler/binder.ts": [
        "bindChildren",
        "bindChildren",
        "bindLabeledStatement"
    ],
    "src/compiler/factory/nodeFactory.ts": [
        "cloneNode",
        "makeSynthetic"
    ],
    "src/compiler/parser.ts": [
        "updateSourceFile",
        "withJSDoc",
        "finishNode",
        "finishNode",
        "parseTupleElementType",
        "tryReparseOptionalChain",
        "parseTaggedTemplateRest",
        "parseDeclaration",
        "parseClassElement",
        "parseJSDocTypeNameWithNamespace"
    ],
    "src/compiler/program.ts": [
        "createSyntheticImport",
        "createSyntheticImport"
    ],
    "src/compiler/utilities.ts": [
        "aggregateChildData",
        "aggregateChildData",
        "setNodeFlags"
    ]
};

// tsc writes flags on fresh factory nodes through bindings to Mutable results.
// binder.ts uses Mutable parameters, and the utilities allocators use this: Mutable<Node>.
// Pin each write with its function and binding so changes to either fail before any edit.
// Other Node.flags writes and unresolved or keyed flags writes refuse unless pinned.
const boundWrites = {
    "src/compiler/binder.ts": [
        { function: "bindContainer", statement: "node.flags &= ~(NodeFlags.ReachabilityAndEmitFlags | NodeFlags.ContainsThis);", binding: "node: Mutable<HasContainerFlags>" },
        { function: "bindContainer", statement: "node.flags |= NodeFlags.HasImplicitReturn;", binding: "node: Mutable<HasContainerFlags>" },
        { function: "bindContainer", statement: "node.flags |= NodeFlags.HasExplicitReturn;", binding: "node: Mutable<HasContainerFlags>" },
        { function: "bindContainer", statement: "node.flags |= NodeFlags.ContainsThis;", binding: "node: Mutable<HasContainerFlags>" },
        { function: "bindContainer", statement: "node.flags |= emitFlags;", binding: "node: Mutable<HasContainerFlags>" },
        { function: "bindContainer", statement: "node.flags = seenThisKeyword ? node.flags | NodeFlags.ContainsThis : node.flags & ~NodeFlags.ContainsThis;", binding: "node: Mutable<HasContainerFlags>" },
        { function: "setExportContextFlag", statement: "node.flags |= NodeFlags.ExportContext;", binding: "node: Mutable<ModuleDeclaration | SourceFile>" },
        { function: "setExportContextFlag", statement: "node.flags &= ~NodeFlags.ExportContext;", binding: "node: Mutable<ModuleDeclaration | SourceFile>" },
    ],
    "src/compiler/factory/nodeFactory.ts": [
        { function: "createIdentifier", statement: "node.flags |= NodeFlags.IdentifierHasExtendedUnicodeEscape;", binding: "createBaseIdentifier(escapeLeadingUnderscores(text))" },
        { function: "createPropertyAccessChain", statement: "node.flags |= NodeFlags.OptionalChain;", binding: "createBasePropertyAccessExpression(\r\n            parenthesizerRules().parenthesizeLeftSideOfAccess(expression, /*optionalChain*/ true),\r\n            questionDotToken,\r\n            asName(name),\r\n        ) as Mutable<PropertyAccessChain>" },
        { function: "createElementAccessChain", statement: "node.flags |= NodeFlags.OptionalChain;", binding: "createBaseElementAccessExpression(\r\n            parenthesizerRules().parenthesizeLeftSideOfAccess(expression, /*optionalChain*/ true),\r\n            questionDotToken,\r\n            asExpression(index),\r\n        ) as Mutable<ElementAccessChain>" },
        { function: "createCallChain", statement: "node.flags |= NodeFlags.OptionalChain;", binding: "createBaseCallExpression(\r\n            parenthesizerRules().parenthesizeLeftSideOfAccess(expression, /*optionalChain*/ true),\r\n            questionDotToken,\r\n            asNodeArray(typeArguments),\r\n            parenthesizerRules().parenthesizeExpressionsOfCommaDelimitedList(createNodeArray(argumentsArray)),\r\n        ) as Mutable<CallChain>" },
        { function: "createNonNullChain", statement: "node.flags |= NodeFlags.OptionalChain;", binding: "createBaseNode<NonNullChain>(SyntaxKind.NonNullExpression)" },
        { function: "createVariableDeclarationList", statement: "node.flags |= flags & NodeFlags.BlockScoped;", binding: "createBaseNode<VariableDeclarationList>(SyntaxKind.VariableDeclarationList)" },
        { function: "createModuleDeclaration", statement: "node.flags |= flags & (NodeFlags.Namespace | NodeFlags.NestedNamespace | NodeFlags.GlobalAugmentation);", binding: "createBaseDeclaration<ModuleDeclaration>(SyntaxKind.ModuleDeclaration)" },
        { function: "createSourceFile", statement: "node.flags |= flags;", binding: "baseFactory.createBaseSourceFileNode(SyntaxKind.SourceFile) as Mutable<SourceFile>" },
        { function: "cloneRedirectedSourceFile", statement: "node.flags |= source.flags & ~NodeFlags.Synthesized;", binding: "createRedirectedSourceFile(source.redirectInfo!) as Mutable<SourceFile>" },
        { function: "cloneSourceFileWorker", statement: "node.flags |= source.flags & ~NodeFlags.Synthesized;", binding: "baseFactory.createBaseSourceFileNode(SyntaxKind.SourceFile) as Mutable<SourceFile>" },
        { function: "cloneGeneratedIdentifier", statement: "clone.flags |= node.flags & ~NodeFlags.Synthesized;", binding: "createBaseIdentifier(node.escapedText) as Mutable<GeneratedIdentifier>" },
        { function: "cloneIdentifier", statement: "clone.flags |= node.flags & ~NodeFlags.Synthesized;", binding: "createBaseIdentifier(node.escapedText)" },
        { function: "cloneGeneratedPrivateIdentifier", statement: "clone.flags |= node.flags & ~NodeFlags.Synthesized;", binding: "createBasePrivateIdentifier(node.escapedText) as Mutable<GeneratedPrivateIdentifier>" },
        { function: "clonePrivateIdentifier", statement: "clone.flags |= node.flags & ~NodeFlags.Synthesized;", binding: "createBasePrivateIdentifier(node.escapedText)" },
    ],
    "src/compiler/utilities.ts": [
        { function: "Node", statement: "this.flags = NodeFlags.None;", binding: "this: Mutable<Node>" },
        { function: "Token", statement: "this.flags = NodeFlags.None;", binding: "this: Mutable<Node>" },
        { function: "Identifier", statement: "this.flags = NodeFlags.None;", binding: "this: Mutable<Node>" },
    ],
};
// The one element write whose constrained key could be "flags" on a Node receiver is pinned by text and function.
const keyedWrites = {
    "src/compiler/factory/nodeFactory.ts": [{ function: "cloneNode", statement: "clone[key] = node[key];" }],
};
const typesFile = "src/compiler/types.ts";
const readonlyFlags = "readonly flags: NodeFlags;";
const writableFlags = "flags: NodeFlags;";

function parse(tree, relative) {
    const file = path.join(tree, relative);
    const original = fs.readFileSync(file, "utf8");
    const source = ts.createSourceFile(file, original, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, relative);
    return { file, original, source, edits: [] };
}

function location(source, node) {
    const { line, character } = source.getLineAndCharacterOfPosition(node.getStart(source));
    return `${line + 1}:${character + 1}`;
}

function planDeclaration(plan) {
    const { source } = plan;
    const nodes = source.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === "Node");
    assert.equal(nodes.length, 1, "expected one Node interface");
    const members = nodes[0].members.filter(member => member.name?.getText(source) === "flags");
    assert.equal(members.length, 1, "expected one Node.flags member");
    const member = members[0];
    assert(ts.isPropertySignature(member), "Node.flags must be a property signature");
    const spelling = member.getText(source);
    assert(spelling === readonlyFlags || spelling === writableFlags, `unexpected Node.flags: ${spelling}`);
    if (spelling === writableFlags) return false;

    const start = member.getStart(source);
    plan.edits.push({ start, end: start + "readonly ".length, text: "", where: location(source, member) });
    return true;
}

function planWrites(plan, expected, pending) {
    const { source } = plan;
    const actual = [];
    function visit(node) {
        if (ts.isPropertyAccessExpression(node) && node.name.text === "flags" &&
            ts.isParenthesizedExpression(node.expression)) {
            const cast = node.expression.expression;
            if (ts.isAsExpression(cast) && ts.isTypeReferenceNode(cast.type) &&
                cast.type.typeName.getText(source) === "Mutable") {
                const write = node.parent;
                assert(ts.isBinaryExpression(write) && write.left === node &&
                    ts.isAssignmentOperator(write.operatorToken.kind) && ts.isExpressionStatement(write.parent),
                `Mutable flags cast does not own one write at ${location(source, node)}`);
                assert(ts.isIdentifier(cast.expression) || ts.isPropertyAccessExpression(cast.expression),
                    `unexpected cast receiver at ${location(source, node)}`);
                actual.push(write.parent.getText(source));
                plan.edits.push({
                    start: node.expression.getStart(source),
                    end: node.expression.end,
                    text: cast.expression.getText(source),
                    where: location(source, node),
                });
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.deepEqual(actual, pending ? expected : [], `unexpected Mutable flags writes in ${source.fileName}`);
}

function writePosition(node) {
    let target = node;
    for (let parent = target.parent; parent; target = parent, parent = target.parent) {
        if (ts.isBinaryExpression(parent)) {
            return parent.left === target && ts.isAssignmentOperator(parent.operatorToken.kind);
        }
        if (ts.isPrefixUnaryExpression(parent) || ts.isPostfixUnaryExpression(parent)) {
            return parent.operator === ts.SyntaxKind.PlusPlusToken || parent.operator === ts.SyntaxKind.MinusMinusToken;
        }
        if (ts.isForInStatement(parent) || ts.isForOfStatement(parent)) return parent.initializer === target;
        if ((ts.isPropertyAssignment(parent) && parent.initializer === target) ||
            ts.isArrayLiteralExpression(parent) || ts.isObjectLiteralExpression(parent) ||
            ts.isSpreadElement(parent) || ts.isSpreadAssignment(parent) ||
            ((ts.isParenthesizedExpression(parent) || ts.isNonNullExpression(parent) ||
                ts.isAsExpression(parent) || ts.isTypeAssertionExpression(parent) ||
                ts.isSatisfiesExpression(parent)) && parent.expression === target)) continue;
        return false;
    }
    return false;
}

function enclosing(node, predicate) {
    for (let parent = node.parent; parent; parent = parent.parent) {
        if (predicate(parent)) return parent;
    }
}

function executionOwner(node) {
    let child = node;
    for (let parent = child.parent; parent; child = parent, parent = child.parent) {
        if (ts.isFunctionLike(parent) || ts.isClassStaticBlockDeclaration(parent) ||
            (ts.isPropertyDeclaration(parent) && parent.initializer === child)) return parent;
    }
}

function receiverBinding(receiver, checker) {
    while (ts.isParenthesizedExpression(receiver) || ts.isNonNullExpression(receiver)) {
        receiver = receiver.expression;
    }
    if (receiver.kind === ts.SyntaxKind.ThisKeyword) {
        return enclosing(receiver, parent => ts.isFunctionLike(parent) && !ts.isArrowFunction(parent))
            ?.parameters.find(parameter => parameter.name.getText() === "this");
    }
    if (ts.isIdentifier(receiver)) return checker.getSymbolAtLocation(receiver)?.valueDeclaration;
}

function bindingMatches(pin, statement, declaration, owner) {
    if (!owner || !ts.isFunctionDeclaration(owner) || !declaration || executionOwner(declaration) !== owner) return false;
    const binding = ts.isVariableDeclaration(declaration) ? declaration.initializer :
        ts.isParameter(declaration) ? declaration : undefined;
    return pin.statement === statement && pin.function === owner.name?.text && pin.binding === binding?.getText();
}

function castFree(statement) {
    const source = ts.createSourceFile("pin.ts", statement, ts.ScriptTarget.Latest, true);
    const receiver = source.statements[0].expression.left.expression;
    const cast = receiver.expression;
    return statement.slice(0, receiver.getStart(source)) + cast.expression.getText(source) + statement.slice(receiver.end);
}

const constrainedKinds = ts.TypeFlags.TypeParameter | ts.TypeFlags.IndexedAccess | ts.TypeFlags.Conditional |
    ts.TypeFlags.Index | ts.TypeFlags.Substitution;
function someReachableType(checker, type, predicate, seen = new Set()) {
    if (seen.has(type)) return false;
    seen.add(type);
    if (predicate(type)) return true;
    if (type.isUnionOrIntersection() && type.types.some(part => someReachableType(checker, part, predicate, seen))) return true;
    const base = type.flags & constrainedKinds ? checker.getBaseConstraintOfType(type) : undefined;
    return base !== undefined && someReachableType(checker, base, predicate, seen);
}
function potentiallyFlags(checker, type) {
    return someReachableType(checker, type, key => !!(key.flags & (ts.TypeFlags.Any | ts.TypeFlags.String)) ||
        (key.isStringLiteral() && key.value === "flags") || checker.isTypeAssignableTo(checker.getStringLiteralType("flags"), key));
}

function checkWriteScope(tree) {
    const { compiler } = require("../47-host-errors/adapt.cjs");
    const roots = ts.sys.readDirectory(path.join(tree, "src"), [".ts"]).sort();
    const program = ts.createProgram(roots, {
        ...compiler(tree, ts).getCompilerOptions(),
        typeRoots: [path.resolve(__dirname, "../../api/node_modules/@types")],
        skipLibCheck: true,
        noEmit: true,
    });
    const checker = program.getTypeChecker();
    const source = program.getSourceFile(path.join(tree, typesFile));
    const nodes = source.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === "Node");
    assert.equal(nodes.length, 1, "expected one Node interface");
    const members = nodes[0].members.filter(node => node.name?.getText(source) === "flags");
    assert.equal(members.length, 1, "expected one Node.flags declaration");
    const [member] = members;
    const flags = checker.getSymbolAtLocation(member.name);
    assert(flags, "missing Node.flags symbol");
    const spelling = member.getText(source);
    assert(spelling === readonlyFlags || spelling === writableFlags, `unexpected Node.flags: ${spelling}`);

    function resolvesToFlags(symbol) {
        return symbol && (symbol === flags || checker.getRootSymbols(symbol).includes(flags) ||
            symbol.declarations?.includes(member));
    }
    function hasNodeFlags(type) {
        return someReachableType(checker, type, part => !!resolvesToFlags(checker.getPropertyOfType(part, "flags")));
    }

    for (const root of roots) {
        const file = program.getSourceFile(root);
        const relative = path.relative(tree, root).split(path.sep).join("/");
        const direct = (writes[relative] ?? []).map((text, index) => ({
            statement: spelling === readonlyFlags ? text : castFree(text),
            function: directFunctions[relative][index],
        }));
        const bound = [...(boundWrites[relative] ?? [])];
        const keyed = [...(keyedWrites[relative] ?? [])];

        function check(node, key, argument) {
            const where = `${relative}:${location(file, node)}`;
            const receiverType = checker.getTypeAtLocation(node.expression);
            const owner = executionOwner(node);
            const ownerName = owner && ts.isFunctionDeclaration(owner) ? owner.name?.text : undefined;
            const expression = node.parent;
            const statement = ts.isExpressionStatement(expression.parent) &&
                expression.parent.expression === expression ? expression.parent.getText(file) : undefined;

            if (!key) {
                if (!potentiallyFlags(checker, checker.getTypeAtLocation(argument)) || !hasNodeFlags(receiverType)) return;
                const index = keyed.findIndex(pin => pin.statement === statement && pin.function === ownerName);
                assert(index !== -1, `unexpected keyed flags write at ${where}`);
                keyed.splice(index, 1);
                return;
            }

            const symbol = checker.getSymbolAtLocation(key);
            assert(symbol && !(receiverType.flags & ts.TypeFlags.Any), `unexpected unresolved flags write at ${where}`);
            if (!resolvesToFlags(symbol)) return;
            const declaration = receiverBinding(node.expression, checker);
            const boundIndex = bound.findIndex(pin => bindingMatches(pin, statement, declaration, owner));
            if (boundIndex !== -1) {
                bound.splice(boundIndex, 1);
                return;
            }
            const directIndex = direct.findIndex(pin => pin.statement === statement && pin.function === ownerName);
            assert(directIndex !== -1, `unexpected Node flags write at ${where}`);
            direct.splice(directIndex, 1);
        }

        function visit(node) {
            if (ts.isPropertyAccessExpression(node) && node.name.text === "flags" && writePosition(node)) {
                check(node, node.name);
            }
            if (ts.isElementAccessExpression(node) && writePosition(node)) {
                const argument = node.argumentExpression;
                if (!ts.isStringLiteralLike(argument)) check(node, undefined, argument);
                else if (argument.text === "flags") check(node, argument);
            }
            ts.forEachChild(node, visit);
        }
        visit(file);
        assert(direct.length + bound.length + keyed.length === 0, `missing pinned Node flags write in ${relative}`);
    }
}

function adapt(tree) {
    checkWriteScope(tree);
    const declaration = parse(tree, typesFile);
    const pending = planDeclaration(declaration);
    const plans = [[typesFile, declaration]];
    for (const [relative, expected] of Object.entries(writes)) {
        const plan = parse(tree, relative);
        planWrites(plan, expected, pending);
        plans.push([relative, plan]);
    }
    for (const [, plan] of plans) {
        assert.equal(fs.readFileSync(plan.file, "utf8"), plan.original, `source changed while planning: ${plan.file}`);
    }
    for (const [, plan] of plans) {
        let changed = plan.original;
        for (const edit of [...plan.edits].reverse()) {
            changed = changed.slice(0, edit.start) + edit.text + changed.slice(edit.end);
        }
        if (changed !== plan.original) fs.writeFileSync(plan.file, changed);
    }
    return {
        declarationEdits: declaration.edits.length,
        castEdits: plans.slice(1).reduce((count, [, plan]) => count + plan.edits.length, 0),
        sites: plans.flatMap(([file, plan]) => plan.edits.map(edit => `${file}:${edit.where}`)),
    };
}

assert.equal(ts.version, "6.0.3");
assert.deepEqual(Object.keys(directFunctions).sort(), Object.keys(writes).sort());
for (const [file, statements] of Object.entries(writes)) {
    assert.equal(directFunctions[file].length, statements.length, file);
    statements.forEach((statement, index) => {
        const owner = directFunctions[file][index];
        const replay = castFree(statement);
        assert(!(boundWrites[file] ?? []).some(pin =>
            pin.function === owner && (pin.statement === statement || pin.statement === replay)),
        `overlapping direct and bound pins in ${file}:${owner}`);
    });
}
assert.equal(process.argv.length, 3, "usage: node adapt.cjs TREE");
console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
