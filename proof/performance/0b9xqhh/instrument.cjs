"use strict";
// Usage: node instrument.cjs <tsTree> <members.json> <copyDir>
// Copies <tsTree>/src to <copyDir>/src and, in every file of the src/compiler program, makes each execution of an
// access to a selected member bump a per-member counter: read, write, compound (read then write), and `in` checks on
// the member's name apart. The counters live on globalThis.__p72; the copy's tsc.ts writes them to $P72_COUNTS at
// exit. Writes instrumented.json (the counter index and static sites) and accesses.csv beside members.json.
//
// The edit mechanics are slice 4's (evidence-slice4/scripts/select-and-instrument.cjs at 420072fe58): text inserted
// into the original source, never reprinted, so nothing but the counters changes. Two things differ. A property read
// `r.m` becomes `r[(count++, "m")]` rather than `(count++, r).m`, so an optional chain still short-circuits before the
// count and a read that never happens is never counted. And edits at one position are ordered by nesting, closes
// before opens, inner closes first and outer opens first, where slice 4's stable sort could nest them backwards.
const fs = require("node:fs");
const path = require("node:path");
const common = require("./common.cjs");

const [tree, membersPath, copyDir] = process.argv.slice(2).map(a => a && path.resolve(a));
if (!tree || !membersPath || !copyDir) {
    console.error("usage: node instrument.cjs <tsTree> <members.json> <copyDir>");
    process.exit(2);
}
if (fs.existsSync(path.join(copyDir, "src"))) {
    console.error(`${copyDir}/src is already there; instrument writes a fresh copy`);
    process.exit(2);
}

const ts = common.loadTypeScript();
const SK = ts.SyntaxKind;
const { program, checker } = common.loadProgram(ts, tree);
const selection = JSON.parse(fs.readFileSync(membersPath, "utf8"));
const members = selection.members.map(id => ({ id, ...selection.declarations[id] }));
const indexById = new Map(members.map((m, i) => [m.id, i]));
const trackedNames = new Set(members.map(m => m.name));

// A member with no declaration is only reachable as the symbol the site's target resolved, so the sites are resolved
// again here, in this program, and each such symbol remembered.
const indexBySymbol = new Map();
for (const site of selection.sites) {
    const sourceFile = program.getSourceFile(path.join(tree, site.file));
    const expression = sourceFile && common.relationExpression(ts, sourceFile, site.position);
    const target = expression && common.siteTarget(ts, checker, expression, site.elementOfArray);
    if (!target) {
        console.log(`warning: site ${site.site} no longer resolves in this program`);
        continue;
    }
    for (const name of site.names) {
        for (const resolved of common.resolveName(ts, checker, target, name)) {
            for (const { id } of common.memberIds(ts, checker, tree, resolved, name)) {
                if (indexById.has(id) && !indexBySymbol.has(resolved.symbol)) indexBySymbol.set(resolved.symbol, indexById.get(id));
            }
        }
    }
}

function memberOfSymbol(symbol) {
    if (!symbol) return -1;
    if (indexBySymbol.has(symbol)) return indexBySymbol.get(symbol);
    for (const declaration of symbol.declarations ?? []) {
        const index = indexById.get(common.declarationId(ts, tree, declaration, ts.symbolName(symbol)));
        if (index !== undefined) return index;
    }
    return -1;
}

// name's property symbols on type and on each of its union or intersection members.
function propertiesOf(type, name) {
    const found = [];
    const own = checker.getNonNullableType(type);
    const whole = checker.getPropertyOfType(own, name);
    if (whole) found.push(whole);
    if (own.isUnionOrIntersection()) {
        for (const part of own.types) {
            const symbol = checker.getPropertyOfType(part, name);
            if (symbol) found.push(symbol);
        }
    }
    return found;
}

// The member an access reaches. A union's property with no declarations is a fresh symbol, so for one of those the
// receiver's own members are asked as well.
function memberOfAccess(symbol, receiver, name) {
    const index = memberOfSymbol(symbol);
    if (index >= 0) return index;
    if (receiver && !(symbol?.declarations?.length)) {
        for (const property of propertiesOf(checker.getTypeAtLocation(receiver), name)) {
            const found = memberOfSymbol(property);
            if (found >= 0) return found;
        }
    }
    return -1;
}

function isWrapper(node) {
    return ts.isParenthesizedExpression(node) || ts.isNonNullExpression(node) || ts.isAsExpression(node) ||
        ts.isTypeAssertionExpression(node) || ts.isSatisfiesExpression(node);
}

// Whether node sits in an object or array literal that is itself assigned to (`({ a: x.m } = y)`, `for ([x.m] of y)`).
function inDestructuringTarget(node) {
    let child = node;
    let parent = node.parent;
    let throughLiteral = false;
    while (parent && (ts.isPropertyAssignment(parent) || ts.isSpreadAssignment(parent) || ts.isSpreadElement(parent) ||
        ts.isArrayLiteralExpression(parent) || ts.isObjectLiteralExpression(parent) || ts.isParenthesizedExpression(parent))) {
        if (ts.isPropertyAssignment(parent) && parent.initializer !== child) return false;
        if (ts.isArrayLiteralExpression(parent) || ts.isObjectLiteralExpression(parent)) throughLiteral = true;
        child = parent;
        parent = parent.parent;
    }
    if (!throughLiteral || !parent) return false;
    if (ts.isBinaryExpression(parent)) return parent.left === child && parent.operatorToken.kind === SK.EqualsToken;
    return (ts.isForOfStatement(parent) || ts.isForInStatement(parent)) && parent.initializer === child;
}

function accessKind(node) {
    let child = node;
    let parent = node.parent;
    while (parent && isWrapper(parent)) {
        child = parent;
        parent = parent.parent;
    }
    if (ts.isBinaryExpression(parent) && parent.left === child) {
        const operator = parent.operatorToken.kind;
        if (operator === SK.EqualsToken) return "write";
        if (operator > SK.EqualsToken && operator <= SK.LastAssignment) return "compound";
    }
    if ((ts.isPrefixUnaryExpression(parent) || ts.isPostfixUnaryExpression(parent)) &&
        (parent.operator === SK.PlusPlusToken || parent.operator === SK.MinusMinusToken)) return "compound";
    if (ts.isDeleteExpression(parent)) return "write";
    if (inDestructuringTarget(child)) return "write";
    return "read";
}

const counterOf = { read: "r", write: "w", compound: "c", in: "i" };
const staticCounts = members.map(() => ({ read: 0, write: 0, compound: 0, in: 0, destructure: 0, uncounted: 0 }));
const accesses = [];
const edits = new Map();
let sequence = 0;

function where(sourceFile, node) {
    const at = sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile));
    return { file: path.relative(tree, sourceFile.fileName), line: at.line + 1, column: at.character + 1 };
}
function record(sourceFile, node, index, kind, counted, via) {
    accesses.push({ member: index, id: members[index].id, kind, ...where(sourceFile, node), counted, via });
    if (!counted) staticCounts[index].uncounted++;
    else if (kind.startsWith("destructure")) staticCounts[index].destructure++;
    else staticCounts[index][kind]++;
}
function editsOf(sourceFile) {
    if (!edits.has(sourceFile.fileName)) edits.set(sourceFile.fileName, []);
    return edits.get(sourceFile.fileName);
}
// rank orders edits at one position: closes, then a replacement, then opens.
function replace(sourceFile, start, end, text) {
    editsOf(sourceFile).push({ start, end, text, rank: 1, key: 0, sequence: sequence++ });
}
function wrap(sourceFile, node, open, close) {
    const start = node.getStart(sourceFile);
    editsOf(sourceFile).push({ start, end: start, text: open, rank: 2, key: -node.end, sequence: sequence++ });
    editsOf(sourceFile).push({ start: node.end, end: node.end, text: close, rank: 0, key: -start, sequence: sequence++ });
}
function insert(sourceFile, at, text, extent) {
    editsOf(sourceFile).push({ start: at, end: at, text, rank: 2, key: -extent, sequence: sequence++ });
}
const bump = (kind, index) => `__p72.${counterOf[kind]}[${index}]++`;

function literalName(node) {
    if (!node) return undefined;
    if (ts.isIdentifier(node) || ts.isStringLiteral(node) || ts.isNumericLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) return node.text;
    if (ts.isComputedPropertyName(node)) return literalName(node.expression);
    return undefined;
}

// The selected members a binding pattern reads, one read per element, through nested patterns too.
function patternReads(pattern) {
    const found = [];
    if (ts.isObjectBindingPattern(pattern)) {
        for (const element of pattern.elements) {
            if (element.dotDotDotToken) continue;
            const name = literalName(element.propertyName ?? element.name);
            if (name !== undefined && trackedNames.has(name)) {
                let index = -1;
                for (const property of propertiesOf(checker.getTypeAtLocation(pattern), name)) {
                    index = memberOfSymbol(property);
                    if (index >= 0) break;
                }
                if (index >= 0) found.push({ element, index });
            }
            if (ts.isObjectBindingPattern(element.name) || ts.isArrayBindingPattern(element.name)) found.push(...patternReads(element.name));
        }
    }
    else if (ts.isArrayBindingPattern(pattern)) {
        for (const element of pattern.elements) {
            if (ts.isBindingElement(element) && (ts.isObjectBindingPattern(element.name) || ts.isArrayBindingPattern(element.name))) found.push(...patternReads(element.name));
        }
    }
    return found;
}

// Counts at the start of a statement or body that runs once per binding: a block gets a statement after its `{`, any
// other statement is put in a block, and an arrow's expression body becomes a comma expression.
function countAtBody(sourceFile, body, reads, via) {
    const counts = reads.map(r => bump("read", r.index)).join(", ");
    if (ts.isBlock(body)) insert(sourceFile, body.getStart(sourceFile) + 1, ` ${counts};`, body.end);
    else if (ts.isExpression(body)) wrap(sourceFile, body, `(${counts}, `, ")");
    else wrap(sourceFile, body, `{ ${counts}; `, " }");
    for (const r of reads) record(sourceFile, r.element, r.index, `destructure-${via}`, true, "binding pattern");
}

function isProgramSource(sourceFile) {
    return !sourceFile.isDeclarationFile && sourceFile.fileName.startsWith(path.join(tree, "src") + path.sep);
}

for (const sourceFile of program.getSourceFiles()) {
    if (!isProgramSource(sourceFile)) continue;
    (function visit(node) {
        if (ts.isPropertyAccessExpression(node) && ts.isIdentifier(node.name) && trackedNames.has(node.name.text)) {
            const name = node.name.text;
            const index = memberOfAccess(checker.getSymbolAtLocation(node.name), node.expression, name);
            if (index >= 0) {
                const kind = accessKind(node);
                if (node.expression.kind === SK.SuperKeyword) record(sourceFile, node, index, kind, false, "super receiver");
                else {
                    const start = node.questionDotToken ? node.questionDotToken.getStart(sourceFile) : node.expression.end;
                    replace(sourceFile, start, node.name.end, `${node.questionDotToken ? "?." : ""}[(${bump(kind, index)}, ${JSON.stringify(name)})]`);
                    record(sourceFile, node, index, kind, true, "property access");
                }
            }
        }
        else if (ts.isElementAccessExpression(node)) {
            const name = literalName(node.argumentExpression);
            if (name !== undefined && trackedNames.has(name) && !ts.isComputedPropertyName(node.argumentExpression)) {
                let symbol = checker.getSymbolAtLocation(node.argumentExpression);
                if (!symbol) symbol = propertiesOf(checker.getTypeAtLocation(node.expression), name)[0];
                const index = memberOfAccess(symbol, node.expression, name);
                if (index >= 0) {
                    const kind = accessKind(node);
                    wrap(sourceFile, node.argumentExpression, `(${bump(kind, index)}, `, ")");
                    record(sourceFile, node, index, kind, true, "element access");
                }
            }
        }
        else if (ts.isVariableDeclaration(node) && (ts.isObjectBindingPattern(node.name) || ts.isArrayBindingPattern(node.name))) {
            const reads = patternReads(node.name);
            if (reads.length) {
                const loop = node.parent?.parent;
                if (node.initializer) {
                    wrap(sourceFile, node.initializer, `(${reads.map(r => bump("read", r.index)).join(", ")}, `, ")");
                    for (const r of reads) record(sourceFile, r.element, r.index, "destructure-initializer", true, "binding pattern");
                }
                else if (loop && (ts.isForOfStatement(loop) || ts.isForInStatement(loop)) && loop.initializer === node.parent) {
                    countAtBody(sourceFile, loop.statement, reads, "loop");
                }
                else for (const r of reads) record(sourceFile, r.element, r.index, "destructure", false, "binding with no initializer");
            }
        }
        else if (ts.isParameter(node) && (ts.isObjectBindingPattern(node.name) || ts.isArrayBindingPattern(node.name))) {
            const reads = patternReads(node.name);
            if (reads.length) {
                const body = node.parent.body;
                if (body) countAtBody(sourceFile, body, reads, "parameter");
                else for (const r of reads) record(sourceFile, r.element, r.index, "destructure", false, "parameter of a bodiless signature");
            }
        }
        else if (ts.isBindingElement(node) || ts.isCatchClause(node)) {
            // Elements are reached through their declaration's pattern above; a catch binding pattern is left uncounted.
            if (ts.isCatchClause(node) && node.variableDeclaration && !ts.isIdentifier(node.variableDeclaration.name)) {
                for (const r of patternReads(node.variableDeclaration.name)) record(sourceFile, r.element, r.index, "destructure", false, "catch binding");
                ts.forEachChild(node.block, visit);
                return;
            }
        }
        else if (ts.isBinaryExpression(node) && node.operatorToken.kind === SK.InKeyword) {
            const name = literalName(node.left);
            if (name !== undefined && trackedNames.has(name) && (ts.isStringLiteral(node.left) || ts.isNoSubstitutionTemplateLiteral(node.left))) {
                let index = -1;
                let via = "right operand's type";
                for (const property of propertiesOf(checker.getTypeAtLocation(node.right), name)) {
                    index = memberOfSymbol(property);
                    if (index >= 0) break;
                }
                // `"m" in x` asks because x's type may lack m; with no member there, the check is the name's.
                if (index < 0) {
                    index = members.findIndex(m => m.name === name);
                    via = "name alone";
                }
                wrap(sourceFile, node.left, `(${bump("in", index)}, `, ")");
                record(sourceFile, node, index, "in", true, via);
            }
        }
        ts.forEachChild(node, visit);
    })(sourceFile);
}

// Apply the edits in source order to each file's copy. An edit starting inside an earlier replacement would mean two
// edits claim the same text, which the rules above never make; it is refused rather than guessed at.
fs.mkdirSync(copyDir, { recursive: true });
fs.cpSync(path.join(tree, "src"), path.join(copyDir, "src"), { recursive: true });
const header = `var __p72: any = ((globalThis as any).__p72 ??= { r: new Float64Array(${members.length}), w: new Float64Array(${members.length}), c: new Float64Array(${members.length}), i: new Float64Array(${members.length}) });\n`;
let editCount = 0;
for (const [fileName, list] of edits) {
    const text = fs.readFileSync(fileName, "utf8");
    list.sort((a, b) => a.start - b.start || a.rank - b.rank || a.key - b.key || (a.rank === 0 ? b.sequence - a.sequence : a.sequence - b.sequence));
    let out = "";
    let cursor = 0;
    for (const edit of list) {
        if (edit.start < cursor) {
            const at = ts.getLineAndCharacterOfPosition(program.getSourceFile(fileName), edit.start);
            throw new Error(`overlapping edits in ${path.relative(tree, fileName)} at ${at.line + 1}:${at.character + 1}`);
        }
        out += text.slice(cursor, edit.start) + edit.text;
        cursor = edit.end;
    }
    out += text.slice(cursor);
    const bom = out.startsWith("﻿") ? "﻿" : "";
    fs.writeFileSync(path.join(copyDir, path.relative(tree, fileName)), bom + header + out.slice(bom.length));
    editCount += list.length;
}

// tsc.ts isn't in the src/compiler program; it only gets the exit hook. The built tree is CommonJS, so require is there.
const entry = path.join(copyDir, "src/tsc/tsc.ts");
const hook = [
    "process.on(\"exit\", () => {",
    "    const out = process.env.P72_COUNTS;",
    "    const counters = (globalThis as any).__p72;",
    "    if (out) require(\"node:fs\").writeFileSync(out, JSON.stringify({ r: Array.from(counters?.r ?? []), w: Array.from(counters?.w ?? []), c: Array.from(counters?.c ?? []), i: Array.from(counters?.i ?? []) }));",
    "});",
    "",
].join("\n");
fs.writeFileSync(entry, hook + fs.readFileSync(entry, "utf8"));

const outDir = path.dirname(membersPath);
fs.writeFileSync(path.join(outDir, "instrumented.json"), JSON.stringify({ members: members.map((m, i) => ({ index: i, ...m, static: staticCounts[i] })) }, null, 1) + "\n");
fs.writeFileSync(path.join(outDir, "accesses.csv"), common.csv(accesses, ["member", "id", "kind", "file", "line", "column", "counted", "via"]));
const counted = accesses.filter(a => a.counted).length;
console.log(`instrumented ${edits.size} files, ${editCount} edits; ${accesses.length} static access sites, ${counted} counted, ${accesses.length - counted} uncounted`);
for (const [i, m] of members.entries()) console.log(`  ${m.id}: ${JSON.stringify(staticCounts[i])}`);
