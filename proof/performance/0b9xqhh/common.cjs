"use strict";
// Shared by select.cjs and instrument.cjs: both must find the same expression at a site and resolve the same member
// names on its target, since a member with no declaration (a mapped type's property) is matched by symbol identity,
// which only holds inside one program, so instrument.cjs resolves the sites again rather than trusting a saved name.
const fs = require("node:fs");
const path = require("node:path");

// The pinned API is stage3/api's install; the job runs at the adamic root, so the default is relative to it.
function loadTypeScript() {
    const where = process.env.P72_TYPESCRIPT || path.resolve("stage3/api/node_modules/typescript");
    const ts = require(where);
    return ts;
}

function flatten(ts, diagnostic) {
    return ts.flattenDiagnosticMessageText(diagnostic.messageText, "\n");
}

function loadProgram(ts, tree) {
    const configPath = path.join(tree, "src/compiler/tsconfig.json");
    const read = ts.readConfigFile(configPath, ts.sys.readFile);
    if (read.error) throw new Error(`${configPath}: ${flatten(ts, read.error)}`);
    const parsed = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
    if (parsed.errors.length) throw new Error(`${configPath}: ${parsed.errors.map(d => flatten(ts, d)).join("; ")}`);
    const program = ts.createProgram(parsed.fileNames, parsed.options);
    return { program, checker: program.getTypeChecker() };
}

// A declaration's id is where its node starts, so select and instrument agree without sharing symbols.
function declarationId(ts, tree, declaration, name) {
    const sourceFile = declaration.getSourceFile();
    const where = sourceFile.getLineAndCharacterOfPosition(declaration.getStart(sourceFile));
    return `${path.relative(tree, sourceFile.fileName)}:${where.line + 1}:${where.character + 1} ${name}`;
}

// The census's column is read as UTF-16 code units first, as the TypeScript API counts; a byte column is tried after,
// for a line with non-ASCII text before the site, in case the census counted bytes.
function candidatePositions(sourceFile, line, column) {
    const starts = sourceFile.getLineStarts();
    if (!(line >= 1 && line <= starts.length) || !(column >= 1)) return [];
    const start = starts[line - 1];
    const end = line < starts.length ? starts[line] : sourceFile.text.length;
    const text = sourceFile.text.slice(start, end);
    const found = [];
    if (column - 1 < text.length) found.push({ position: start + column - 1, unit: "char" });
    const bytes = Buffer.from(text, "utf8");
    if (column - 1 < bytes.length) {
        const characters = bytes.subarray(0, column - 1).toString("utf8").length;
        if (characters !== column - 1) found.push({ position: start + characters, unit: "byte" });
    }
    return found;
}

// Every expression starting at position, outermost first.
function expressionsStartingAt(ts, sourceFile, position) {
    const found = [];
    (function visit(node) {
        if (position < node.pos || position >= node.end) return;
        if (ts.isExpression(node) && node.getStart(sourceFile) === position) found.push(node);
        ts.forEachChild(node, visit);
    })(sourceFile);
    return found;
}

// The relation's expression: the outermost expression starting at the site, since the innermost one starting there is
// only the head of an argument like `a.b` (its `a`). A site on an assignment's left side means its right side.
function relationExpression(ts, sourceFile, position) {
    const found = expressionsStartingAt(ts, sourceFile, position);
    if (!found.length) return relationOfStatement(ts, sourceFile, position);
    let expression = found[0];
    if (ts.isBinaryExpression(expression) && expression.operatorToken.kind >= ts.SyntaxKind.FirstAssignment &&
        expression.operatorToken.kind <= ts.SyntaxKind.LastAssignment) {
        expression = expression.right;
    }
    return expression;
}

// A site on the relation's own node rather than its expression (`return`, a declaration's name, a property's name)
// means that node's expression or initializer.
function relationOfStatement(ts, sourceFile, position) {
    let found;
    (function visit(node) {
        if (found || position < node.pos || position >= node.end) return;
        if (node.getStart(sourceFile) === position) {
            if (ts.isReturnStatement(node) && node.expression) found = node.expression;
            else if ((ts.isVariableDeclaration(node) || ts.isPropertyAssignment(node)) && node.initializer) found = node.initializer;
            if (found) return;
        }
        ts.forEachChild(node, visit);
    })(sourceFile);
    return found;
}

// The target with `undefined` and `null` taken out, as the census's `T | undefined` widening does, and an array
// target's element when the site's member list came from its element (nodeFactory.ts:7489).
function siteTarget(ts, checker, expression, elementOfArray) {
    const contextual = checker.getContextualType(expression);
    if (!contextual) return undefined;
    let target = checker.getNonNullableType(contextual);
    if (elementOfArray) {
        const element = checker.isArrayType(target) ? checker.getTypeArguments(target)[0] : target.getNumberIndexType();
        if (!element) return undefined;
        target = checker.getNonNullableType(element);
    }
    return target;
}

// name on each member type of the target that declares it (the target itself when it isn't a union).
function resolveName(ts, checker, target, name) {
    const owners = target.isUnion() ? target.types : [target];
    const found = [];
    for (const owner of owners) {
        const symbol = checker.getPropertyOfType(owner, name);
        if (symbol) found.push({ symbol, owner });
    }
    return found;
}

// A member's ids: one per declaration, or one naming its owner when it has none.
function memberIds(ts, checker, tree, resolved, name) {
    const declarations = resolved.symbol.declarations ?? [];
    if (declarations.length) return declarations.map(d => ({ id: declarationId(ts, tree, d, name), declaration: d }));
    const owner = checker.typeToString(resolved.owner, undefined, ts.TypeFormatFlags.NoTruncation).slice(0, 160);
    return [{ id: `(no declaration) ${name} on ${owner}`, declaration: undefined }];
}

function readTsv(file) {
    const lines = fs.readFileSync(file, "utf8").split("\n").map(l => l.replace(/\r$/, "")).filter(l => l.length);
    const [header, ...rows] = lines;
    return { header: header.split("\t"), rows: rows.map(l => l.split("\t")) };
}

function csv(rows, columns) {
    const cell = v => {
        const s = String(v ?? "");
        return /[",\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s;
    };
    return [columns.join(","), ...rows.map(r => columns.map(c => cell(r[c])).join(","))].join("\n") + "\n";
}

module.exports = {
    loadTypeScript,
    loadProgram,
    declarationId,
    candidatePositions,
    relationExpression,
    siteTarget,
    resolveName,
    memberIds,
    readTsv,
    csv,
};
