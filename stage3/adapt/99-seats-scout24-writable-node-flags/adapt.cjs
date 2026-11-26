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

const directPaths = {
    "src/compiler/binder.ts": [
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","isPotentiallyExecutableNode(node)"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","isPotentiallyExecutableNode(node)"],["Block","statements"],["IfStatement","thenStatement","currentFlow === unreachableFlow"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","!activeLabelList.referenced"],["Block","statements"],["FunctionDeclaration","body"]],
    ],
    "src/compiler/factory/nodeFactory.ts": [
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]],
    ],
    "src/compiler/parser.ts": [
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","hasDeprecatedTag"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","contextFlags"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","parseErrorBeforeNextFinishedNode"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","isJSDocNullableType(type) && type.pos === type.type.pos"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["WhileStatement","statement","isNonNullExpression(node)"],["Block","statements"],["IfStatement","thenStatement","expr.flags & NodeFlags.OptionalChain"],["Block","statements"],["IfStatement","thenStatement","isNonNullExpression(node)"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","questionDotToken || tag.flags & NodeFlags.OptionalChain"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["ForOfStatement","statement","const m","modifiers!",false],["Block","statements"],["IfStatement","thenStatement","isAmbient"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["ForOfStatement","statement","const m","modifiers!",false],["Block","statements"],["IfStatement","thenStatement","isAmbient"],["Block","statements"],["IfStatement","thenStatement","tokenIsIdentifierOrKeyword(token()) ||\r\n            token() === SyntaxKind.StringLiteral ||\r\n            token() === SyntaxKind.NumericLiteral ||\r\n            token() === SyntaxKind.BigIntLiteral ||\r\n            token() === SyntaxKind.AsteriskToken ||\r\n            token() === SyntaxKind.OpenBracketToken"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","nested"],["Block","statements"],["FunctionDeclaration","body"]],
    ],
    "src/compiler/program.ts": [
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]],
    ],
    "src/compiler/utilities.ts": [
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","thisNodeOrAnySubNodesHasError"],["Block","statements"],["IfStatement","thenStatement","!(node.flags & NodeFlags.HasAggregatedChildData)"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","!(node.flags & NodeFlags.HasAggregatedChildData)"],["Block","statements"],["FunctionDeclaration","body"]],
        [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","node"],["Block","statements"],["FunctionDeclaration","body"]],
    ],
};

// tsc writes flags on fresh factory nodes through bindings to Mutable results.
// binder.ts uses Mutable parameters, and the utilities allocators use this: Mutable<Node>.
// Pin each write with its function and binding so changes to either fail before any edit.
// Other Node.flags writes and unresolved or keyed flags writes refuse unless pinned.
const boundWrites = {
    "src/compiler/binder.ts": [
        { function: "bindContainer", statement: "node.flags &= ~(NodeFlags.ReachabilityAndEmitFlags | NodeFlags.ContainsThis);", binding: "node: Mutable<HasContainerFlags>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","containerFlags & ContainerFlags.IsControlFlowContainer"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "bindContainer", statement: "node.flags |= NodeFlags.HasImplicitReturn;", binding: "node: Mutable<HasContainerFlags>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","!(currentFlow.flags & FlowFlags.Unreachable) && containerFlags & ContainerFlags.IsFunctionLike && nodeIsPresent((node as FunctionLikeDeclaration | ClassStaticBlockDeclaration).body)"],["Block","statements"],["IfStatement","thenStatement","containerFlags & ContainerFlags.IsControlFlowContainer"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "bindContainer", statement: "node.flags |= NodeFlags.HasExplicitReturn;", binding: "node: Mutable<HasContainerFlags>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["IfStatement","thenStatement","hasExplicitReturn"],["Block","statements"],["IfStatement","thenStatement","!(currentFlow.flags & FlowFlags.Unreachable) && containerFlags & ContainerFlags.IsFunctionLike && nodeIsPresent((node as FunctionLikeDeclaration | ClassStaticBlockDeclaration).body)"],["Block","statements"],["IfStatement","thenStatement","containerFlags & ContainerFlags.IsControlFlowContainer"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "bindContainer", statement: "node.flags |= NodeFlags.ContainsThis;", binding: "node: Mutable<HasContainerFlags>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","seenThisKeyword"],["Block","statements"],["IfStatement","thenStatement","containerFlags & ContainerFlags.IsControlFlowContainer"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "bindContainer", statement: "node.flags |= emitFlags;", binding: "node: Mutable<HasContainerFlags>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","node.kind === SyntaxKind.SourceFile"],["Block","statements"],["IfStatement","thenStatement","containerFlags & ContainerFlags.IsControlFlowContainer"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "bindContainer", statement: "node.flags = seenThisKeyword ? node.flags | NodeFlags.ContainsThis : node.flags & ~NodeFlags.ContainsThis;", binding: "node: Mutable<HasContainerFlags>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","containerFlags & ContainerFlags.IsInterface"],["IfStatement","elseStatement","containerFlags & ContainerFlags.IsControlFlowContainer"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "setExportContextFlag", statement: "node.flags |= NodeFlags.ExportContext;", binding: "node: Mutable<ModuleDeclaration | SourceFile>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","thenStatement","node.flags & NodeFlags.Ambient && !hasExportDeclarations(node)"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "setExportContextFlag", statement: "node.flags &= ~NodeFlags.ExportContext;", binding: "node: Mutable<ModuleDeclaration | SourceFile>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["IfStatement","elseStatement","node.flags & NodeFlags.Ambient && !hasExportDeclarations(node)"],["Block","statements"],["FunctionDeclaration","body"]] },
    ],
    "src/compiler/factory/nodeFactory.ts": [
        { function: "createIdentifier", statement: "node.flags |= NodeFlags.IdentifierHasExtendedUnicodeEscape;", binding: "createBaseIdentifier(escapeLeadingUnderscores(text))", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["IfStatement","thenStatement","hasExtendedUnicodeEscape"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "createPropertyAccessChain", statement: "node.flags |= NodeFlags.OptionalChain;", binding: "createBasePropertyAccessExpression(\r\n            parenthesizerRules().parenthesizeLeftSideOfAccess(expression, /*optionalChain*/ true),\r\n            questionDotToken,\r\n            asName(name),\r\n        ) as Mutable<PropertyAccessChain>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "createElementAccessChain", statement: "node.flags |= NodeFlags.OptionalChain;", binding: "createBaseElementAccessExpression(\r\n            parenthesizerRules().parenthesizeLeftSideOfAccess(expression, /*optionalChain*/ true),\r\n            questionDotToken,\r\n            asExpression(index),\r\n        ) as Mutable<ElementAccessChain>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "createCallChain", statement: "node.flags |= NodeFlags.OptionalChain;", binding: "createBaseCallExpression(\r\n            parenthesizerRules().parenthesizeLeftSideOfAccess(expression, /*optionalChain*/ true),\r\n            questionDotToken,\r\n            asNodeArray(typeArguments),\r\n            parenthesizerRules().parenthesizeExpressionsOfCommaDelimitedList(createNodeArray(argumentsArray)),\r\n        ) as Mutable<CallChain>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "createNonNullChain", statement: "node.flags |= NodeFlags.OptionalChain;", binding: "createBaseNode<NonNullChain>(SyntaxKind.NonNullExpression)", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "createVariableDeclarationList", statement: "node.flags |= flags & NodeFlags.BlockScoped;", binding: "createBaseNode<VariableDeclarationList>(SyntaxKind.VariableDeclarationList)", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "createModuleDeclaration", statement: "node.flags |= flags & (NodeFlags.Namespace | NodeFlags.NestedNamespace | NodeFlags.GlobalAugmentation);", binding: "createBaseDeclaration<ModuleDeclaration>(SyntaxKind.ModuleDeclaration)", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "createSourceFile", statement: "node.flags |= flags;", binding: "baseFactory.createBaseSourceFileNode(SyntaxKind.SourceFile) as Mutable<SourceFile>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "cloneRedirectedSourceFile", statement: "node.flags |= source.flags & ~NodeFlags.Synthesized;", binding: "createRedirectedSourceFile(source.redirectInfo!) as Mutable<SourceFile>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "cloneSourceFileWorker", statement: "node.flags |= source.flags & ~NodeFlags.Synthesized;", binding: "baseFactory.createBaseSourceFileNode(SyntaxKind.SourceFile) as Mutable<SourceFile>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "cloneGeneratedIdentifier", statement: "clone.flags |= node.flags & ~NodeFlags.Synthesized;", binding: "createBaseIdentifier(node.escapedText) as Mutable<GeneratedIdentifier>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "cloneIdentifier", statement: "clone.flags |= node.flags & ~NodeFlags.Synthesized;", binding: "createBaseIdentifier(node.escapedText)", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "cloneGeneratedPrivateIdentifier", statement: "clone.flags |= node.flags & ~NodeFlags.Synthesized;", binding: "createBasePrivateIdentifier(node.escapedText) as Mutable<GeneratedPrivateIdentifier>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "clonePrivateIdentifier", statement: "clone.flags |= node.flags & ~NodeFlags.Synthesized;", binding: "createBasePrivateIdentifier(node.escapedText)", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
    ],
    "src/compiler/utilities.ts": [
        { function: "Node", statement: "this.flags = NodeFlags.None;", binding: "this: Mutable<Node>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "Token", statement: "this.flags = NodeFlags.None;", binding: "this: Mutable<Node>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
        { function: "Identifier", statement: "this.flags = NodeFlags.None;", binding: "this: Mutable<Node>", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["FunctionDeclaration","body"]] },
    ],
};
// The one element write whose constrained key could be "flags" on a Node receiver is pinned by text and function.
const keyedWrites = {
    "src/compiler/factory/nodeFactory.ts": [{ function: "cloneNode", statement: "clone[key] = node[key];", path: [["BinaryExpression","left"],["ExpressionStatement","expression"],["Block","statements"],["ForInStatement","statement","const key","node",false],["Block","statements"],["FunctionDeclaration","body"]] }],
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

function ownerPath(node) {
    const path = [];
    for (let child = node, parent = child.parent; parent; child = parent, parent = child.parent) {
        const slot = Object.keys(parent).find(key => {
            const value = parent[key];
            return key !== "parent" && (value === child || (Array.isArray(value) && value.includes(child)));
        });
        const text = item => item?.getText() ?? "";
        const step = [ts.SyntaxKind[parent.kind], slot ?? "<unknown>"];
        if (ts.isIfStatement(parent) || ts.isWhileStatement(parent) || ts.isDoStatement(parent) ||
            ts.isSwitchStatement(parent) || ts.isCaseClause(parent)) {
            step.push(text(parent.expression));
        } else if (ts.isForStatement(parent)) {
            step.push(text(parent.initializer), text(parent.condition), text(parent.incrementor));
        } else if (ts.isForInStatement(parent) || ts.isForOfStatement(parent)) {
            step.push(text(parent.initializer), text(parent.expression), Boolean(parent.awaitModifier));
        } else if (ts.isCatchClause(parent)) {
            step.push(text(parent.variableDeclaration));
        }
        path.push(step);
        if (ts.isFunctionLike(parent)) return { owner: parent, path };
    }
    return { owner: undefined, path };
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
    if (!owner || !ts.isFunctionDeclaration(owner) || !declaration || ownerPath(declaration).owner !== owner) return false;
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

function ownerKey(node, source) {
    const key = [];
    for (let current = node; current && current !== source; current = current.parent) {
        const parent = current.parent;
        const slot = Object.keys(parent).find(key => {
            const value = parent[key];
            return key !== "parent" && (value === current || (Array.isArray(value) && value.includes(current)));
        });
        key.push([ts.SyntaxKind[current.kind], slot ?? "",
            typeof current.name?.getText === "function" ? current.name.getText(source) : ""]);
    }
    return key.reverse();
}

function checkWriteScope(tree) {
    const { compiler } = require("../47-host-errors/adapt.cjs");
    const suffixes = [".ts", ".tsx", ".mts", ".cts", ".d.ts", ".d.mts", ".d.cts"];
    const roots = ts.sys.readDirectory(path.join(tree, "src"), suffixes).sort();
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

    for (const root of roots) assert(program.getSourceFile(root), `missing program root ${root}`);
    const files = program.getSourceFiles()
        .map(file => ({ file, relative: path.relative(tree, file.fileName).split(path.sep).join("/") }))
        .filter(({ relative }) => relative.startsWith("src/"))
        .sort((left, right) => left.relative < right.relative ? -1 : left.relative > right.relative ? 1 : 0);
    for (const { file, relative } of files) {
        const direct = (writes[relative] ?? []).map((text, index) => ({
            statement: spelling === readonlyFlags ? text : castFree(text),
            function: directFunctions[relative][index],
            path: directPaths[relative][index],
        }));
        const bound = [...(boundWrites[relative] ?? [])];
        const keyed = [...(keyedWrites[relative] ?? [])];

        const matchedOwners = new Map();
        function remember(pin, owner, where, currentPath, kind) {
            const entry = owners.get(relative)?.get(pin.function);
            const previous = matchedOwners.get(entry);
            if (previous && previous.owner !== owner) {
                assert.fail(`unexpected ${kind} flags write at ${where}
Owner ${relative}:${pin.function} names two nodes
Pinned key: ${JSON.stringify(entry.key)}
Write's owner key: ${JSON.stringify(ownerKey(owner, file))}`);
            }
            if (!previous) matchedOwners.set(entry, { owner, where, currentPath, pin, kind });
        }

        function check(node, key, argument) {
            const where = `${relative}:${location(file, node)}`;
            const receiverType = checker.getTypeAtLocation(node.expression);
            const { owner, path: currentPath } = ownerPath(node);
            const ownerName = owner && ts.isFunctionDeclaration(owner) ? owner.name?.text : undefined;
            const expression = node.parent;
            const statement = ts.isExpressionStatement(expression.parent) &&
                expression.parent.expression === expression ? expression.parent.getText(file) : undefined;

            if (!key) {
                if (!potentiallyFlags(checker, checker.getTypeAtLocation(argument)) || !hasNodeFlags(receiverType)) return;
                const index = keyed.findIndex(pin => pin.statement === statement && pin.function === ownerName);
                assert(index !== -1, `unexpected keyed flags write at ${where}`);
                remember(keyed[index], owner, where, currentPath, "keyed");
                keyed.splice(index, 1);
                return;
            }

            const symbol = checker.getSymbolAtLocation(key);
            assert(symbol && !(receiverType.flags & ts.TypeFlags.Any), `unexpected unresolved flags write at ${where}`);
            if (!resolvesToFlags(symbol)) return;
            const declaration = receiverBinding(node.expression, checker);
            const boundIndex = bound.findIndex(pin => bindingMatches(pin, statement, declaration, owner));
            if (boundIndex !== -1) {
                remember(bound[boundIndex], owner, where, currentPath, "Node");
                bound.splice(boundIndex, 1);
                return;
            }
            const directIndex = direct.findIndex(pin => pin.statement === statement && pin.function === ownerName);
            assert(directIndex !== -1, `unexpected Node flags write at ${where}`);
            remember(direct[directIndex], owner, where, currentPath, "Node");
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
        const entries = owners.get(relative);
        if (entries) {
            const byKey = new Map();
            function indexOwners(node) {
                if (ts.isFunctionLike(node) && node.body && entries.has(node.name?.getText(file))) {
                    const key = JSON.stringify(ownerKey(node, file));
                    if (!byKey.has(key)) byKey.set(key, []);
                    byKey.get(key).push(node);
                }
                ts.forEachChild(node, indexOwners);
            }
            indexOwners(file);
            const place = node => `${relative}:${location(file, node)}`;
            for (const entry of entries.values()) {
                const key = JSON.stringify(entry.key);
                const nodes = byKey.get(key) ?? [];
                const match = matchedOwners.get(entry);
                const prefix = `unexpected ${match.kind} flags write at ${match.where}\nOwner ${relative}:${entry.function}`;
                const paths = () => `\nPinned key: ${JSON.stringify(entry.key)}\nWrite's owner key: ${JSON.stringify(ownerKey(match.owner, file))}`;
                if (nodes.length === 0) assert.fail(`${prefix} names no node${paths()}`);
                if (nodes.length !== 1) assert.fail(`${prefix} names ${nodes.length} nodes at ${nodes.map(place).join(", ")}${paths()}`);
                if (nodes[0] !== match.owner) {
                    assert.fail(`${prefix} is the node at ${place(nodes[0])}, not the write's owner at ${place(match.owner)}${paths()}`);
                }
            }
        }
        for (const [entry, match] of matchedOwners) {
            const expected = spelling === readonlyFlags ? entry.before : entry.after;
            const actual = match.owner.getText(file);
            if (actual === expected) continue;
            let offset = 0;
            while (offset < actual.length && offset < expected.length && actual[offset] === expected[offset]) offset++;
            const point = file.getLineAndCharacterOfPosition(match.owner.getStart(file) + offset);
            assert.fail(`unexpected ${match.kind} flags write at ${match.where}
` +
                `Owner ${relative}:${entry.function} first differs at ${relative}:${point.line + 1}:${point.character + 1}
` +
                `Current path: ${JSON.stringify(match.currentPath)}
Pinned path: ${JSON.stringify(match.pin.path)}`);
        }
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
const owners = new Map();
for (const entry of require("./owners.json")) {
    if (!owners.has(entry.file)) owners.set(entry.file, new Map());
    const functions = owners.get(entry.file);
    assert(!functions.has(entry.function), `duplicate owner ${entry.file}:${entry.function}`);
    assert.equal(typeof entry.before, "string");
    assert.equal(typeof entry.after, "string");
    assert(Array.isArray(entry.key) && entry.key.length > 0 && entry.key.every(step =>
        Array.isArray(step) && step.length === 3 && step.every(part => typeof part === "string")),
        `invalid owner key ${entry.file}:${entry.function}`);
    assert.equal(entry.key[0][1], "statements");
    assert.equal(entry.key.at(-1)[2], entry.function);
    assert(![...functions.values()].some(other => JSON.stringify(other.key) === JSON.stringify(entry.key)),
        `duplicate owner key ${entry.file}:${entry.function}`);
    functions.set(entry.function, entry);
}
const requiredOwners = new Map();
for (const file of new Set([...Object.keys(writes), ...Object.keys(boundWrites), ...Object.keys(keyedWrites)])) {
    const names = new Set([...(directFunctions[file] ?? []),
        ...(boundWrites[file] ?? []).map(pin => pin.function), ...(keyedWrites[file] ?? []).map(pin => pin.function)]);
    requiredOwners.set(file, names);
    assert.deepEqual(new Set(owners.get(file)?.keys()), names, file);
    for (const name of names) {
        const { before, after } = owners.get(file).get(name);
        const direct = (writes[file] ?? []).filter((_, index) => directFunctions[file][index] === name);
        const source = ts.createSourceFile(file, before, ts.ScriptTarget.Latest, true);
        const statements = [];
        function collect(node) {
            if (ts.isExpressionStatement(node) && direct.includes(node.getText(source))) statements.push(node);
            ts.forEachChild(node, collect);
        }
        collect(source);
        assert.deepEqual(statements.map(node => node.getText(source)), direct, `direct statements for owner ${file}:${name}`);
        let expected = before;
        for (const node of statements.reverse()) {
            expected = expected.slice(0, node.getStart(source)) + castFree(node.getText(source)) + expected.slice(node.end);
        }
        assert(expected === after, `incorrect replay text for owner ${file}:${name}`);
    }
}
assert.deepEqual(new Set(owners.keys()), new Set(requiredOwners.keys()));
assert.equal(process.argv.length, 3, "usage: node adapt.cjs TREE");
console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
