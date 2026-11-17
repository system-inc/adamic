// #r11wppy item 1: the readonly fixpoint for tsc's shared emptyArray and resolvingEmptyArray. Both are typed
// `readonly never[]`, then every slot that receives one is made readonly too, round after round, until a round
// makes no edit. A slot tsc really writes through becomes `readonly never[] | <original>` so the rounds converge
// (an alias slot is put back instead), CompilerOptions.lib and .types are never edited, and a type alias under src
// is edited as its own slot. Never merged: a proof only.
//
// With --hold-public (the third pass) a Public slot is never edited either: it is Held. Once the rounds converge,
// each readonly value still going into a Held or Dropped slot has its source put back and marked Blocked, round
// after round, so what stays Readonly is the Local and Internal change alone, and each Held slot shows the listed
// sites it holds. The table is then table3.md.
//
// Usage: node fixpoint.mjs <adapted tree> <never-sites.tsv> <out directory> [--hold-public]
//
// Every edit is an insertion of text with no line break ("readonly ", "Readonly", "(readonly ", ")" and a written
// slot's "readonly never[] | "), so line
// numbers never move, and each insertion is kept at its offset in the untouched copy. A file's text is always the
// untouched text plus its live insertions, so a revert drops a slot's insertions and offsets never go stale.
import fs from 'node:fs';
import path from 'node:path';
import {createRequire} from 'node:module';
import {fileURLToPath} from 'node:url';

function failLoudly(error) {
    console.log(`VERDICT FAIL ${error && error.stack ? error.stack : String(error)}`);
    process.exit(1);
}
process.on('uncaughtException', failLoudly);
process.on('unhandledRejection', failLoudly);

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
// The script lives at proof/performance/r11wppy-fixpoint, three directories under the tree's root
const treeRoot = path.resolve(scriptDirectory, '..', '..', '..');
const typeScriptPath = path.join(treeRoot, 'stage3', 'api', 'node_modules', 'typescript');
const typeRootsPath = path.join(treeRoot, 'stage3', 'api', 'node_modules', '@types');
if (!fs.existsSync(typeScriptPath)) throw new Error(`no TypeScript at ${typeScriptPath}: run after sealed_npm stage3/api`);
if (!fs.existsSync(path.join(typeRootsPath, 'node'))) throw new Error(`no @types/node under ${typeRootsPath}`);
const ts = createRequire(import.meta.url)(typeScriptPath);

const [adaptedArgument, sitesArgument, outArgument, ...flagArguments] = process.argv.slice(2);
if (!adaptedArgument || !sitesArgument || !outArgument) {
    throw new Error('usage: node fixpoint.mjs <adapted tree> <never-sites.tsv> <out directory> [--hold-public]');
}
for (const flag of flagArguments) {
    if (flag !== '--hold-public') throw new Error(`unknown argument ${flag}: the only flag is --hold-public`);
}
const holdPublic = flagArguments.includes('--hold-public');
const adaptedDirectory = path.resolve(adaptedArgument);
const sitesPath = path.resolve(sitesArgument);
fs.mkdirSync(path.resolve(outArgument), {recursive: true});
// The real path, so the held texts' names match the names module resolution builds
const outDirectory = fs.realpathSync(path.resolve(outArgument));
const outSource = path.join(outDirectory, 'src');
const maximumRounds = 80;
const maximumBlockRounds = 40;
const tableName = holdPublic ? 'table3.md' : 'table.md';
const sourceLineLimit = 160;
const programDirectories = ['compiler', 'services', 'server', 'jsTyping', 'deprecatedCompat', 'tsc', 'tsserver',
    'typescript', 'typingsInstaller', 'typingsInstallerCore', 'watchGuard'];
const writeMethods = new Set(['push', 'pop', 'shift', 'unshift', 'splice', 'sort', 'reverse', 'fill', 'copyWithin']);
const readonlyPhrase = "is 'readonly' and cannot be assigned to the mutable type";
const passThroughOperators = new Set([ts.SyntaxKind.BarBarToken, ts.SyntaxKind.QuestionQuestionToken]);
const began = Date.now();

function elapsed() {
    return `${((Date.now() - began) / 1000).toFixed(1)} s`;
}

function log(line) {
    console.log(`fixpoint [${elapsed()}] ${line}`);
}

// SyntaxKind's reverse map holds marker names like FirstStatement; prefer the real kind's name
const kindNames = new Map();
for (const [name, value] of Object.entries(ts.SyntaxKind)) {
    if (typeof value !== 'number' || /^(First|Last)/.test(name) || kindNames.has(value)) continue;
    kindNames.set(value, name);
}
function kindName(kind) {
    return kindNames.get(kind) ?? String(kind);
}

// 1. The copy
if (!fs.existsSync(path.join(adaptedDirectory, 'src'))) throw new Error(`no src under ${adaptedDirectory}`);
if (fs.existsSync(outSource)) throw new Error(`${outSource} is already there: give a fresh out directory`);
fs.cpSync(path.join(adaptedDirectory, 'src'), outSource, {recursive: true});
log(`TypeScript ${ts.version}, the adapted src copied to ${outSource}`);

function walkTypeScriptFiles(directory, found) {
    for (const entry of fs.readdirSync(directory, {withFileTypes: true})) {
        const full = path.join(directory, entry.name);
        if (entry.isDirectory()) walkTypeScriptFiles(full, found);
        else if (entry.isFile() && entry.name.endsWith('.ts')) found.push(full);
    }
    return found;
}

// Every .ts under the copy is held in memory; edits only ever touch these texts
const originalText = new Map();
const currentText = new Map();
for (const file of walkTypeScriptFiles(outSource, [])) {
    const text = fs.readFileSync(file, 'utf8');
    originalText.set(file, text);
    currentText.set(file, text);
}

function relativeName(absolute) {
    const relative = path.relative(outDirectory, absolute);
    return relative.startsWith('..') || path.isAbsolute(relative) ? absolute : relative.split(path.sep).join('/');
}

// Line starts the way TypeScript counts them: CR, LF, CRLF, U+2028 and U+2029
const lineStartsCache = new Map();
function lineStartsOf(text) {
    const cached = lineStartsCache.get(text);
    if (cached) return cached;
    const starts = [0];
    for (let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        if (code === 13) {
            if (text.charCodeAt(index + 1) === 10) index++;
            starts.push(index + 1);
        } else if (code === 10 || code === 0x2028 || code === 0x2029) {
            starts.push(index + 1);
        }
    }
    // Held texts are a few thousand at most, and checker.ts alone is worth keeping rather than recounting
    if (lineStartsCache.size > 5000) lineStartsCache.clear();
    lineStartsCache.set(text, starts);
    return starts;
}
function lineIndexOf(starts, position) {
    let low = 0;
    let high = starts.length - 1;
    while (low < high) {
        const middle = (low + high + 1) >> 1;
        if (starts[middle] <= position) low = middle;
        else high = middle - 1;
    }
    return low;
}
function lineText(text, lineIndex) {
    const starts = lineStartsOf(text);
    const end = lineIndex + 1 < starts.length ? starts[lineIndex + 1] : text.length;
    return text.slice(starts[lineIndex], end).replace(/[\r\n\u2028\u2029]+$/, '');
}
function clip(text, limit = sourceLineLimit) {
    const single = text.replace(/\s+/g, ' ').trim();
    return single.length > limit ? `${single.slice(0, limit - 3)}...` : single;
}

// Insertions by file, each at an offset in the untouched text, ordered by offset and then by rank (an outer
// array's "readonly " goes before an inner element's "(readonly " at the same offset)
const insertionsByFile = new Map();
function insertionsOf(absolute) {
    return insertionsByFile.get(absolute) ?? [];
}
function sortInsertions(list) {
    list.sort((left, right) => left.original - right.original || left.rank - right.rank);
}
// A position inside inserted text maps to the offset the insertion sits at
function toOriginal(absolute, position) {
    let shift = 0;
    for (const insertion of insertionsOf(absolute)) {
        const currentAt = insertion.original + shift;
        if (position < currentAt) break;
        if (position < currentAt + insertion.text.length) return insertion.original;
        shift += insertion.text.length;
    }
    return position - shift;
}
function toCurrent(absolute, target) {
    let shift = 0;
    for (const insertion of insertionsOf(absolute)) {
        if (insertion === target) return insertion.original + shift;
        shift += insertion.text.length;
    }
    throw new Error(`${relativeName(absolute)}: an insertion at ${target.original} is not live`);
}
// The insertions must be ordered as sortInsertions orders them
function renderWith(text, insertions) {
    let result = '';
    let cursor = 0;
    for (const insertion of insertions) {
        result += text.slice(cursor, insertion.original) + insertion.text;
        cursor = insertion.original;
    }
    return result + text.slice(cursor);
}
function render(absolute) {
    return renderWith(originalText.get(absolute), insertionsOf(absolute));
}
function originalLineColumn(absolute, originalPosition) {
    const starts = lineStartsOf(originalText.get(absolute));
    const index = lineIndexOf(starts, originalPosition);
    return {line: index + 1, column: originalPosition - starts[index] + 1};
}

// 2. The program
const configPath = path.join(outSource, 'tsconfig-base.json');
const configRead = ts.readConfigFile(configPath, ts.sys.readFile);
if (configRead.error) throw new Error(`${configPath}: ${ts.flattenDiagnosticMessageText(configRead.error.messageText, '\n')}`);
const configParsed = ts.parseJsonConfigFileContent(configRead.config, ts.sys, path.dirname(configPath), undefined, configPath);
// 18003 is "no inputs were found": the base config names no files, and the roots are given here
for (const error of configParsed.errors.filter(error => error.code !== 18003)) {
    log(`tsconfig-base.json note ${error.code}: ${clip(ts.flattenDiagnosticMessageText(error.messageText, '\n'), 200)}`);
}
const baseOptions = {...configParsed.options};
delete baseOptions.tsBuildInfoFile;
const options = {
    ...baseOptions,
    noEmit: true,
    composite: false,
    declaration: false,
    declarationMap: false,
    emitDeclarationOnly: false,
    isolatedDeclarations: false,
    incremental: false,
    types: ['node'],
    typeRoots: [typeRootsPath],
};
const rootNames = [];
for (const directory of programDirectories) {
    const full = path.join(outSource, directory);
    if (!fs.existsSync(full)) throw new Error(`no src/${directory} in the adapted tree`);
    rootNames.push(...walkTypeScriptFiles(full, []));
}
rootNames.sort();
log(`${rootNames.length} root files from src/{${programDirectories.join(',')}}`);

// The host serves the held texts, and hands back the same SourceFile while a text is unchanged so the old program's
// structure is reused
const host = ts.createCompilerHost(options, true);
const baseGetSourceFile = host.getSourceFile.bind(host);
const baseReadFile = host.readFile.bind(host);
const baseFileExists = host.fileExists.bind(host);
const heldSourceFiles = new Map();
const outsideSourceFiles = new Map();
host.getSourceFile = (fileName, languageVersionOrOptions, onError, shouldCreateNewSourceFile) => {
    const absolute = path.resolve(fileName);
    const text = currentText.get(absolute);
    if (text === undefined) {
        const cached = outsideSourceFiles.get(absolute);
        if (cached) return cached;
        const created = baseGetSourceFile(fileName, languageVersionOrOptions, onError, shouldCreateNewSourceFile);
        if (created) outsideSourceFiles.set(absolute, created);
        return created;
    }
    const held = heldSourceFiles.get(absolute);
    if (held && held.text === text) return held;
    const created = ts.createSourceFile(fileName, text, languageVersionOrOptions, true);
    heldSourceFiles.set(absolute, created);
    return created;
};
host.readFile = fileName => currentText.get(path.resolve(fileName)) ?? baseReadFile(fileName);
host.fileExists = fileName => currentText.has(path.resolve(fileName)) || baseFileExists(fileName);

function buildProgram(oldProgram) {
    return ts.createProgram({rootNames, options, host, oldProgram});
}

function describeDiagnostic(diagnostic) {
    const message = ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n');
    if (!diagnostic.file || diagnostic.start === undefined) {
        return {diagnostic, message, code: diagnostic.code, relative: '<global>', key: `<global>|${diagnostic.code}|${message}`, site: '<global>'};
    }
    const absolute = path.resolve(diagnostic.file.fileName);
    const relative = relativeName(absolute);
    const held = currentText.has(absolute);
    const length = diagnostic.length ?? 0;
    const originalStart = held ? toOriginal(absolute, diagnostic.start) : diagnostic.start;
    const originalEnd = held ? toOriginal(absolute, diagnostic.start + length) : diagnostic.start + length;
    let line;
    let column;
    if (held) {
        ({line, column} = originalLineColumn(absolute, originalStart));
    } else {
        const position = diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start);
        line = position.line + 1;
        column = position.character + 1;
    }
    return {
        diagnostic, message, code: diagnostic.code, absolute, relative, held, originalStart, originalEnd, line, column,
        key: `${relative}|${originalStart}|${diagnostic.code}|${message}`,
        site: `${relative}:${line}:${column}`,
    };
}

function allDiagnostics(checkedProgram) {
    const seen = new Set();
    const described = [];
    const diagnostics = [...checkedProgram.getSyntacticDiagnostics(), ...checkedProgram.getGlobalDiagnostics(), ...checkedProgram.getSemanticDiagnostics()];
    for (const diagnostic of diagnostics) {
        const record = describeDiagnostic(diagnostic);
        if (seen.has(record.key)) continue;
        seen.add(record.key);
        described.push(record);
    }
    return described;
}

// The deepest node whose span (without its leading trivia) holds the position, found by hand with forEachChild
function deepestNodeAt(sourceFile, position) {
    let current = sourceFile;
    for (;;) {
        let next;
        ts.forEachChild(current, child => {
            if (child.getStart(sourceFile) <= position && position < child.end) {
                next = child;
                return true;
            }
            return undefined;
        });
        if (!next) return current;
        current = next;
    }
}
// The node a diagnostic was reported on: the outermost node starting at its start and ending inside its span
function errorNodeAt(sourceFile, start, length) {
    let node = deepestNodeAt(sourceFile, start);
    while (node.parent && !ts.isSourceFile(node.parent) && node.parent.getStart(sourceFile) === start && node.parent.end <= start + length) {
        node = node.parent;
    }
    return node;
}

function nameText(name, sourceFile) {
    if (!name) return undefined;
    if (ts.isIdentifier(name) || ts.isPrivateIdentifier(name) || ts.isStringLiteralLike(name) || ts.isNumericLiteral(name)) return name.text;
    return name.getText(sourceFile);
}
function functionName(fn, sourceFile) {
    const own = nameText(fn.name, sourceFile);
    if (own) return own;
    if (ts.isConstructorDeclaration(fn)) return 'constructor';
    const parent = fn.parent;
    if (parent && (ts.isVariableDeclaration(parent) || ts.isPropertyAssignment(parent) || ts.isPropertyDeclaration(parent))) {
        return nameText(parent.name, sourceFile) ?? '<anonymous>';
    }
    return '<anonymous>';
}
function containerName(member, sourceFile) {
    for (let node = member.parent; node && !ts.isSourceFile(node); node = node.parent) {
        if (ts.isClassLike(node) || ts.isInterfaceDeclaration(node) || ts.isTypeAliasDeclaration(node)) {
            return nameText(node.name, sourceFile) ?? '<class>';
        }
        if (ts.isFunctionLike(node)) return `${functionName(node, sourceFile)}()`;
    }
    return '<literal>';
}
function declarationLabel(declaration, sourceFile) {
    if (ts.isVariableDeclaration(declaration)) return `var ${nameText(declaration.name, sourceFile)}`;
    if (ts.isParameter(declaration)) {
        const prefix = declaration.dotDotDotToken ? '...' : '';
        return `param ${functionName(declaration.parent, sourceFile)}(${prefix}${nameText(declaration.name, sourceFile)})`;
    }
    if (ts.isPropertyDeclaration(declaration) || ts.isPropertySignature(declaration)) {
        return `prop ${containerName(declaration, sourceFile)}.${nameText(declaration.name, sourceFile)}`;
    }
    if (ts.isFunctionLike(declaration)) return `return ${functionName(declaration, sourceFile)}()`;
    const name = 'name' in declaration ? nameText(declaration.name, sourceFile) : undefined;
    return `${kindName(declaration.kind)} ${name ?? '<unnamed>'}`;
}

function hasInternalTag(node, sourceFile) {
    if (ts.getJSDocTags(node).some(tag => tag.tagName.text === 'internal')) return true;
    return sourceFile.text.slice(node.pos, node.getStart(sourceFile)).includes('@internal');
}
function hasModifier(node, kind) {
    return ts.canHaveModifiers(node) && (ts.getModifiers(node) ?? []).some(modifier => modifier.kind === kind);
}
// Public when a declaration file would show it; Local when it lives in a body, is private, or sits under a
// statement the module doesn't export; Internal when it or an ancestor carries @internal
function visibilityOf(declaration) {
    const sourceFile = declaration.getSourceFile();
    let local = false;
    let internal = false;
    for (let node = declaration; node && !ts.isSourceFile(node); node = node.parent) {
        if (hasInternalTag(node, sourceFile)) internal = true;
        if (ts.isBlock(node)) local = true;
        if (hasModifier(node, ts.SyntaxKind.PrivateKeyword)) local = true;
        if ('name' in node && node.name && ts.isPrivateIdentifier(node.name)) local = true;
        if (node.parent && ts.isSourceFile(node.parent) && !sourceFile.isDeclarationFile && !ts.isModuleDeclaration(node)
            && !hasModifier(node, ts.SyntaxKind.ExportKeyword)) {
            local = true;
        }
    }
    if (local) return 'Local';
    return internal ? 'Internal' : 'Public';
}

// The slots, by key: declaration file:line plus a label, with "[element]" once per array level below the declared type
const slots = new Map();
const readonlyDiagnostics = new Map();
const scriptErrors = [];
let currentRound = 0;
// Adamic's ruling: these two are never edited, and are recorded as Dropped with their sites
const droppedSlots = new Set(['src/compiler/types.ts prop CompilerOptions.lib', 'src/compiler/types.ts prop CompilerOptions.types']);

function slotIdentity(declaration, elementDepth) {
    const sourceFile = declaration.getSourceFile();
    const absolute = path.resolve(sourceFile.fileName);
    const relative = relativeName(absolute);
    const line = sourceFile.getLineAndCharacterOfPosition(declaration.getStart(sourceFile)).line + 1;
    const label = declarationLabel(declaration, sourceFile) + '[element]'.repeat(elementDepth);
    return {key: `${relative}:${line} ${label}`, absolute, relative, line, label};
}

// The type node a slot's readonly goes on: the declared type (a function's return type), then one array level down
// per element step
function slotTypeNode(declaration, elementDepth) {
    let typeNode = declaration.type;
    if (!typeNode) return {missing: true};
    for (let step = 0; step < elementDepth; step++) {
        while (ts.isParenthesizedTypeNode(typeNode)) typeNode = typeNode.type;
        if (ts.isArrayTypeNode(typeNode)) {
            typeNode = typeNode.elementType;
        } else if (ts.isTypeReferenceNode(typeNode) && ['Array', 'ReadonlyArray'].includes(typeNode.typeName.getText())
            && typeNode.typeArguments?.length === 1) {
            typeNode = typeNode.typeArguments[0];
        } else {
            return {named: `element of <${kindName(typeNode.kind)}> ${clip(typeNode.getText(), 60)}`};
        }
    }
    return {typeNode};
}

function isKeywordOrLiteralType(typeNode) {
    return ts.isLiteralTypeNode(typeNode) || (typeNode.kind >= ts.SyntaxKind.FirstKeyword && typeNode.kind <= ts.SyntaxKind.LastKeyword);
}
function newPlan(checker) {
    return {checker, insertions: [], changed: [], named: [], other: [], aliases: [], already: false};
}
// A rank orders insertions at one offset: an outer array's "readonly " (lower rank) before its element's "(readonly "
function pushArrayInsertion(typeNode, sourceFile, plan, rank) {
    const start = typeNode.getStart(sourceFile);
    // An element type of an array type needs parentheses: (readonly T[])[]
    if (ts.isArrayTypeNode(typeNode.parent) && typeNode.parent.elementType === typeNode) {
        plan.insertions.push({position: start, text: '(readonly ', rank}, {position: typeNode.end, text: ')', rank});
    } else {
        plan.insertions.push({position: start, text: 'readonly ', rank});
    }
    plan.changed.push(typeNode.getText(sourceFile));
}
// The type alias a reference names, when the alias is declared in the copy of src
function heldAliasOf(typeReference, checker) {
    if (!checker) return undefined;
    let symbol = checker.getSymbolAtLocation(typeReference.typeName);
    if (!symbol) return undefined;
    if (symbol.flags & ts.SymbolFlags.Alias) symbol = checker.getAliasedSymbol(symbol);
    const declaration = (symbol.declarations ?? []).find(candidate => ts.isTypeAliasDeclaration(candidate));
    if (!declaration || !currentText.has(path.resolve(declaration.getSourceFile().fileName))) return undefined;
    return declaration;
}
function planReadonly(typeNode, sourceFile, plan) {
    if (ts.isParenthesizedTypeNode(typeNode)) {
        planReadonly(typeNode.type, sourceFile, plan);
        return;
    }
    if (ts.isArrayTypeNode(typeNode)) {
        pushArrayInsertion(typeNode, sourceFile, plan, undefined);
        return;
    }
    if (ts.isTypeReferenceNode(typeNode)) {
        const name = typeNode.typeName.getText(sourceFile);
        if (name === 'Array') {
            plan.insertions.push({position: typeNode.typeName.getStart(sourceFile), text: 'Readonly'});
            plan.changed.push(typeNode.getText(sourceFile));
        } else if (name === 'ReadonlyArray') {
            plan.already = true;
        } else {
            const alias = heldAliasOf(typeNode, plan.checker);
            if (alias) plan.aliases.push(alias);
            else plan.named.push(name);
        }
        return;
    }
    if (ts.isTypeOperatorNode(typeNode) && typeNode.operator === ts.SyntaxKind.ReadonlyKeyword) {
        plan.already = true;
        return;
    }
    if (ts.isUnionTypeNode(typeNode)) {
        for (const member of typeNode.types) {
            if (!isKeywordOrLiteralType(member)) planReadonly(member, sourceFile, plan);
        }
        return;
    }
    plan.other.push(kindName(typeNode.kind));
}

// An alias body: every mutable X[] or Array<X> anywhere inside it becomes readonly, and aliases it names are
// gathered to be edited as their own slots
function planDeep(typeNode, sourceFile, plan, rank) {
    if (ts.isTypeOperatorNode(typeNode) && typeNode.operator === ts.SyntaxKind.ReadonlyKeyword) {
        // The array right under readonly already is; its elements still count
        if (ts.isArrayTypeNode(typeNode.type)) planDeep(typeNode.type.elementType, sourceFile, plan, rank + 1);
        else planDeep(typeNode.type, sourceFile, plan, rank + 1);
        return;
    }
    if (ts.isArrayTypeNode(typeNode)) {
        pushArrayInsertion(typeNode, sourceFile, plan, rank);
        planDeep(typeNode.elementType, sourceFile, plan, rank + 1);
        return;
    }
    if (ts.isTypeReferenceNode(typeNode)) {
        const name = typeNode.typeName.getText(sourceFile);
        if (name === 'Array') {
            plan.insertions.push({position: typeNode.typeName.getStart(sourceFile), text: 'Readonly', rank});
            plan.changed.push(typeNode.getText(sourceFile));
        } else if (name !== 'ReadonlyArray') {
            const alias = heldAliasOf(typeNode, plan.checker);
            if (alias && !plan.aliases.includes(alias)) plan.aliases.push(alias);
        }
        for (const argument of typeNode.typeArguments ?? []) planDeep(argument, sourceFile, plan, rank + 1);
        return;
    }
    ts.forEachChild(typeNode, child => {
        planDeep(child, sourceFile, plan, rank);
        return undefined;
    });
}

function isReadonlyTypeNode(typeNode) {
    if (!typeNode) return false;
    if (ts.isParenthesizedTypeNode(typeNode)) return isReadonlyTypeNode(typeNode.type);
    if (ts.isTypeOperatorNode(typeNode)) return typeNode.operator === ts.SyntaxKind.ReadonlyKeyword;
    if (ts.isTypeReferenceNode(typeNode)) return typeNode.typeName.getText() === 'ReadonlyArray';
    if (ts.isUnionTypeNode(typeNode)) {
        const members = typeNode.types.filter(member => !isKeywordOrLiteralType(member));
        return members.length > 0 && members.every(isReadonlyTypeNode);
    }
    return false;
}

// Where a written slot's `readonly never[] | <original>` goes, in untouched offsets, and which parentheses it needs
function typeSpanOf(typeNode, sourceFile, absolute, rank) {
    const parent = typeNode.parent;
    const outer = ts.isArrayTypeNode(parent) || ts.isTypeOperatorNode(parent) || ts.isIndexedAccessTypeNode(parent)
        || ts.isIntersectionTypeNode(parent) || ts.isOptionalTypeNode(parent) || ts.isUnionTypeNode(parent);
    const inner = ts.isFunctionTypeNode(typeNode) || ts.isConstructorTypeNode(typeNode) || ts.isConditionalTypeNode(typeNode);
    return {start: toOriginal(absolute, typeNode.getStart(sourceFile)), end: toOriginal(absolute, typeNode.end), outer, inner, rank};
}

function containingFunction(node) {
    for (let current = node.parent; current && !ts.isSourceFile(current); current = current.parent) {
        if (ts.isFunctionLike(current)) return current;
    }
    return undefined;
}

function declarationOfSymbol(symbol, checker) {
    let target = symbol;
    if (target.flags & ts.SymbolFlags.Alias) target = checker.getAliasedSymbol(target);
    if (target.valueDeclaration) return {declaration: target.valueDeclaration};
    const declarations = target.declarations ?? [];
    if (declarations.length === 1) return {declaration: declarations[0]};
    return {unresolved: `${declarations.length} declarations of ${target.name}`};
}

// The slot an expression reads: a variable, parameter or property by its symbol, a call by its signature's
// declaration, an element access one array level down. With aliases followed, an unannotated variable's
// initializer is followed too, so `const list = type.list; list.push(x)` reaches type.list.
function expressionSlot(expression, checker, elementDepth, followAliases, via) {
    let target = expression;
    while (ts.isParenthesizedExpression(target) || ts.isNonNullExpression(target)) target = target.expression;
    let symbol;
    if (ts.isElementAccessExpression(target)) {
        if (!ts.isStringLiteralLike(target.argumentExpression)) {
            return expressionSlot(target.expression, checker, elementDepth + 1, followAliases, via);
        }
        symbol = checker.getSymbolAtLocation(target.argumentExpression);
    } else if (ts.isCallExpression(target) || ts.isNewExpression(target)) {
        const signature = checker.getResolvedSignature(target);
        const declaration = signature?.declaration;
        if (declaration && !ts.isJSDocSignature(declaration) && ts.isFunctionLike(declaration)) return {declaration, elementDepth, via};
        return {unresolved: `${kindName(target.kind)} with no signature declaration`};
    } else if (ts.isIdentifier(target)) {
        symbol = checker.getSymbolAtLocation(target);
    } else if (ts.isPropertyAccessExpression(target)) {
        symbol = checker.getSymbolAtLocation(target.name);
    } else {
        return {unresolved: kindName(target.kind)};
    }
    if (!symbol) return {unresolved: `no symbol for ${clip(target.getText(), 60)}`};
    const found = declarationOfSymbol(symbol, checker);
    if (found.unresolved) return found;
    const declaration = found.declaration;
    if (followAliases && ts.isVariableDeclaration(declaration) && !declaration.type && declaration.initializer && via.length < 5) {
        const followed = expressionSlot(declaration.initializer, checker, elementDepth, true, [...via, nameText(declaration.name) ?? '<binding>']);
        // An initializer that leads nowhere (a literal, a `??`) leaves the alias itself as the receiver
        if (!followed.unresolved) return followed;
    }
    return {declaration, elementDepth, via};
}

function argumentSlot(call, argument, checker, elementDepth) {
    const signature = checker.getResolvedSignature(call);
    const declaration = signature?.declaration;
    if (!declaration || ts.isJSDocSignature(declaration)) return {unresolved: `${kindName(call.kind)} with no signature declaration`};
    const index = call.arguments.indexOf(argument);
    const parameters = declaration.parameters;
    if (index < parameters.length && !parameters[index].dotDotDotToken) return {declaration: parameters[index], elementDepth, via: []};
    const last = parameters[parameters.length - 1];
    if (!last || !last.dotDotDotToken) return {unresolved: `argument ${index} past ${parameters.length} parameters`};
    // A plain argument fills one element of the rest array; a spread argument stands for the rest array itself
    return {declaration: last, elementDepth: ts.isSpreadElement(argument) ? elementDepth : elementDepth + 1, via: []};
}

function objectPropertySlot(objectLiteral, nameNode, checker, elementDepth) {
    const contextual = checker.getContextualType(objectLiteral);
    if (!contextual) return {unresolved: 'object literal with no contextual type'};
    const name = ts.isIdentifier(nameNode) || ts.isStringLiteralLike(nameNode) || ts.isNumericLiteral(nameNode) ? nameNode.text : undefined;
    if (name === undefined) return {unresolved: `object literal property named by ${kindName(nameNode.kind)}`};
    const property = checker.getPropertyOfType(checker.getNonNullableType(contextual), name);
    if (!property) return {unresolved: `contextual type has no property ${name}`};
    const found = declarationOfSymbol(property, checker);
    if (found.unresolved) return found;
    return {declaration: found.declaration, elementDepth, via: []};
}

// A class member that no longer fits the interface or base member it implements. The class member's own type is
// the slot when every base member's type is already readonly; otherwise each base member's type is, found through
// the heritage clauses' types.
function heritageSlot(member, checker) {
    const classNode = member.parent;
    const name = nameText(member.name);
    if (!ts.isClassLike(classNode) || name === undefined) return {unresolved: `${kindName(member.kind)} outside a class`};
    const className = nameText(classNode.name) ?? '<class>';
    const bases = [];
    for (const clause of classNode.heritageClauses ?? []) {
        for (const typeExpression of clause.types) {
            const property = checker.getPropertyOfType(checker.getTypeAtLocation(typeExpression), name);
            for (const declaration of property?.declarations ?? []) {
                if (!bases.includes(declaration)) bases.push(declaration);
            }
        }
    }
    if (bases.length === 0) return {unresolved: `${className}.${name} has no base member`};
    if (bases.every(declaration => isReadonlyTypeNode(declaration.type))) {
        return {declaration: member, elementDepth: 0, via: [], heritage: `class side: ${className}.${name}, its base already readonly`};
    }
    return {list: bases.map(declaration => ({declaration, elementDepth: 0, via: [], heritage: `base side, for ${className}.${name}`}))};
}

// The receiving slot of a readonly value going into a mutable one, from the node the diagnostic sits on, walking up
function resolveReceivingSlot(errorNode, checker, code) {
    let node = errorNode;
    let elementDepth = 0;
    for (let step = 0; step < 64; step++) {
        if (ts.isReturnStatement(node)) {
            const fn = containingFunction(node);
            return fn ? {declaration: fn, elementDepth, via: []} : {unresolved: 'return outside a function'};
        }
        if ((ts.isShorthandPropertyAssignment(node) || ts.isPropertyAssignment(node)) && node.parent && ts.isObjectLiteralExpression(node.parent)) {
            return objectPropertySlot(node.parent, node.name, checker, elementDepth);
        }
        if ((ts.isVariableDeclaration(node) || ts.isPropertyDeclaration(node) || ts.isParameter(node)) && step === 0) {
            return {declaration: node, elementDepth, via: []};
        }
        const parent = node.parent;
        if (!parent) return {unresolved: `${kindName(node.kind)} with no parent`};
        // 2416 is a member not assignable to the same member in a base; a property with no initializer can only
        // miss through its base
        if (parent.name === node && ts.isClassLike(parent.parent) && (ts.isMethodDeclaration(parent) || ts.isGetAccessorDeclaration(parent)
            || (ts.isPropertyDeclaration(parent) && (code === 2416 || !parent.initializer)))) {
            return heritageSlot(parent, checker);
        }
        if ((ts.isVariableDeclaration(parent) || ts.isPropertyDeclaration(parent) || ts.isParameter(parent))
            && (parent.name === node || parent.initializer === node)) {
            return {declaration: parent, elementDepth, via: []};
        }
        if (ts.isBinaryExpression(parent) && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken) {
            return expressionSlot(parent.left, checker, elementDepth, false, []);
        }
        if (ts.isReturnStatement(parent)) {
            const fn = containingFunction(parent);
            return fn ? {declaration: fn, elementDepth, via: []} : {unresolved: 'return outside a function'};
        }
        if (ts.isArrowFunction(parent) && parent.body === node) return {declaration: parent, elementDepth, via: []};
        if ((ts.isCallExpression(parent) || ts.isNewExpression(parent)) && parent.arguments?.includes(node)) {
            return argumentSlot(parent, node, checker, elementDepth);
        }
        if ((ts.isPropertyAssignment(parent) && (parent.initializer === node || parent.name === node))
            || (ts.isShorthandPropertyAssignment(parent) && parent.name === node)) {
            return objectPropertySlot(parent.parent, parent.name, checker, elementDepth);
        }
        if (ts.isArrayLiteralExpression(parent)) {
            // An element sits one array level below the literal; a spread's operand stands for the literal itself
            if (!ts.isSpreadElement(node)) elementDepth++;
            node = parent;
            continue;
        }
        if (ts.isParenthesizedExpression(parent) || ts.isSpreadElement(parent)
            || (ts.isConditionalExpression(parent) && parent.condition !== node)
            || (ts.isBinaryExpression(parent) && passThroughOperators.has(parent.operatorToken.kind))) {
            node = parent;
            continue;
        }
        return {unresolved: `${kindName(node.kind)} under ${kindName(parent.kind)}`};
    }
    return {unresolved: 'walked 64 levels up'};
}

// A slot made or found from a resolution in the current program. A new Readonly slot's insertions are converted
// to untouched offsets now, while the current text still matches the program that reported it.
// Every slot made here is pushed onto `created`, alias slots too.
function slotFromResolution(resolution, describedDiagnostic, checker, created) {
    if (resolution.unresolved) {
        const key = `${describedDiagnostic.relative}:${describedDiagnostic.line} unresolved: ${resolution.unresolved}`;
        let slot = slots.get(key);
        if (!slot) {
            slot = newSlot({key, relative: describedDiagnostic.relative, line: describedDiagnostic.line, label: `unresolved: ${resolution.unresolved}`});
            slot.state = 'Unresolved';
            slot.name = resolution.unresolved;
            slots.set(key, slot);
        }
        return slot;
    }
    const identity = slotIdentity(resolution.declaration, resolution.elementDepth);
    const existing = slots.get(identity.key);
    if (existing) return existing;
    const slot = newSlot(identity);
    slots.set(identity.key, slot);
    created.push(slot);
    if (!currentText.has(identity.absolute)) {
        slot.state = 'Outside';
        slot.name = 'declared outside the copy of src';
        return slot;
    }
    slot.visibility = visibilityOf(resolution.declaration);
    // A Held slot is checked after its plan, so only a slot this script would have edited is Held
    if (droppedSlots.has(`${identity.relative} ${identity.label}`)) {
        slot.state = 'Dropped';
        slot.name = 'never edited, by Adamic\'s ruling';
        return slot;
    }
    const target = slotTypeNode(resolution.declaration, resolution.elementDepth);
    if (target.missing) {
        slot.state = 'Unannotated';
        slot.name = kindName(resolution.declaration.kind);
        return slot;
    }
    if (target.named) {
        slot.state = 'NamedType';
        slot.name = target.named;
        return slot;
    }
    const sourceFile = resolution.declaration.getSourceFile();
    const plan = newPlan(checker);
    planReadonly(target.typeNode, sourceFile, plan);
    slot.originalType = clip(target.typeNode.getText(sourceFile), 100);
    const aliasSlots = plan.aliases.map(declaration => aliasSlot(declaration, checker, created));
    if (plan.insertions.length > 0 && holdPublic && slot.visibility === 'Public') {
        holdSlot(slot, plan, identity.absolute, resolution.elementDepth);
    } else if (plan.insertions.length > 0) {
        slot.state = 'Readonly';
        slot.changed = plan.changed.map(text => clip(text, 60));
        slot.untouchedMembers = [...plan.named, ...plan.other.map(kind => `<${kind}>`), ...aliasSlots.map(alias => `alias slot ${alias.key}`)];
        slot.pendingInsertions = pendingFrom(plan, identity.absolute, slot, resolution.elementDepth);
        slot.typeSpan = typeSpanOf(target.typeNode, sourceFile, identity.absolute, resolution.elementDepth);
    } else if (aliasSlots.length > 0) {
        // The alias is edited as its own slot, and this slot's sites count under it
        slot.state = 'AliasRef';
        slot.aliasSlotKeys = aliasSlots.map(alias => alias.key);
        slot.name = `${slot.originalType}, edited as ${slot.aliasSlotKeys.join(', ')}`;
    } else if (plan.already) {
        slot.state = 'AlreadyReadonly';
        slot.name = slot.originalType;
    } else {
        slot.state = 'NamedType';
        slot.name = [...plan.named, ...plan.other.map(kind => `<${kind}>`)].join(', ') || '<nothing to edit>';
    }
    return slot;
}

function pendingFrom(plan, absolute, slot, defaultRank) {
    return plan.insertions.map(insertion => {
        if (/[\r\n\u2028\u2029]/.test(insertion.text)) throw new Error('an insertion may not hold a line break');
        return {absolute, original: toOriginal(absolute, insertion.position), text: insertion.text, rank: insertion.rank ?? defaultRank, slotKey: slot.key};
    });
}

// A Public slot under --hold-public: never edited, but its own insertions are kept so the table can say what its
// one line of readonly would change in the declaration file
function holdSlot(slot, plan, absolute, defaultRank) {
    slot.state = 'Held';
    slot.name = 'public, never edited under --hold-public';
    slot.changed = plan.changed.map(text => clip(text, 60));
    slot.heldInsertions = pendingFrom(plan, absolute, slot, defaultRank);
}

// A type alias declared under src, edited as its own slot of kind Alias: every mutable array inside its body
function aliasSlot(declaration, checker, created) {
    const sourceFile = declaration.getSourceFile();
    const absolute = path.resolve(sourceFile.fileName);
    const relative = relativeName(absolute);
    const line = sourceFile.getLineAndCharacterOfPosition(declaration.getStart(sourceFile)).line + 1;
    const label = `alias ${declaration.name.text}`;
    const key = `${relative}:${line} ${label}`;
    const existing = slots.get(key);
    if (existing) return existing;
    const slot = newSlot({key, absolute, relative, line, label});
    slot.kind = 'Alias';
    slots.set(key, slot);
    created.push(slot);
    slot.visibility = visibilityOf(declaration);
    slot.originalType = clip(declaration.type.getText(sourceFile), 100);
    const plan = newPlan(checker);
    planDeep(declaration.type, sourceFile, plan, 0);
    const nested = plan.aliases.filter(inner => inner !== declaration).map(inner => aliasSlot(inner, checker, created));
    if (plan.insertions.length > 0 && holdPublic && slot.visibility === 'Public') {
        holdSlot(slot, plan, absolute, 0);
    } else if (plan.insertions.length > 0) {
        slot.state = 'Readonly';
        slot.changed = plan.changed.map(text => clip(text, 60));
        slot.untouchedMembers = nested.map(alias => `alias slot ${alias.key}`);
        slot.pendingInsertions = pendingFrom(plan, absolute, slot, 0);
    } else {
        slot.state = 'NamedType';
        slot.name = `alias with no mutable array inside${nested.length ? `; names ${nested.map(alias => alias.key).join(', ')}` : ''}`;
    }
    return slot;
}

function newSlot(identity) {
    return {
        key: identity.key, relative: identity.relative, line: identity.line, label: identity.label, absolute: identity.absolute,
        kind: 'Slot', state: undefined, visibility: undefined, name: undefined, originalType: undefined, changed: [], untouchedMembers: [],
        pendingInsertions: [], insertions: [], writes: [], diagnosticSites: new Set(), notes: new Set(), round: currentRound, seed: false,
        typeSpan: undefined, aliasSlotKeys: [], heldInsertions: [], blockedBy: undefined, blockedVia: undefined, blockedAt: undefined,
        blockedRound: undefined, listedSiteNames: [],
    };
}

// Returns true, or a description of the clash, leaving the slot unedited, when another live insertion holds the
// same offset and rank, since then their order in the text would be a guess
function applyInsertions(slot) {
    for (const insertion of slot.pendingInsertions) {
        const clash = insertionsOf(insertion.absolute).find(other => other.original === insertion.original && other.rank === insertion.rank);
        if (clash) {
            slot.pendingInsertions = [];
            return `its insertion at ${insertion.original} clashes with ${clash.slotKey}'s`;
        }
    }
    for (const insertion of slot.pendingInsertions) {
        const list = insertionsByFile.get(insertion.absolute) ?? [];
        list.push(insertion);
        sortInsertions(list);
        insertionsByFile.set(insertion.absolute, list);
        slot.insertions.push(insertion);
    }
    slot.pendingInsertions = [];
    return true;
}

// Before a revert, the slot's declaration found again in the current program must hold each of its insertions
function checkRevert(slot, declaration) {
    const sourceFile = declaration.getSourceFile();
    const absolute = path.resolve(sourceFile.fileName);
    const text = currentText.get(absolute);
    for (const insertion of slot.insertions) {
        const at = toCurrent(absolute, insertion);
        if (text.slice(at, at + insertion.text.length) !== insertion.text || at < declaration.getStart(sourceFile) || at >= declaration.end) {
            throw new Error(`${slot.key}: the insertion "${insertion.text}" is not where its declaration is in the current program`);
        }
    }
}

function revertInsertions(slot) {
    const touched = new Set();
    for (const insertion of slot.insertions) {
        const list = insertionsByFile.get(insertion.absolute) ?? [];
        const index = list.indexOf(insertion);
        if (index < 0) throw new Error(`${slot.key}: reverting an insertion that is not live`);
        list.splice(index, 1);
        touched.add(insertion.absolute);
    }
    slot.insertions = [];
    return touched;
}

// The receiver a write diagnostic writes through, or undefined when it isn't one of the write shapes
function writeReceiver(record, errorNode) {
    if (record.code === 2339 || record.code === 2540) {
        let access;
        if (ts.isIdentifier(errorNode) && errorNode.parent && ts.isPropertyAccessExpression(errorNode.parent) && errorNode.parent.name === errorNode) {
            access = errorNode.parent;
        } else if (ts.isPropertyAccessExpression(errorNode)) {
            access = errorNode;
        }
        if (!access) return undefined;
        const property = access.name.text;
        if (record.code === 2339 && !writeMethods.has(property)) return undefined;
        if (record.code === 2540 && property !== 'length') return undefined;
        return {receiver: access.expression, property};
    }
    if (record.code === 2542) {
        for (let node = errorNode, step = 0; node && step < 4; node = node.parent, step++) {
            if (ts.isElementAccessExpression(node)) return {receiver: node.expression, property: '[index]'};
        }
    }
    return undefined;
}

// The one test for a readonly value going into a mutable slot, for the rounds and for blocking alike
function isReadonlyIntoMutable(record) {
    return record.code === 4104 || record.message.includes(readonlyPhrase);
}

const sharedArrayNames = new Set(['emptyArray', 'resolvingEmptyArray']);
function skipParenthesesAndAssertions(expression) {
    let target = expression;
    while (ts.isParenthesizedExpression(target) || ts.isNonNullExpression(target)) target = target.expression;
    return target;
}
function resolvedSymbolOf(node, checker) {
    let symbol = checker.getSymbolAtLocation(node);
    if (symbol && symbol.flags & ts.SymbolFlags.Alias) symbol = checker.getAliasedSymbol(symbol);
    return symbol;
}
// An expression naming emptyArray or resolvingEmptyArray as declared in core.ts or utilities.ts
function isSharedArray(expression, checker) {
    const target = skipParenthesesAndAssertions(expression);
    const name = ts.isIdentifier(target) ? target : ts.isPropertyAccessExpression(target) ? target.name : undefined;
    if (!name || !sharedArrayNames.has(name.text)) return false;
    const declaration = resolvedSymbolOf(name, checker)?.valueDeclaration;
    if (!declaration || !ts.isVariableDeclaration(declaration)) return false;
    return ['src/compiler/core.ts', 'src/compiler/utilities.ts'].includes(relativeName(path.resolve(declaration.getSourceFile().fileName)));
}

// The slot at the end of a chain of blocking: a Held or Dropped slot is its own root, a Blocked slot names its
// root, and an alias reference reads through to its alias slot
function rootOf(slot) {
    const effective = slot && slot.state === 'AliasRef' ? slots.get(slot.aliasSlotKeys[0]) : slot;
    if (!effective) return undefined;
    if (effective.state === 'Held' || effective.state === 'Dropped') return effective.key;
    if (effective.state === 'Blocked') return effective.blockedBy;
    return undefined;
}
// A resolution's slot when this script already holds one; blocking never makes a slot
function existingSlotOf(resolution, record) {
    if (resolution.unresolved) return slots.get(`${record.relative}:${record.line} unresolved: ${resolution.unresolved}`);
    return slots.get(slotIdentity(resolution.declaration, resolution.elementDepth).key);
}

// The value a readonly-into-mutable diagnostic carries, read from the node it sits on the way resolveReceivingSlot
// reads the receiver: a return's expression, a declaration's initializer, an assignment's right side, a property's
// initializer, a shorthand property's value, a class member that misses its base, or the expression itself
function diagnosticValue(errorNode, code) {
    if (ts.isReturnStatement(errorNode)) {
        return errorNode.expression ? {expression: errorNode.expression} : {unresolved: 'return with no expression'};
    }
    if (ts.isShorthandPropertyAssignment(errorNode)) return {shorthand: errorNode};
    if (ts.isPropertyAssignment(errorNode)) return {expression: errorNode.initializer};
    if (ts.isVariableDeclaration(errorNode) || ts.isParameter(errorNode) || ts.isPropertyDeclaration(errorNode)) {
        return errorNode.initializer ? {expression: errorNode.initializer} : {unresolved: `${kindName(errorNode.kind)} with no initializer`};
    }
    const parent = errorNode.parent;
    if (parent && parent.name === errorNode) {
        // The same test resolveReceivingSlot makes: then the member itself is the value its base won't take
        if (ts.isClassLike(parent.parent) && (ts.isMethodDeclaration(parent) || ts.isGetAccessorDeclaration(parent)
            || (ts.isPropertyDeclaration(parent) && (code === 2416 || !parent.initializer)))) {
            return {member: parent};
        }
        if (ts.isShorthandPropertyAssignment(parent)) return {shorthand: parent};
        if (ts.isVariableDeclaration(parent) || ts.isParameter(parent) || ts.isPropertyDeclaration(parent) || ts.isPropertyAssignment(parent)) {
            return parent.initializer ? {expression: parent.initializer} : {unresolved: `${kindName(parent.kind)} with no initializer`};
        }
    }
    if (parent && ts.isBinaryExpression(parent) && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken && parent.left === errorNode) {
        return {expression: parent.right};
    }
    if (ts.isExpression(errorNode)) return {expression: errorNode};
    return {unresolved: `the value at a ${kindName(errorNode.kind)}`};
}

// The expressions a value can come from: through parentheses, `as`, `satisfies`, `!` and a spread argument, both
// arms of `?:`, both operands of `||` and `??`, and the plain elements of an array literal (a spread inside one
// copies, so its operand's readonly doesn't flow on)
function valueLeaves(expression, leaves) {
    let target = expression;
    while (ts.isParenthesizedExpression(target) || ts.isNonNullExpression(target) || ts.isAsExpression(target)
        || ts.isSatisfiesExpression(target) || ts.isTypeAssertionExpression(target) || ts.isSpreadElement(target)) {
        target = target.expression;
    }
    if (ts.isConditionalExpression(target)) {
        valueLeaves(target.whenTrue, leaves);
        valueLeaves(target.whenFalse, leaves);
    } else if (ts.isBinaryExpression(target) && passThroughOperators.has(target.operatorToken.kind)) {
        valueLeaves(target.left, leaves);
        valueLeaves(target.right, leaves);
    } else if (ts.isArrayLiteralExpression(target)) {
        for (const element of target.elements) {
            if (!ts.isSpreadElement(element) && !ts.isOmittedExpression(element)) valueLeaves(element, leaves);
        }
    } else {
        leaves.push(target);
    }
    return leaves;
}

// A resolved source, or, for an unannotated variable declared in the copy of src, the sources of its initializer
// (up to five deep, as expressionSlot follows aliases), so `const tags = a || emptyArray` reaches both
function followSource(resolution, leaf, checker, via, found) {
    const declaration = resolution.declaration;
    if (ts.isVariableDeclaration(declaration) && !declaration.type && declaration.initializer && via.length < 5
        && currentText.has(path.resolve(declaration.getSourceFile().fileName))) {
        const followed = [...via, nameText(declaration.name) ?? '<binding>'];
        for (const inner of valueLeaves(declaration.initializer, [])) leafSources(inner, checker, resolution.elementDepth, followed, found);
        return;
    }
    found.push({resolution, leaf, via});
}
function leafSources(leaf, checker, elementDepth, via, found) {
    if (isSharedArray(leaf, checker)) {
        found.push({sentinel: true, leaf, via});
        return;
    }
    const resolution = expressionSlot(leaf, checker, elementDepth, false, via);
    if (resolution.unresolved) {
        found.push({unresolved: resolution.unresolved, leaf, via});
        return;
    }
    followSource(resolution, leaf, checker, via, found);
}

// A node's site, in a held file by its untouched offsets, otherwise as the program has it
function siteOfAnyNode(node) {
    const sourceFile = node.getSourceFile();
    if (currentText.has(path.resolve(sourceFile.fileName))) return siteOfNode(node);
    const position = sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile));
    return {site: `${sourceFile.fileName}:${position.line + 1}:${position.character + 1}`, line: position.line + 1, source: ''};
}

// Blocking, under --hold-public. Each readonly value still going into a Held, Dropped or Blocked slot once the rounds converge: when
// its source is a slot this script made readonly, the slot is marked Blocked under the receiver's root and put back
// after the round; when its source is the shared array itself, the site is sentinel residue under that root.
const sentinelResidue = new Map();
const unhandledSources = new Map();
// Returns true when the receiver has a root, so the diagnostic is blocking's to read
function blockFrom(record, checker, step, blocking) {
    const sourceFile = program.getSourceFile(record.absolute);
    const errorNode = errorNodeAt(sourceFile, record.diagnostic.start, record.diagnostic.length ?? 0);
    const resolution = resolveReceivingSlot(errorNode, checker, record.code);
    let receiver;
    let root;
    for (const each of resolution.list ?? [resolution]) {
        const slot = existingSlotOf(each, record);
        const eachRoot = rootOf(slot);
        if (eachRoot) {
            receiver = slot;
            root = eachRoot;
            break;
        }
    }
    // A residue into any other slot (a named type, an unannotated one) stays as the rounds left it
    if (!root) return false;
    const value = diagnosticValue(errorNode, record.code);
    const found = [];
    if (value.unresolved) {
        found.push({unresolved: value.unresolved, leaf: errorNode, via: []});
    } else if (value.member) {
        found.push({resolution: {declaration: value.member, elementDepth: 0, via: []}, leaf: value.member.name, via: []});
    } else if (value.shorthand) {
        const symbol = checker.getShorthandAssignmentValueSymbol(value.shorthand);
        const declared = symbol ? declarationOfSymbol(symbol, checker) : {unresolved: 'no value symbol for a shorthand property'};
        if (declared.unresolved) found.push({unresolved: declared.unresolved, leaf: value.shorthand.name, via: []});
        else followSource({declaration: declared.declaration, elementDepth: 0, via: []}, value.shorthand.name, checker, [], found);
    } else {
        for (const leaf of valueLeaves(value.expression, [])) leafSources(leaf, checker, 0, [], found);
    }
    let handled = false;
    const reasons = [];
    for (const source of found) {
        const viaText = source.via.length ? ` via ${source.via.join(', ')}` : '';
        if (source.unresolved) {
            reasons.push(`${clip(source.leaf.getText(), 60)}: unresolved, ${source.unresolved}${viaText}`);
            continue;
        }
        const slot = source.sentinel ? undefined : slots.get(slotIdentity(source.resolution.declaration, source.resolution.elementDepth).key);
        if (source.sentinel || slot?.seed) {
            const leafFile = source.leaf.getSourceFile();
            const leafAbsolute = path.resolve(leafFile.fileName);
            const leafStart = currentText.has(leafAbsolute) ? toOriginal(leafAbsolute, source.leaf.getStart(leafFile)) : undefined;
            const where = siteOfAnyNode(source.leaf);
            sentinelResidue.set(`${record.relative}|${record.originalStart}|${where.site}`, {
                site: where.site, source: where.source, diagnosticSite: record.site, relative: record.relative, originalStart: record.originalStart,
                originalEnd: record.originalEnd, leafRelative: relativeName(leafAbsolute), leafStart, receiver: receiver.key, root, via: source.via, lastStep: step,
            });
            handled = true;
            continue;
        }
        if (slot && slot.state === 'Readonly') {
            // The declaration was found again in this program, so the revert checks its insertions against it
            checkRevert(slot, source.resolution.declaration);
            slot.state = 'Blocked';
            slot.blockedBy = root;
            slot.blockedVia = receiver.key;
            slot.blockedAt = record.site;
            slot.blockedRound = step;
            blocking.set(slot.key, slot);
            handled = true;
            continue;
        }
        if (slot && blocking.has(slot.key)) {
            handled = true;
            continue;
        }
        const sourceKey = slot ? slot.key : slotIdentity(source.resolution.declaration, source.resolution.elementDepth).key;
        reasons.push(`${clip(source.leaf.getText(), 60)}: ${sourceKey}${viaText} is ${slot ? slot.state : 'no slot of this script'}`);
    }
    if (!handled) {
        unhandledSources.set(record.key, {site: record.site, code: record.code, receiver: receiver.key, root,
            reasons: reasons.length ? reasons : ['no source found'], source: clip(lineText(originalText.get(record.absolute), record.line - 1)), lastStep: step});
    }
    return true;
}

// 3. The baseline
log('building the baseline program');
let program = buildProgram(undefined);
const baselineRecords = allDiagnostics(program);
const baselineKeys = new Set(baselineRecords.map(record => record.key));
log(`baseline: ${baselineRecords.length} diagnostics, ${program.getSourceFiles().length} source files in the program`);

// 4. The seed
function findSeed(relativeFile, name) {
    const absolute = path.join(outSource, relativeFile);
    const sourceFile = program.getSourceFile(absolute);
    if (!sourceFile) throw new Error(`src/${relativeFile} is not in the program`);
    for (const statement of sourceFile.statements) {
        if (!ts.isVariableStatement(statement) || !hasModifier(statement, ts.SyntaxKind.ExportKeyword)) continue;
        for (const declaration of statement.declarationList.declarations) {
            if (ts.isIdentifier(declaration.name) && declaration.name.text === name) return declaration;
        }
    }
    throw new Error(`no exported variable statement declaring ${name} in src/${relativeFile}`);
}
const seedSlots = [];
for (const [relativeFile, name] of [['compiler/core.ts', 'emptyArray'], ['compiler/utilities.ts', 'resolvingEmptyArray']]) {
    const declaration = findSeed(relativeFile, name);
    const slot = slotFromResolution({declaration, elementDepth: 0, via: []}, undefined, program.getTypeChecker(), []);
    if (slot.state !== 'Readonly') {
        throw new Error(`${name} in src/${relativeFile} has type ${slot.originalType ?? '<none>'}, state ${slot.state}: expected a mutable array to make readonly`);
    }
    slot.seed = true;
    seedSlots.push(slot);
}
const touchedBySeed = new Set();
for (const slot of seedSlots) {
    const applied = applyInsertions(slot);
    if (applied !== true) throw new Error(`the seed ${slot.key} couldn't be applied: ${applied}`);
    touchedBySeed.add(slot.absolute);
}
for (const absolute of touchedBySeed) currentText.set(absolute, render(absolute));
log(`seeded: ${seedSlots.map(slot => `${slot.key} is now ${slot.changed.map(text => `readonly ${text}`).join(', ')}`).join('; ')}`);

// 5. The rounds
const unattributedWrites = new Map();
const unclassified = new Map();
const roundSummaries = [];
let finalRecords = [];
let rounds = 0;
let converged = false;
for (let round = 1; round <= maximumRounds; round++) {
    rounds = round;
    currentRound = round;
    const roundBegan = Date.now();
    program = buildProgram(program);
    const checker = program.getTypeChecker();
    const fresh = allDiagnostics(program).filter(record => !baselineKeys.has(record.key));
    const created = [];
    const reverting = new Map();
    let readonlyCount = 0;
    let writeCount = 0;
    let otherCount = 0;
    for (const record of fresh) {
        try {
            const kind = classify(record, round, checker, created, reverting);
            if (kind === 'Readonly') readonlyCount++;
            else if (kind === 'Write') writeCount++;
            else otherCount++;
        } catch (error) {
            scriptErrors.push(`${record.site} code ${record.code}: ${error && error.stack ? error.stack.split('\n').slice(0, 3).join(' / ') : String(error)}`);
            console.error(`fixpoint: classifying ${record.site} code ${record.code} threw: ${error && error.stack ? error.stack : error}`);
            otherCount++;
        }
    }
    finalRecords = fresh;
    // Every offset above was converted while the text still matched this program; only now does the text move
    const touched = new Set();
    for (const slot of reverting.values()) {
        for (const absolute of revertInsertions(slot)) touched.add(absolute);
        if (slot.state !== 'WrittenUnion') continue;
        // Adamic ruled that written slots stay refused in the build. This union exists only so the fixpoint converges
        // and counts: the slot still takes the shared array, and its writes stay errors under the union.
        const span = slot.typeSpan;
        const prefix = `${span.outer ? '(' : ''}readonly never[] | ${span.inner ? '(' : ''}`;
        const suffix = `${span.inner ? ')' : ''}${span.outer ? ')' : ''}`;
        slot.pendingInsertions = [{absolute: slot.absolute, original: span.start, text: prefix, rank: span.rank, slotKey: slot.key}];
        if (suffix) slot.pendingInsertions.push({absolute: slot.absolute, original: span.end, text: suffix, rank: span.rank, slotKey: slot.key});
        const applied = applyInsertions(slot);
        if (applied !== true) {
            scriptErrors.push(`${slot.key}: its union couldn't be applied, ${applied}; left as written and put back`);
            slot.state = 'Written';
        }
    }
    const forward = [];
    for (const slot of created) {
        if (slot.state !== 'Readonly' || slot.pendingInsertions.length === 0) continue;
        const applied = applyInsertions(slot);
        if (applied === true) {
            forward.push(slot);
            touched.add(slot.absolute);
        } else {
            scriptErrors.push(`${slot.key}: ${applied}; left unedited`);
            slot.state = 'Unresolved';
            slot.name = `insertion clash: ${applied}`;
        }
    }
    for (const absolute of touched) currentText.set(absolute, render(absolute));
    const summary = {round, fresh: fresh.length, readonlyCount, writeCount, otherCount, madeReadonly: forward.length, written: reverting.size,
        newSlots: created.length, seconds: (Date.now() - roundBegan) / 1000};
    roundSummaries.push(summary);
    log(`round ${round}: ${fresh.length} new diagnostics (${readonlyCount} readonly into mutable, ${writeCount} writes, ${otherCount} other), `
        + `${created.length} new slots, ${forward.length} made readonly, ${reverting.size} written, ${summary.seconds.toFixed(1)} s`);
    if (forward.length === 0 && reverting.size === 0) {
        converged = true;
        break;
    }
}

function classify(record, round, checker, created, reverting) {
    const isReadonlyClass = isReadonlyIntoMutable(record);
    const sourceFile = record.absolute ? program.getSourceFile(record.absolute) : undefined;
    if (isReadonlyClass && sourceFile && record.held) {
        record.kind = 'Readonly';
        const errorNode = errorNodeAt(sourceFile, record.diagnostic.start, record.diagnostic.length ?? 0);
        const resolution = resolveReceivingSlot(errorNode, checker, record.code);
        // A class member that misses its base can resolve to several base members; the first holds the site
        let primary;
        for (const each of resolution.list ?? [resolution]) {
            const slot = slotFromResolution(each, record, checker, created);
            if (each.heritage) slot.notes.add(each.heritage);
            slot.diagnosticSites.add(record.site);
            primary ??= slot;
        }
        record.slotKey = primary.key;
        const diagnosticKey = `${record.relative}|${record.originalStart}|${record.code}`;
        if (!readonlyDiagnostics.has(diagnosticKey)) {
            readonlyDiagnostics.set(diagnosticKey, {relative: record.relative, absolute: record.absolute, originalStart: record.originalStart,
                originalEnd: record.originalEnd, line: record.line, site: record.site, slotKey: primary.key, round});
        }
        return 'Readonly';
    }
    if ((record.code === 2339 || record.code === 2540 || record.code === 2542) && sourceFile && record.held) {
        const errorNode = errorNodeAt(sourceFile, record.diagnostic.start, record.diagnostic.length ?? 0);
        const write = writeReceiver(record, errorNode);
        if (write) {
            record.kind = 'Write';
            const source = clip(lineText(originalText.get(record.absolute), record.line - 1));
            const target = expressionSlot(write.receiver, checker, 0, true, []);
            let slot;
            if (!target.unresolved) slot = slots.get(slotIdentity(target.declaration, target.elementDepth).key);
            // A WrittenUnion slot's writes stay errors under its union every round, so they are recorded once each
            const live = slot && (slot.state === 'Readonly' || slot.state === 'WrittenUnion' || (slot.state === 'Written' && reverting.has(slot.key)));
            if (live) {
                if (slot.state === 'Readonly') {
                    checkRevert(slot, target.declaration);
                    // An alias slot, or one with no span, is put back; any other written slot converges as a union
                    slot.state = slot.kind === 'Alias' || !slot.typeSpan ? 'Written' : 'WrittenUnion';
                    slot.writtenRound = round;
                    reverting.set(slot.key, slot);
                }
                if (!slot.writes.some(each => each.site === record.site)) {
                    slot.writes.push({site: record.site, code: record.code, property: write.property, via: target.via, source});
                }
            } else {
                const receiverDeclaration = target.unresolved
                    ? `unresolved: ${target.unresolved}`
                    : `${slotIdentity(target.declaration, target.elementDepth).key}${slot ? ` (${slot.state})` : ''}`;
                unattributedWrites.set(record.key, {site: record.site, code: record.code, property: write.property,
                    receiver: clip(write.receiver.getText(sourceFile), 60), receiverDeclaration, via: target.via ?? [], source, lastRound: round});
            }
            return 'Write';
        }
    }
    record.kind = 'Other';
    unclassified.set(record.key, {site: record.site, code: record.code, message: record.message.replace(/\s+/g, ' ').slice(0, 200), lastRound: round});
    return 'Other';
}

// The blocking rounds, under --hold-public: each reads the program, blocks the sources of every readonly value still
// going into a Held, Dropped or Blocked slot, then puts those slots back, until a round blocks nothing. Putting a
// slot back makes its own sources flow readonly into it, which the next round reads.
const blockRoundSummaries = [];
let blockRounds = 0;
let blockSettled = false;
if (holdPublic && !converged) log('blocking skipped: the rounds did not converge');
if (holdPublic && converged) {
    for (let blockRound = 1; blockRound <= maximumBlockRounds; blockRound++) {
        blockRounds = blockRound;
        // Steps continue the rounds' count, so "seen in the final round" reads the same across both phases
        const step = rounds + blockRound;
        currentRound = step;
        const roundBegan = Date.now();
        // The converged round changed no text, so the program it built is the first blocking round's
        if (blockRound > 1) program = buildProgram(program);
        const checker = program.getTypeChecker();
        const fresh = allDiagnostics(program).filter(record => !baselineKeys.has(record.key));
        const blocking = new Map();
        let readonlyCount = 0;
        let rootedCount = 0;
        let otherCount = 0;
        for (const record of fresh) {
            try {
                if (isReadonlyIntoMutable(record) && record.held && program.getSourceFile(record.absolute)) {
                    record.kind = 'Readonly';
                    readonlyCount++;
                    if (blockFrom(record, checker, step, blocking)) rootedCount++;
                    continue;
                }
                // Writes and the rest are read as the rounds read them, so the final tables stay whole. Putting a
                // slot back only takes readonly away, so nothing here should make a slot or find one written.
                const created = [];
                const reverting = new Map();
                classify(record, step, checker, created, reverting);
                if (created.length > 0 || reverting.size > 0) {
                    scriptErrors.push(`${record.site} code ${record.code}: blocking round ${blockRound} made ${created.length} slots and found ${reverting.size} written, which converged rounds should not`);
                }
                otherCount++;
            } catch (error) {
                scriptErrors.push(`blocking round ${blockRound}, ${record.site} code ${record.code}: ${error && error.stack ? error.stack.split('\n').slice(0, 3).join(' / ') : String(error)}`);
                console.error(`fixpoint: blocking at ${record.site} code ${record.code} threw: ${error && error.stack ? error.stack : error}`);
                otherCount++;
            }
        }
        finalRecords = fresh;
        // Every check above ran while the text still matched this program; only now does the text move
        const touched = new Set();
        for (const slot of blocking.values()) {
            for (const absolute of revertInsertions(slot)) touched.add(absolute);
        }
        for (const absolute of touched) currentText.set(absolute, render(absolute));
        const residue = [...sentinelResidue.values()].filter(row => row.lastStep === step).length;
        const unhandled = [...unhandledSources.values()].filter(row => row.lastStep === step).length;
        const summary = {round: blockRound, fresh: fresh.length, readonlyCount, rootedCount, otherCount, blocked: blocking.size, residue, unhandled,
            seconds: (Date.now() - roundBegan) / 1000};
        blockRoundSummaries.push(summary);
        log(`blocking round ${blockRound}: ${fresh.length} new diagnostics (${readonlyCount} readonly into mutable, ${rootedCount} into a held, dropped `
            + `or blocked slot), ${blocking.size} blocked, ${residue} sentinel residue, ${unhandled} unhandled, ${summary.seconds.toFixed(1)} s`);
        if (blocking.size === 0) {
            blockSettled = true;
            break;
        }
    }
    if (!blockSettled) {
        log(`FAIL: blocking still put slots back after ${maximumBlockRounds} rounds`);
        // The last round moved the text, and the sections below read node positions against it
        program = buildProgram(program);
    }
}
// The last step any round read, for what counts as present in the final round
const finalStep = rounds + blockRounds;

// The final texts go to disk, so the copy under out/src is the fixpoint
for (const [absolute, text] of currentText) {
    if (text !== originalText.get(absolute)) fs.writeFileSync(absolute, text);
}

// 7. The declaration-file lines for public slots left readonly, or under --hold-public for every file the final
// edits touch
function changedLineCount(before, after) {
    const counts = new Map();
    for (const line of before.split('\n')) counts.set(line, (counts.get(line) ?? 0) + 1);
    for (const line of after.split('\n')) counts.set(line, (counts.get(line) ?? 0) - 1);
    let total = 0;
    const lines = [];
    for (const [line, count] of counts) {
        if (count === 0) continue;
        total += Math.abs(count);
        lines.push(`${count > 0 ? '-' : '+'}${line}`);
    }
    return {total, lines};
}
const declarationOptions = {...baseOptions, stripInternal: true, isolatedDeclarations: true, declarationMap: false, sourceMap: false};
// An untouched file's declaration text is read once, since the held slots' previews reuse it
const untouchedDeclarations = new Map();
function declarationDiff(absolute, after) {
    const before = originalText.get(absolute);
    try {
        let diff;
        if (absolute.endsWith('.d.ts')) {
            diff = changedLineCount(before, after);
        } else {
            const fileName = absolute;
            if (!untouchedDeclarations.has(absolute)) {
                untouchedDeclarations.set(absolute, ts.transpileDeclaration(before, {compilerOptions: declarationOptions, fileName}).outputText);
            }
            const declarationAfter = ts.transpileDeclaration(after, {compilerOptions: declarationOptions, fileName}).outputText;
            diff = changedLineCount(untouchedDeclarations.get(absolute), declarationAfter);
        }
        return {relative: relativeName(absolute), total: diff.total, lines: diff.lines};
    } catch (error) {
        return {relative: relativeName(absolute), total: 0, lines: [], error: clip(String(error && error.message ? error.message : error), 200)};
    }
}
const publicFiles = [...new Set([...slots.values()].filter(slot => slot.state === 'Readonly' && slot.visibility === 'Public').map(slot => slot.absolute))].sort();
const editedFiles = [...insertionsByFile.keys()].filter(absolute => insertionsOf(absolute).length > 0).sort();
const declarationDiffs = (holdPublic ? editedFiles : publicFiles).map(absolute => declarationDiff(absolute, currentText.get(absolute)));
const declarationLineTotal = declarationDiffs.reduce((sum, diff) => sum + diff.total, 0);
const declarationErrors = declarationDiffs.filter(diff => diff.error).length;
if (holdPublic) log(`declaration check: ${declarationLineTotal} changed lines over ${declarationDiffs.length} edited files, ${declarationErrors} that threw`);

// 8. The sentinels
const sentinelPattern = /(?:[!=]==?\s*(?:ts\.)?(?:emptyArray|resolvingEmptyArray)\b)|(?:\b(?:emptyArray|resolvingEmptyArray)\s*[!=]==?)/;
const sentinels = [];
for (const absolute of [...currentText.keys()].sort()) {
    const text = currentText.get(absolute);
    if (!text.includes('mptyArray')) continue;
    const starts = lineStartsOf(text);
    for (let index = 0; index < starts.length; index++) {
        const line = lineText(text, index);
        if (sentinelPattern.test(line)) sentinels.push({site: `${relativeName(absolute)}:${index + 1}`, source: clip(line)});
    }
}

// 9. The listed sites
const finalResidue = [...sentinelResidue.values()].filter(row => row.lastStep === finalStep);
const finalUnhandled = [...unhandledSources.values()].filter(row => row.lastStep === finalStep);
const siteRows = [];
for (const raw of fs.readFileSync(sitesPath, 'utf8').split('\n')) {
    if (!raw.trim()) continue;
    const tab = raw.indexOf('\t');
    const site = tab < 0 ? raw : raw.slice(0, tab);
    const words = tab < 0 ? '' : raw.slice(tab + 1);
    const match = /^(.*):(\d+):(\d+)$/.exec(site);
    if (!match) {
        siteRows.push({site, words, group: 'unmatched', note: 'not file:line:col'});
        continue;
    }
    const relative = match[1];
    const line = Number(match[2]);
    const column = Number(match[3]);
    const absolute = path.join(outDirectory, relative);
    const text = originalText.get(absolute);
    if (text === undefined) {
        siteRows.push({site, words, group: 'unmatched', note: 'file not in the copy'});
        continue;
    }
    const starts = lineStartsOf(text);
    if (line < 1 || line > starts.length) {
        siteRows.push({site, words, group: 'unmatched', note: 'line past the file'});
        continue;
    }
    const offset = starts[line - 1] + column - 1;
    const candidates = [...readonlyDiagnostics.values()].filter(diagnostic => diagnostic.relative === relative);
    const covering = candidates.find(diagnostic => diagnostic.originalStart <= offset && (offset < diagnostic.originalEnd || offset === diagnostic.originalStart))
        ?? candidates.find(diagnostic => diagnostic.line === line);
    // Sentinel residue claims the site first: its shared array, its diagnostic's span, or the diagnostic that covers it
    const residue = finalResidue.find(row => (row.leafRelative === relative && row.leafStart === offset)
        || (row.relative === relative && ((row.originalStart <= offset && offset < row.originalEnd) || covering?.originalStart === row.originalStart)));
    if (!covering && !residue) {
        siteRows.push({site, words, group: 'unmatched', note: 'no readonly-into-mutable diagnostic here'});
        continue;
    }
    const slotKey = covering ? covering.slotKey : residue.receiver;
    const slot = slots.get(slotKey);
    const byLine = covering ? !(covering.originalStart <= offset && offset < covering.originalEnd) : false;
    if (residue) {
        siteRows.push({site, words, group: `Sentinel residue by ${residue.root}`, root: residue.root, slotKey, residue: residue.site, byLine});
    } else {
        siteRows.push({site, words, group: siteGroupOf(slot), root: rootOf(slot), slotKey, byLine});
    }
    if (slot) {
        slot.listedSites = (slot.listedSites ?? 0) + 1;
        slot.listedSiteNames.push(site);
    }
    // A slot whose type is an alias edited on its own counts its sites under that alias slot too
    for (const aliasKey of slot?.aliasSlotKeys ?? []) {
        const alias = slots.get(aliasKey);
        if (alias) {
            alias.listedSites = (alias.listedSites ?? 0) + 1;
            alias.listedSiteNames.push(site);
        }
    }
}
// WrittenUnion counts as Written, a Blocked slot by its root; an alias reference counts as its alias slot's state
function siteGroupOf(slot) {
    if (!slot) return 'unmatched';
    if (slot.state === 'AliasRef') {
        const alias = slots.get(slot.aliasSlotKeys[0]);
        return alias ? `${siteGroupOf(alias)} (alias)` : 'AliasRef';
    }
    if (slot.state === 'Blocked') return `Blocked by ${slot.blockedBy}`;
    return slot.state === 'WrittenUnion' ? 'Written' : slot.state;
}
const siteGroups = new Map();
for (const row of siteRows) siteGroups.set(row.group, (siteGroups.get(row.group) ?? 0) + 1);

// The final program is the fixpoint's when the rounds converged: the last round built it and changed nothing
const finalChecker = program.getTypeChecker();
const heldProgramFiles = program.getSourceFiles().filter(sourceFile => currentText.has(path.resolve(sourceFile.fileName)));
const comparisonOperators = new Set([ts.SyntaxKind.EqualsEqualsEqualsToken, ts.SyntaxKind.ExclamationEqualsEqualsToken,
    ts.SyntaxKind.EqualsEqualsToken, ts.SyntaxKind.ExclamationEqualsToken]);

function isWriteTarget(node) {
    const parent = node.parent;
    if (ts.isBinaryExpression(parent) && parent.left === node
        && parent.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && parent.operatorToken.kind <= ts.SyntaxKind.LastAssignment) {
        return true;
    }
    if ((ts.isPrefixUnaryExpression(parent) || ts.isPostfixUnaryExpression(parent))
        && (parent.operator === ts.SyntaxKind.PlusPlusToken || parent.operator === ts.SyntaxKind.MinusMinusToken)) {
        return true;
    }
    return ts.isDeleteExpression(parent);
}
// How a reference writes through its array, or undefined when it reads it or assigns a whole value to it
function referenceWrite(reference) {
    let outer = reference;
    while (ts.isParenthesizedExpression(outer.parent) || ts.isNonNullExpression(outer.parent)) outer = outer.parent;
    const parent = outer.parent;
    if (ts.isPropertyAccessExpression(parent) && parent.expression === outer) {
        if (writeMethods.has(parent.name.text) && ts.isCallExpression(parent.parent) && parent.parent.expression === parent) return `.${parent.name.text}()`;
        if (parent.name.text === 'length' && isWriteTarget(parent)) return 'length =';
    }
    if (ts.isElementAccessExpression(parent) && parent.expression === outer && isWriteTarget(parent)) return '[i] =';
    return undefined;
}
function siteOfNode(node) {
    const sourceFile = node.getSourceFile();
    const absolute = path.resolve(sourceFile.fileName);
    const {line, column} = originalLineColumn(absolute, toOriginal(absolute, node.getStart(sourceFile)));
    return {site: `${relativeName(absolute)}:${line}:${column}`, line, source: clip(lineText(originalText.get(absolute), line - 1))};
}

// 6. Sentinels by slot: each comparison with the shared array, the other operand's declaration, and every reference
// to that symbol in the program
const sentinelComparisons = [];
const sentinelTargets = new Map();
for (const sourceFile of heldProgramFiles) {
    if (!sourceFile.text.includes('mptyArray')) continue;
    const visit = node => {
        if (ts.isBinaryExpression(node) && comparisonOperators.has(node.operatorToken.kind)) {
            const leftShared = isSharedArray(node.left, finalChecker);
            const rightShared = isSharedArray(node.right, finalChecker);
            if (leftShared !== rightShared) {
                const operand = skipParenthesesAndAssertions(leftShared ? node.right : node.left);
                const where = siteOfNode(node);
                const nameNode = ts.isPropertyAccessExpression(operand) ? operand.name : ts.isIdentifier(operand) ? operand : undefined;
                const symbol = nameNode ? resolvedSymbolOf(nameNode, finalChecker) : undefined;
                const declaration = symbol?.valueDeclaration ?? (symbol?.declarations?.length === 1 ? symbol.declarations[0] : undefined);
                const row = {site: where.site, source: where.source, operand: clip(operand.getText(sourceFile), 60)};
                if (!declaration) {
                    row.slot = nameNode ? 'no declaration for the operand' : `operand is a ${kindName(operand.kind)}, with no symbol`;
                    row.classification = 'Unknown';
                } else {
                    const key = slotIdentity(declaration, 0).key;
                    row.slot = key;
                    if (!sentinelTargets.has(key)) {
                        sentinelTargets.set(key, {key, symbol, declaration, name: symbol.name, references: 0, writes: []});
                    }
                }
                sentinelComparisons.push(row);
            }
        }
        ts.forEachChild(node, visit);
    };
    visit(sourceFile);
}
const targetsByName = new Map();
for (const target of sentinelTargets.values()) {
    const list = targetsByName.get(target.name) ?? [];
    list.push(target);
    targetsByName.set(target.name, list);
}
if (targetsByName.size > 0) {
    for (const sourceFile of heldProgramFiles) {
        const visit = node => {
            if (ts.isIdentifier(node) && targetsByName.has(node.text)) {
                const candidates = targetsByName.get(node.text);
                const symbol = resolvedSymbolOf(node, finalChecker);
                const target = symbol && candidates.find(candidate => symbol === candidate.symbol || (symbol.declarations ?? []).includes(candidate.declaration));
                const isDeclarationName = target && node.parent && node.parent.name === node && node.parent === target.declaration;
                if (target && !isDeclarationName) {
                    target.references++;
                    const reference = ts.isPropertyAccessExpression(node.parent) && node.parent.name === node ? node.parent : node;
                    const how = referenceWrite(reference);
                    if (how) target.writes.push({...siteOfNode(reference), how});
                }
            }
            ts.forEachChild(node, visit);
        };
        visit(sourceFile);
    }
}
for (const row of sentinelComparisons) {
    const target = sentinelTargets.get(row.slot);
    if (!target) continue;
    row.classification = target.writes.length > 0 ? 'Written' : 'AssignedOnly';
    row.references = target.references;
    row.firstWrite = target.writes.length > 0 ? `${target.writes[0].site} ${target.writes[0].how}: ${target.writes[0].source}` : '';
    row.writeCount = target.writes.length;
}
const sentinelClassCounts = new Map();
for (const row of sentinelComparisons) sentinelClassCounts.set(row.classification, (sentinelClassCounts.get(row.classification) ?? 0) + 1);

// 7. Unmatched sites explained, in the final program
function arrayKindOf(type) {
    const name = type.getSymbol()?.name;
    if (name === 'ReadonlyArray') return 'Readonly';
    if (name === 'Array') return 'Mutable';
    if (finalChecker.isTupleType(type)) return type.target?.readonly ? 'Readonly' : 'Mutable';
    return undefined;
}
function contextualReadonly(type) {
    if (!type) return 'no contextual type';
    if (type.isUnion()) {
        const kinds = type.types.map(arrayKindOf).filter(kind => kind !== undefined);
        if (kinds.length === 0) return 'no array members';
        return kinds.every(kind => kind === 'Readonly') ? 'yes' : 'no';
    }
    const kind = arrayKindOf(type);
    if (kind === undefined) return 'not an array';
    return kind === 'Readonly' ? 'yes' : 'no';
}
function currentFromOriginal(absolute, original) {
    let shift = 0;
    for (const insertion of insertionsOf(absolute)) {
        if (insertion.original > original) break;
        shift += insertion.text.length;
    }
    return original + shift;
}
const unmatchedExplained = [];
for (const row of siteRows.filter(each => each.group === 'unmatched')) {
    const match = /^(.*):(\d+):(\d+)$/.exec(row.site);
    const explained = {site: row.site, words: row.words};
    unmatchedExplained.push(explained);
    if (!match) {
        explained.note = 'not file:line:col';
        continue;
    }
    const absolute = path.join(outDirectory, match[1]);
    const sourceFile = program.getSourceFile(absolute);
    const text = originalText.get(absolute);
    if (!sourceFile || text === undefined) {
        explained.note = 'file not in the program';
        continue;
    }
    const starts = lineStartsOf(text);
    const line = Number(match[2]);
    if (line < 1 || line > starts.length) {
        explained.note = 'line past the file';
        continue;
    }
    const position = currentFromOriginal(absolute, starts[line - 1] + Number(match[3]) - 1);
    try {
        const node = deepestNodeAt(sourceFile, position);
        explained.kind = kindName(node.kind);
        const chain = [];
        for (let parent = node.parent; parent && chain.length < 3; parent = parent.parent) {
            chain.push(kindName(parent.kind));
            if (ts.isStatement(parent) || ts.isSourceFile(parent)) break;
        }
        explained.parents = chain.join(' < ');
        if (node.getStart(sourceFile) !== position) {
            explained.note = `no node starts here; the deepest node holding it is a ${explained.kind}`;
        } else if (!ts.isExpression(node)) {
            explained.note = `not an expression: ${explained.kind}`;
        } else {
            explained.type = clip(finalChecker.typeToString(finalChecker.getTypeAtLocation(node)), 120);
            const contextual = finalChecker.getContextualType(node);
            explained.contextual = contextual ? clip(finalChecker.typeToString(contextual), 120) : '<none>';
            explained.contextualReadonly = contextualReadonly(contextual);
        }
    } catch (error) {
        explained.note = `explaining threw: ${clip(String(error && error.message ? error.message : error), 160)}`;
    }
}

// edits.diff: every live insertion, as the untouched line and the final line
const diffLines = [];
for (const absolute of [...insertionsByFile.keys()].sort()) {
    const list = insertionsOf(absolute);
    if (list.length === 0) continue;
    const starts = lineStartsOf(originalText.get(absolute));
    const lineIndexes = [...new Set(list.map(insertion => lineIndexOf(starts, insertion.original)))].sort((left, right) => left - right);
    for (const index of lineIndexes) {
        diffLines.push(`${relativeName(absolute)}:${index + 1}`, `-${lineText(originalText.get(absolute), index)}`, `+${lineText(currentText.get(absolute), index)}`);
    }
}
fs.writeFileSync(path.join(outDirectory, 'edits.diff'), diffLines.join('\n') + '\n');
log(`edits.diff: ${diffLines.length} lines`);

// table.md, or table3.md under --hold-public
function cell(value) {
    return String(value ?? '').replace(/\|/g, '\\|').replace(/[\r\n]+/g, ' ');
}
function table(headers, rows) {
    if (rows.length === 0) return '_none_\n';
    return [`| ${headers.join(' | ')} |`, `| ${headers.map(() => '---').join(' | ')} |`, ...rows.map(row => `| ${row.map(cell).join(' | ')} |`)].join('\n') + '\n';
}
function bySite(left, right) {
    return left.localeCompare(right, 'en', {numeric: true});
}
const allSlots = [...slots.values()].sort((left, right) => bySite(left.key, right.key));
const stateCounts = new Map();
for (const slot of allSlots) stateCounts.set(slot.state, (stateCounts.get(slot.state) ?? 0) + 1);
const readonlySlots = allSlots.filter(slot => slot.state === 'Readonly');
const visibilityCounts = new Map();
for (const slot of readonlySlots) visibilityCounts.set(slot.visibility, (visibilityCounts.get(slot.visibility) ?? 0) + 1);
const finalUnattributed = [...unattributedWrites.values()].filter(write => write.lastRound === finalStep);
const finalUnclassified = [...unclassified.values()].filter(entry => entry.lastRound === finalStep);
const finalConflicts = finalRecords.filter(record => record.kind === 'Readonly').length;
const formatCounts = counts => [...counts].sort((left, right) => String(left[0]).localeCompare(String(right[0]))).map(([name, count]) => `${name} ${count}`).join(', ') || 'none';

// Under --hold-public every slot left Readonly must be Local or Internal, or the change is not the internal one
const publicReadonly = holdPublic ? readonlySlots.filter(slot => slot.visibility !== 'Local' && slot.visibility !== 'Internal') : [];
if (publicReadonly.length > 0) log(`FAIL: ${publicReadonly.length} slots left Readonly are not Local or Internal: ${publicReadonly.map(slot => slot.key).join('; ')}`);
const blockedSlots = allSlots.filter(slot => slot.state === 'Blocked');
// The minimal public set: each Held slot at the root of a Blocked chain or of sentinel residue, with the listed sites
// it holds, and what its own readonly alone would change in its declaration file
const minimalPublicSet = [];
const heldWithSitesOnly = [];
for (const held of allSlots.filter(slot => slot.state === 'Held')) {
    const blockedUnder = blockedSlots.filter(slot => slot.blockedBy === held.key);
    const residueUnder = finalResidue.filter(row => row.root === held.key);
    const sitesUnder = siteRows.filter(row => row.root === held.key).map(row => row.site);
    if (sitesUnder.length === 0) continue;
    if (blockedUnder.length === 0 && residueUnder.length === 0) {
        heldWithSitesOnly.push(`${held.key} (${sitesUnder.join(', ')})`);
        continue;
    }
    let preview;
    if (held.heldInsertions.length > 0) {
        const insertions = [...held.heldInsertions];
        sortInsertions(insertions);
        preview = declarationDiff(held.absolute, renderWith(originalText.get(held.absolute), insertions));
    }
    minimalPublicSet.push({held, blockedUnder, residueUnder, sitesUnder, preview, source: clip(lineText(originalText.get(held.absolute), held.line - 1))});
}
minimalPublicSet.sort((left, right) => right.sitesUnder.length - left.sitesUnder.length || bySite(left.held.key, right.held.key));
const minimalPublicSites = new Set(minimalPublicSet.flatMap(row => row.sitesUnder));
// Listed sites by group, in the order the ruling reads them: Readonly, Blocked, sentinel residue, Written, the rest
function groupRank(group) {
    if (group.startsWith('Readonly')) return 0;
    if (group.startsWith('Blocked')) return 1;
    if (group.startsWith('Sentinel residue')) return 2;
    if (group.startsWith('Written')) return 3;
    return 4;
}
const orderedSiteGroups = [...siteGroups].sort((left, right) => groupRank(left[0]) - groupRank(right[0]) || bySite(left[0], right[0]));

const markdown = [];
markdown.push('# #r11wppy item 1: the readonly fixpoint for emptyArray and resolvingEmptyArray\n');
if (baselineRecords.length > 50) {
    markdown.push(`**The baseline holds ${baselineRecords.length} diagnostics, over 50: the program's setup may not match tsc's own build. The first 30 are listed at the end.**\n`);
}
markdown.push('## Summary\n');
if (holdPublic && !blockSettled) {
    markdown.push(`**Blocking still put slots back after ${maximumBlockRounds} rounds: what is left Readonly is not settled.**\n`);
}
if (publicReadonly.length > 0) {
    markdown.push(`**${publicReadonly.length} slots left Readonly are not Local or Internal; see the internal manifest.**\n`);
}
markdown.push(table(['Measure', 'Value'], [
    ['Baseline diagnostics', baselineRecords.length],
    ['TypeScript', ts.version],
    ['Program', `${rootNames.length} root files, ${program.getSourceFiles().length} source files; src/harness and src/testRunner left out, since they need mocha and chai types the box doesn't have`],
    [holdPublic ? 'Rounds, forward phase' : 'Rounds', `${rounds}${converged ? ', converged' : `, not converged within ${maximumRounds}`}`],
    ...(holdPublic ? [['Rounds, blocking phase', `${blockRounds}${blockSettled ? ', settled' : converged ? `, not settled within ${maximumBlockRounds}` : ', skipped'}`]] : []),
    ['Slots by state', formatCounts(stateCounts)],
    ['Readonly slots by visibility', formatCounts(visibilityCounts)],
    holdPublic
        ? ['Declaration check', `${declarationLineTotal} changed lines over ${declarationDiffs.length} edited files, expected 0${declarationErrors ? `; ${declarationErrors} files threw` : ''}`]
        : ['Public d.ts lines changed', `${declarationLineTotal} over ${declarationDiffs.length} files`],
    ['Sentinel reads', `${sentinels.length} lines by text; ${sentinelComparisons.length} comparisons in the program by slot class: ${formatCounts(sentinelClassCounts)}`],
    [`Listed sites (${siteRows.length}) by group`, formatCounts(siteGroups)],
    ...(holdPublic ? [
        ['Sentinel residue in the final round', finalResidue.length],
        ['Unhandled blocking sources in the final round', finalUnhandled.length],
        ['Minimal public set', `${minimalPublicSet.length} held slots holding ${minimalPublicSites.size} listed sites`],
        ['Slots left Readonly that are not Local or Internal', publicReadonly.length],
    ] : []),
    ['Readonly-into-mutable diagnostics left in the final round', finalConflicts],
    ['Unattributed writes in the final round', `${finalUnattributed.length} (${unattributedWrites.size} ever seen)`],
    ['Unclassified diagnostics in the final round', `${finalUnclassified.length} (${unclassified.size} ever seen)`],
    ['Script errors while classifying or editing', scriptErrors.length],
    ['edits.diff lines', diffLines.length],
]));
markdown.push(table(['Round', 'New diagnostics', 'Readonly into mutable', 'Writes', 'Other', 'New slots', 'Made readonly', 'Written', 'Seconds'],
    roundSummaries.map(summary => [summary.round, summary.fresh, summary.readonlyCount, summary.writeCount, summary.otherCount, summary.newSlots,
        summary.madeReadonly, summary.written, summary.seconds.toFixed(1)])));
if (holdPublic) {
    markdown.push(table(['Blocking round', 'New diagnostics', 'Readonly into mutable', 'Into a held, dropped or blocked slot', 'Blocked', 'Sentinel residue', 'Unhandled', 'Other', 'Seconds'],
        blockRoundSummaries.map(summary => [summary.round, summary.fresh, summary.readonlyCount, summary.rootedCount, summary.blocked, summary.residue,
            summary.unhandled, summary.otherCount, summary.seconds.toFixed(1)])));
    markdown.push('### Listed sites by group\n');
    markdown.push('Readonly: the site\'s slot stays readonly, and every such slot is Local or Internal. Blocked by a root: the slot was put back because its readonly value reached that Held or Dropped slot. Sentinel residue by a root: the shared array itself goes into a slot under that root. Held or Dropped: the site goes straight into that slot.\n');
    markdown.push(table(['Group', 'Listed sites'], orderedSiteGroups.map(([group, count]) => [group, count])));
}
markdown.push('Visibility: Public shows in a declaration file; Internal carries @internal on itself or an ancestor; Local lives in a function body, is private, or sits under a statement the module doesn\'t export, so no declaration file shows it.\n');
markdown.push('Written slots stay refused in the build, by Adamic\'s ruling. WrittenUnion is a written slot typed `readonly never[] | <original>` only so the fixpoint converges and counts; its sites count under Written. Dropped slots are never edited, by the same ruling. AliasRef is a slot typed by an alias under src, which is edited as its own slot of kind Alias.\n');
if (holdPublic) {
    markdown.push('Held is a Public slot the rounds would have made readonly, left as tsc wrote it under --hold-public. Blocked is a slot the rounds made readonly and blocking put back, because its readonly value reached a Held or Dropped slot, directly or through other Blocked slots; it names the slot at the end of that chain as its root.\n');
}

markdown.push('## Slots written\n');
markdown.push(table(['Slot', 'Edit', 'Writes', 'Listed sites', 'Diagnostics'], allSlots.filter(slot => slot.state === 'Written' || slot.state === 'WrittenUnion').map(slot => [
    `${slot.key}${slot.seed ? ' (seed)' : ''}`,
    slot.state === 'WrittenUnion' ? `readonly never[] | ${slot.originalType}` : 'put back',
    slot.writes.map(write => `${write.site} ${write.property}${write.via.length ? ` via ${write.via.join(', ')}` : ''}: ${write.source}`).join(' ; '),
    slot.listedSites ?? 0,
    slot.diagnosticSites.size,
])));

if (holdPublic) {
    markdown.push('## Internal manifest\n');
    markdown.push('Every slot left Readonly: the change that builds alone, with every Public slot as tsc wrote it.\n');
    if (publicReadonly.length > 0) markdown.push(`**Not Local or Internal: ${publicReadonly.map(slot => `${slot.key} (${slot.visibility})`).join(', ')}.**\n`);
    markdown.push(table(['Slot', 'Kind', 'Visibility', 'Type', 'Members left', 'Note', 'Listed sites', 'Which', 'Diagnostics', 'Round'], readonlySlots.map(slot => [
        `${slot.key}${slot.seed ? ' (seed)' : ''}`, slot.kind, slot.visibility, slot.originalType, slot.untouchedMembers.join(', '),
        [...slot.notes].join(' ; '), slot.listedSiteNames.length, slot.listedSiteNames.join(', '), slot.diagnosticSites.size, slot.round,
    ])));
    markdown.push('### Declaration check\n');
    markdown.push(`transpileDeclaration with stripInternal, before and after the final edits, over every file they touch: ${declarationLineTotal} changed lines, expected 0.\n`);
    if (declarationLineTotal !== 0) markdown.push(`**${declarationLineTotal} declaration lines changed where 0 were expected.**\n`);
    markdown.push(table(['File', 'Changed lines', 'First changed lines'], declarationDiffs.map(diff => [
        diff.relative, diff.error ? `transpileDeclaration threw: ${diff.error}` : diff.total, diff.lines.slice(0, 6).map(line => clip(line, 100)).join(' ; '),
    ])));

    markdown.push('## Minimal public set\n');
    markdown.push('Each Held slot at the root of a Blocked chain or of sentinel residue, holding at least one listed site: the slots blocked under it, its residue, the listed sites they hold (a site going straight into the Held slot counts too), and the declaration lines its own readonly alone would change, read by transpileDeclaration on the untouched file with only that insertion.\n');
    markdown.push(table(['Held slot', 'Source', 'Blocked under it', 'Sentinel residue', 'Listed sites', 'Which', 'd.ts lines with its readonly'], minimalPublicSet.map(row => [
        row.held.key, row.source, row.blockedUnder.map(slot => slot.key).join(' ; '), row.residueUnder.map(residue => residue.site).join(' ; '),
        row.sitesUnder.length, row.sitesUnder.join(', '),
        row.preview
            ? (row.preview.error ? `transpileDeclaration threw: ${row.preview.error}` : `${row.preview.total}: ${row.preview.lines.slice(0, 6).map(line => clip(line, 100)).join(' ; ')}`)
            : 'no insertion kept',
    ])));
    if (heldWithSitesOnly.length > 0) {
        markdown.push(`Held slots that take listed sites directly but root no Blocked slot and no residue, so their sources went unhandled: ${heldWithSitesOnly.join('; ')}.\n`);
    }

    markdown.push('## Slots blocked\n');
    markdown.push(table(['Slot', 'Visibility', 'Type', 'Root', 'Through', 'At', 'Blocking round', 'Listed sites'], blockedSlots.map(slot => [
        slot.key, slot.visibility, slot.originalType, slot.blockedBy, slot.blockedVia, slot.blockedAt, slot.blockedRound - rounds, slot.listedSiteNames.join(', '),
    ])));

    markdown.push('## Sentinel residue\n');
    markdown.push('The shared array itself going into a Held, Dropped or Blocked slot in the final round.\n');
    markdown.push(table(['Site', 'Source', 'Root', 'Receiver', 'Diagnostic', 'Via'], finalResidue.sort((left, right) => bySite(left.site, right.site)).map(row => [
        row.site, row.source, row.root, row.receiver, row.diagnosticSite, row.via.join(', '),
    ])));

    markdown.push('## Unhandled blocking sources\n');
    markdown.push('A readonly value going into a Held, Dropped or Blocked slot in the final round whose source is neither a slot this script left readonly nor the shared array, so nothing was put back for it.\n');
    markdown.push(table(['Site', 'Code', 'Receiver', 'Root', 'Why', 'Source'], finalUnhandled.sort((left, right) => bySite(left.site, right.site)).map(row => [
        row.site, row.code, row.receiver, row.root, row.reasons.join(' ; '), row.source,
    ])));
} else {
    markdown.push('## Slots made readonly\n');
    markdown.push(table(['Slot', 'Kind', 'Visibility', 'Type', 'Members left', 'Note', 'Listed sites', 'Diagnostics', 'Round'], readonlySlots.map(slot => [
        `${slot.key}${slot.seed ? ' (seed)' : ''}`, slot.kind, slot.visibility, slot.originalType, slot.untouchedMembers.join(', '),
        [...slot.notes].join(' ; '), slot.listedSites ?? 0, slot.diagnosticSites.size, slot.round,
    ])));
    markdown.push('### Public d.ts lines by file\n');
    markdown.push(table(['File', 'Changed lines', 'First changed lines'], declarationDiffs.map(diff => [
        diff.relative, diff.error ? `transpileDeclaration threw: ${diff.error}` : diff.total, diff.lines.slice(0, 6).map(line => clip(line, 100)).join(' ; '),
    ])));
}

markdown.push('## Named types and unannotated\n');
markdown.push(table(['Slot', 'State', 'Name', 'Note', 'Listed sites', 'Diagnostics'],
    allSlots.filter(slot => ['NamedType', 'Unannotated', 'AlreadyReadonly', 'Outside', 'Unresolved', 'Dropped', 'Held', 'AliasRef'].includes(slot.state)).map(slot => [
        slot.key, slot.state, slot.name, [...slot.notes].join(' ; '), slot.listedSites ?? 0,
        [...slot.diagnosticSites].slice(0, 4).join(' ; ') + (slot.diagnosticSites.size > 4 ? ` ; and ${slot.diagnosticSites.size - 4} more` : ''),
    ])));

markdown.push('## Unattributed writes\n');
markdown.push(`Present in the final round. ${unattributedWrites.size - finalUnattributed.length} more were seen in earlier rounds only.\n`);
markdown.push(table(['Site', 'Code', 'Write', 'Receiver', 'Receiver declaration', 'Source'], finalUnattributed.sort((left, right) => bySite(left.site, right.site)).map(write => [
    write.site, write.code, write.property, `${write.receiver}${write.via.length ? ` via ${write.via.join(', ')}` : ''}`, write.receiverDeclaration, write.source,
])));
markdown.push('## Unclassified diagnostics\n');
markdown.push(`Present in the final round. ${unclassified.size - finalUnclassified.length} more were seen in earlier rounds only.\n`);
markdown.push(table(['Code', 'Site', 'Message'], finalUnclassified.sort((left, right) => bySite(left.site, right.site)).map(entry => [entry.code, entry.site, entry.message])));
if (scriptErrors.length > 0) {
    markdown.push('### Script errors\n');
    markdown.push(table(['Error'], scriptErrors.slice(0, 40).map(error => [clip(error, 400)])));
}

markdown.push('## Sentinels\n');
markdown.push(table(['Site', 'Source'], sentinels.map(sentinel => [sentinel.site, sentinel.source])));

markdown.push('## Sentinels by slot\n');
markdown.push('Each comparison with the shared array in the program, the other operand\'s declaration, and every reference to it found by symbol across the program. Written means some reference calls a mutating method, assigns an element, or assigns length; AssignedOnly means every reference reads it or assigns a whole value. A write through a local alias of the slot is not seen.\n');
markdown.push(table(['Site', 'Operand', 'Slot', 'Class', 'References', 'Writes', 'First write', 'Source'], sentinelComparisons.map(row => [
    row.site, row.operand, row.slot, row.classification, row.references ?? '', row.writeCount ?? '', row.firstWrite ?? '', row.source,
])));

markdown.push('## Unmatched sites\n');
markdown.push(table(['Site', 'Why', 'Words'], siteRows.filter(row => row.group === 'unmatched').map(row => [row.site, row.note, clip(row.words)])));

markdown.push('## Unmatched sites explained\n');
markdown.push('The deepest node starting at each unmatched site, read in the final program: its own type, its contextual type, whether that contextual type is readonly (a readonly array, or a union whose array members all are), and up to three parent kinds toward the statement.\n');
markdown.push(table(['Site', 'Node', 'Type', 'Contextual type', 'Contextual readonly', 'Parents', 'Note'], unmatchedExplained.map(row => [
    row.site, row.kind ?? '', row.type ?? '', row.contextual ?? '', row.contextualReadonly ?? '', row.parents ?? '', row.note ?? '',
])));

markdown.push('## Listed sites by slot\n');
markdown.push(table(['Site', 'Group', 'Slot', 'Matched by'], siteRows.filter(row => row.group !== 'unmatched').map(row => [row.site, row.group, row.slotKey, row.byLine ? 'line' : 'span'])));

if (baselineRecords.length > 0) {
    markdown.push('## Baseline diagnostics (first 30)\n');
    markdown.push(table(['Code', 'Site', 'Message'], baselineRecords.slice(0, 30).map(record => [record.code, record.site, clip(record.message, 200)])));
}

fs.writeFileSync(path.join(outDirectory, tableName), markdown.join('\n'));
fs.writeFileSync(path.join(outDirectory, 'result.json'), JSON.stringify({
    typeScript: ts.version, baseline: baselineRecords.length, rounds, converged, roundSummaries,
    slots: allSlots.map(slot => ({...slot, diagnosticSites: [...slot.diagnosticSites], notes: [...slot.notes], insertions: slot.insertions.map(insertion => ({original: insertion.original, text: insertion.text})),
        heldInsertions: slot.heldInsertions.map(insertion => ({original: insertion.original, text: insertion.text})), pendingInsertions: undefined, absolute: undefined})),
    declarationDiffs, sentinels, sentinelComparisons, unmatchedExplained, siteRows, unattributedWrites: [...unattributedWrites.values()], unclassified: [...unclassified.values()], scriptErrors,
    ...(holdPublic ? {blockRounds, blockSettled, blockRoundSummaries, sentinelResidue: finalResidue, unhandledSources: finalUnhandled, publicReadonly: publicReadonly.map(slot => slot.key),
        minimalPublicSet: minimalPublicSet.map(row => ({held: row.held.key, blocked: row.blockedUnder.map(slot => slot.key), residue: row.residueUnder.map(residue => residue.site), sites: row.sitesUnder, declaration: row.preview}))} : {}),
}, null, 2));
log(`${tableName} written: ${allSlots.length} slots, ${readonlySlots.length} readonly, ${(stateCounts.get('Written') ?? 0) + (stateCounts.get('WrittenUnion') ?? 0)} written, `
    + `${siteRows.length - (siteGroups.get('unmatched') ?? 0)} of ${siteRows.length} listed sites matched`);

if (scriptErrors.length > 0) process.exit(5);
if (!converged) process.exit(4);
if (holdPublic && !blockSettled) process.exit(6);
if (publicReadonly.length > 0) process.exit(7);
