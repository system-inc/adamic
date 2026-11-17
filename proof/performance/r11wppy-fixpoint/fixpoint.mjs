// #r11wppy item 1: the readonly fixpoint for tsc's shared emptyArray and resolvingEmptyArray. Both are typed
// `readonly never[]`, then every slot that receives one is made readonly too, round after round, until a round
// makes no edit. A slot tsc really writes through is put back and marked written. Never merged: a proof only.
//
// Usage: node fixpoint.mjs <adapted tree> <never-sites.tsv> <out directory>
//
// Every edit is an insertion of text with no line break ("readonly ", "Readonly", "(readonly " and ")"), so line
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

const [adaptedArgument, sitesArgument, outArgument] = process.argv.slice(2);
if (!adaptedArgument || !sitesArgument || !outArgument) {
    throw new Error('usage: node fixpoint.mjs <adapted tree> <never-sites.tsv> <out directory>');
}
const adaptedDirectory = path.resolve(adaptedArgument);
const sitesPath = path.resolve(sitesArgument);
fs.mkdirSync(path.resolve(outArgument), {recursive: true});
// The real path, so the held texts' names match the names module resolution builds
const outDirectory = fs.realpathSync(path.resolve(outArgument));
const outSource = path.join(outDirectory, 'src');
const maximumRounds = 80;
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
function render(absolute) {
    const text = originalText.get(absolute);
    let result = '';
    let cursor = 0;
    for (const insertion of insertionsOf(absolute)) {
        result += text.slice(cursor, insertion.original) + insertion.text;
        cursor = insertion.original;
    }
    return result + text.slice(cursor);
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
function planReadonly(typeNode, sourceFile, plan) {
    if (ts.isParenthesizedTypeNode(typeNode)) {
        planReadonly(typeNode.type, sourceFile, plan);
        return;
    }
    if (ts.isArrayTypeNode(typeNode)) {
        const start = typeNode.getStart(sourceFile);
        // An element type of an array type needs parentheses: (readonly T[])[]
        if (ts.isArrayTypeNode(typeNode.parent) && typeNode.parent.elementType === typeNode) {
            plan.insertions.push({position: start, text: '(readonly '}, {position: typeNode.end, text: ')'});
        } else {
            plan.insertions.push({position: start, text: 'readonly '});
        }
        plan.changed.push(typeNode.getText(sourceFile));
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
            plan.named.push(name);
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

// The receiving slot of a readonly value going into a mutable one, from the node the diagnostic sits on, walking up
function resolveReceivingSlot(errorNode, checker) {
    let node = errorNode;
    let elementDepth = 0;
    for (let step = 0; step < 64; step++) {
        if (ts.isReturnStatement(node)) {
            const fn = containingFunction(node);
            return fn ? {declaration: fn, elementDepth, via: []} : {unresolved: 'return outside a function'};
        }
        if ((ts.isVariableDeclaration(node) || ts.isPropertyDeclaration(node) || ts.isParameter(node)) && step === 0) {
            return {declaration: node, elementDepth, via: []};
        }
        const parent = node.parent;
        if (!parent) return {unresolved: `${kindName(node.kind)} with no parent`};
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
function slotFromResolution(resolution, describedDiagnostic) {
    if (resolution.unresolved) {
        const key = `${describedDiagnostic.relative}:${describedDiagnostic.line} unresolved: ${resolution.unresolved}`;
        let slot = slots.get(key);
        if (!slot) {
            slot = newSlot({key, relative: describedDiagnostic.relative, line: describedDiagnostic.line, label: `unresolved: ${resolution.unresolved}`});
            slot.state = 'Unresolved';
            slot.name = resolution.unresolved;
            slots.set(key, slot);
        }
        return {slot, created: false};
    }
    const identity = slotIdentity(resolution.declaration, resolution.elementDepth);
    const existing = slots.get(identity.key);
    if (existing) return {slot: existing, created: false};
    const slot = newSlot(identity);
    slots.set(identity.key, slot);
    if (!currentText.has(identity.absolute)) {
        slot.state = 'Outside';
        slot.name = 'declared outside the copy of src';
        return {slot, created: true};
    }
    slot.visibility = visibilityOf(resolution.declaration);
    const target = slotTypeNode(resolution.declaration, resolution.elementDepth);
    if (target.missing) {
        slot.state = 'Unannotated';
        slot.name = kindName(resolution.declaration.kind);
        return {slot, created: true};
    }
    if (target.named) {
        slot.state = 'NamedType';
        slot.name = target.named;
        return {slot, created: true};
    }
    const sourceFile = resolution.declaration.getSourceFile();
    const plan = {insertions: [], changed: [], named: [], other: [], already: false};
    planReadonly(target.typeNode, sourceFile, plan);
    slot.originalType = clip(target.typeNode.getText(sourceFile), 100);
    if (plan.insertions.length > 0) {
        slot.state = 'Readonly';
        slot.changed = plan.changed.map(text => clip(text, 60));
        slot.untouchedMembers = [...plan.named, ...plan.other.map(kind => `<${kind}>`)];
        slot.pendingInsertions = plan.insertions.map(insertion => {
            if (/[\r\n\u2028\u2029]/.test(insertion.text)) throw new Error('an insertion may not hold a line break');
            return {absolute: identity.absolute, original: toOriginal(identity.absolute, insertion.position), text: insertion.text, rank: resolution.elementDepth, slotKey: slot.key};
        });
    } else if (plan.already) {
        slot.state = 'AlreadyReadonly';
        slot.name = slot.originalType;
    } else {
        slot.state = 'NamedType';
        slot.name = [...plan.named, ...plan.other.map(kind => `<${kind}>`)].join(', ') || '<nothing to edit>';
    }
    return {slot, created: true};
}

function newSlot(identity) {
    return {
        key: identity.key, relative: identity.relative, line: identity.line, label: identity.label, absolute: identity.absolute,
        state: undefined, visibility: undefined, name: undefined, originalType: undefined, changed: [], untouchedMembers: [],
        pendingInsertions: [], insertions: [], writes: [], diagnosticSites: new Set(), round: undefined, seed: false,
    };
}

// Returns false, leaving the slot unedited, when another live insertion holds the same offset and rank, since
// then their order in the text would be a guess
function applyInsertions(slot) {
    for (const insertion of slot.pendingInsertions) {
        const clash = insertionsOf(insertion.absolute).find(other => other.original === insertion.original && other.rank === insertion.rank);
        if (clash) {
            scriptErrors.push(`${slot.key}: its insertion at ${insertion.original} clashes with ${clash.slotKey}'s; left unedited`);
            slot.state = 'Unresolved';
            slot.name = `insertion clash with ${clash.slotKey}`;
            slot.pendingInsertions = [];
            return false;
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
    const {slot} = slotFromResolution({declaration, elementDepth: 0, via: []}, undefined);
    if (slot.state !== 'Readonly') {
        throw new Error(`${name} in src/${relativeFile} has type ${slot.originalType ?? '<none>'}, state ${slot.state}: expected a mutable array to make readonly`);
    }
    slot.seed = true;
    slot.round = 0;
    seedSlots.push(slot);
}
const touchedBySeed = new Set();
for (const slot of seedSlots) {
    if (!applyInsertions(slot)) throw new Error(`the seed ${slot.key} couldn't be applied: ${slot.name}`);
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
    }
    const forward = created.filter(slot => slot.state === 'Readonly' && slot.pendingInsertions.length > 0)
        .filter(slot => applyInsertions(slot));
    for (const slot of forward) touched.add(slot.absolute);
    for (const absolute of touched) currentText.set(absolute, render(absolute));
    const summary = {round, fresh: fresh.length, readonlyCount, writeCount, otherCount, madeReadonly: forward.length, written: reverting.size,
        newSlots: created.length, seconds: (Date.now() - roundBegan) / 1000};
    roundSummaries.push(summary);
    log(`round ${round}: ${fresh.length} new diagnostics (${readonlyCount} readonly into mutable, ${writeCount} writes, ${otherCount} other), `
        + `${created.length} new slots, ${forward.length} made readonly, ${reverting.size} written and put back, ${summary.seconds.toFixed(1)} s`);
    if (forward.length === 0 && reverting.size === 0) {
        converged = true;
        break;
    }
}

function classify(record, round, checker, created, reverting) {
    const isReadonlyClass = record.code === 4104 || record.message.includes(readonlyPhrase);
    const sourceFile = record.absolute ? program.getSourceFile(record.absolute) : undefined;
    if (isReadonlyClass && sourceFile && record.held) {
        record.kind = 'Readonly';
        const errorNode = errorNodeAt(sourceFile, record.diagnostic.start, record.diagnostic.length ?? 0);
        const resolution = resolveReceivingSlot(errorNode, checker);
        const result = slotFromResolution(resolution, record);
        if (result.created) {
            result.slot.round = round;
            created.push(result.slot);
        }
        result.slot.diagnosticSites.add(record.site);
        record.slotKey = result.slot.key;
        const diagnosticKey = `${record.relative}|${record.originalStart}|${record.code}`;
        if (!readonlyDiagnostics.has(diagnosticKey)) {
            readonlyDiagnostics.set(diagnosticKey, {relative: record.relative, absolute: record.absolute, originalStart: record.originalStart,
                originalEnd: record.originalEnd, line: record.line, site: record.site, slotKey: result.slot.key, round});
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
            const live = slot && (slot.state === 'Readonly' || (slot.state === 'Written' && reverting.has(slot.key)));
            if (live) {
                if (slot.state === 'Readonly') {
                    checkRevert(slot, target.declaration);
                    slot.state = 'Written';
                    slot.writtenRound = round;
                    reverting.set(slot.key, slot);
                }
                slot.writes.push({site: record.site, code: record.code, property: write.property, via: target.via, source});
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

// The final texts go to disk, so the copy under out/src is the fixpoint
for (const [absolute, text] of currentText) {
    if (text !== originalText.get(absolute)) fs.writeFileSync(absolute, text);
}

// 7. The declaration-file lines for public slots left readonly
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
const publicFiles = [...new Set([...slots.values()].filter(slot => slot.state === 'Readonly' && slot.visibility === 'Public').map(slot => slot.absolute))].sort();
const declarationDiffs = [];
for (const absolute of publicFiles) {
    const before = originalText.get(absolute);
    const after = currentText.get(absolute);
    try {
        let diff;
        if (absolute.endsWith('.d.ts')) {
            diff = changedLineCount(before, after);
        } else {
            const fileName = absolute;
            const declarationBefore = ts.transpileDeclaration(before, {compilerOptions: declarationOptions, fileName}).outputText;
            const declarationAfter = ts.transpileDeclaration(after, {compilerOptions: declarationOptions, fileName}).outputText;
            diff = changedLineCount(declarationBefore, declarationAfter);
        }
        declarationDiffs.push({relative: relativeName(absolute), total: diff.total, lines: diff.lines});
    } catch (error) {
        declarationDiffs.push({relative: relativeName(absolute), total: 0, lines: [], error: clip(String(error && error.message ? error.message : error), 200)});
    }
}
const declarationLineTotal = declarationDiffs.reduce((sum, diff) => sum + diff.total, 0);

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
    if (!covering) {
        siteRows.push({site, words, group: 'unmatched', note: 'no readonly-into-mutable diagnostic here'});
        continue;
    }
    const slot = slots.get(covering.slotKey);
    siteRows.push({site, words, group: slot?.state ?? 'unmatched', slotKey: covering.slotKey, byLine: !(covering.originalStart <= offset && offset < covering.originalEnd)});
    if (slot) slot.listedSites = (slot.listedSites ?? 0) + 1;
}
const siteGroups = new Map();
for (const row of siteRows) siteGroups.set(row.group, (siteGroups.get(row.group) ?? 0) + 1);

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

// table.md
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
const finalUnattributed = [...unattributedWrites.values()].filter(write => write.lastRound === rounds);
const finalUnclassified = [...unclassified.values()].filter(entry => entry.lastRound === rounds);
const finalConflicts = finalRecords.filter(record => record.kind === 'Readonly').length;
const formatCounts = counts => [...counts].sort((left, right) => String(left[0]).localeCompare(String(right[0]))).map(([name, count]) => `${name} ${count}`).join(', ') || 'none';

const markdown = [];
markdown.push('# #r11wppy item 1: the readonly fixpoint for emptyArray and resolvingEmptyArray\n');
if (baselineRecords.length > 50) {
    markdown.push(`**The baseline holds ${baselineRecords.length} diagnostics, over 50: the program's setup may not match tsc's own build. The first 30 are listed at the end.**\n`);
}
markdown.push('## Summary\n');
markdown.push(table(['Measure', 'Value'], [
    ['Baseline diagnostics', baselineRecords.length],
    ['TypeScript', ts.version],
    ['Program', `${rootNames.length} root files, ${program.getSourceFiles().length} source files; src/harness and src/testRunner left out, since they need mocha and chai types the box doesn't have`],
    ['Rounds', `${rounds}${converged ? ', converged' : `, not converged within ${maximumRounds}`}`],
    ['Slots by state', formatCounts(stateCounts)],
    ['Readonly slots by visibility', formatCounts(visibilityCounts)],
    ['Public d.ts lines changed', `${declarationLineTotal} over ${declarationDiffs.length} files`],
    ['Sentinel reads', sentinels.length],
    [`Listed sites (${siteRows.length}) by group`, formatCounts(siteGroups)],
    ['Readonly-into-mutable diagnostics left in the final round', finalConflicts],
    ['Unattributed writes in the final round', `${finalUnattributed.length} (${unattributedWrites.size} ever seen)`],
    ['Unclassified diagnostics in the final round', `${finalUnclassified.length} (${unclassified.size} ever seen)`],
    ['Script errors while classifying', scriptErrors.length],
    ['edits.diff lines', diffLines.length],
]));
markdown.push(table(['Round', 'New diagnostics', 'Readonly into mutable', 'Writes', 'Other', 'New slots', 'Made readonly', 'Written', 'Seconds'],
    roundSummaries.map(summary => [summary.round, summary.fresh, summary.readonlyCount, summary.writeCount, summary.otherCount, summary.newSlots,
        summary.madeReadonly, summary.written, summary.seconds.toFixed(1)])));
markdown.push('Visibility: Public shows in a declaration file; Internal carries @internal on itself or an ancestor; Local lives in a function body, is private, or sits under a statement the module doesn\'t export, so no declaration file shows it.\n');

markdown.push('## Slots written\n');
markdown.push(table(['Slot', 'Writes', 'Listed sites', 'Diagnostics'], allSlots.filter(slot => slot.state === 'Written').map(slot => [
    `${slot.key}${slot.seed ? ' (seed)' : ''}`,
    slot.writes.map(write => `${write.site} ${write.property}${write.via.length ? ` via ${write.via.join(', ')}` : ''}: ${write.source}`).join(' ; '),
    slot.listedSites ?? 0,
    slot.diagnosticSites.size,
])));

markdown.push('## Slots made readonly\n');
markdown.push(table(['Slot', 'Visibility', 'Type', 'Members left', 'Listed sites', 'Diagnostics', 'Round'], readonlySlots.map(slot => [
    `${slot.key}${slot.seed ? ' (seed)' : ''}`, slot.visibility, slot.originalType, slot.untouchedMembers.join(', '),
    slot.listedSites ?? 0, slot.diagnosticSites.size, slot.round,
])));
markdown.push('### Public d.ts lines by file\n');
markdown.push(table(['File', 'Changed lines', 'First changed lines'], declarationDiffs.map(diff => [
    diff.relative, diff.error ? `transpileDeclaration threw: ${diff.error}` : diff.total, diff.lines.slice(0, 6).map(line => clip(line, 100)).join(' ; '),
])));

markdown.push('## Named types and unannotated\n');
markdown.push(table(['Slot', 'State', 'Name', 'Listed sites', 'Diagnostics'],
    allSlots.filter(slot => ['NamedType', 'Unannotated', 'AlreadyReadonly', 'Outside', 'Unresolved'].includes(slot.state)).map(slot => [
        slot.key, slot.state, slot.name, slot.listedSites ?? 0, [...slot.diagnosticSites].slice(0, 4).join(' ; ') + (slot.diagnosticSites.size > 4 ? ` ; and ${slot.diagnosticSites.size - 4} more` : ''),
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

markdown.push('## Unmatched sites\n');
markdown.push(table(['Site', 'Why', 'Words'], siteRows.filter(row => row.group === 'unmatched').map(row => [row.site, row.note, clip(row.words)])));

markdown.push('## Listed sites by slot\n');
markdown.push(table(['Site', 'Group', 'Slot', 'Matched by'], siteRows.filter(row => row.group !== 'unmatched').map(row => [row.site, row.group, row.slotKey, row.byLine ? 'line' : 'span'])));

if (baselineRecords.length > 0) {
    markdown.push('## Baseline diagnostics (first 30)\n');
    markdown.push(table(['Code', 'Site', 'Message'], baselineRecords.slice(0, 30).map(record => [record.code, record.site, clip(record.message, 200)])));
}

fs.writeFileSync(path.join(outDirectory, 'table.md'), markdown.join('\n'));
fs.writeFileSync(path.join(outDirectory, 'result.json'), JSON.stringify({
    typeScript: ts.version, baseline: baselineRecords.length, rounds, converged, roundSummaries,
    slots: allSlots.map(slot => ({...slot, diagnosticSites: [...slot.diagnosticSites], insertions: slot.insertions.map(insertion => ({original: insertion.original, text: insertion.text})), pendingInsertions: undefined, absolute: undefined})),
    declarationDiffs, sentinels, siteRows, unattributedWrites: [...unattributedWrites.values()], unclassified: [...unclassified.values()], scriptErrors,
}, null, 2));
log(`table.md written: ${allSlots.length} slots, ${readonlySlots.length} readonly, ${stateCounts.get('Written') ?? 0} written, `
    + `${siteRows.length - (siteGroups.get('unmatched') ?? 0)} of ${siteRows.length} listed sites matched`);

if (scriptErrors.length > 0) process.exit(5);
if (!converged) process.exit(4);
