#!/usr/bin/env node
'use strict';
// #rv75m9k road B: each method read as a value whose member is public (or written after its object is built) is read
// through a view instead, or, where a view can't serve, through an exact-parameter arrow. Generated, never hand-edited.
//
// View (the default): before the statement holding the read, `const <name>View: { readonly m: (typeof R)["m"] } = R;`,
// and the read becomes `<name>View.m`. The value read is the same function object tsc reads today, unbound as tsc reads
// it, so it's value-identical and adds no call. One view serves every read of one receiver in one statement. The
// save-then-override sites take a view and never an arrow: the member is overwritten on the next line, so an arrow that
// reads it through its receiver at call time would call the override and recurse.
//
// Arrow: where the object behind the receiver holds the member as a literal method (sys.ts's System object), a view
// stops at the read (erasedMethodField), so the read becomes `(p: T) => R.m(p)` with the member's own parameters. A
// private static method (IdentifierNameMap.toKey) becomes a private static readonly arrow field beside it, made once.
//
// It refuses, before writing anything, a receiver that isn't an identifier or a chain of property reads on one, a read
// whose statement can't take a statement before it, a statement that writes the receiver's root,  and a site it can't find. Usage: NODE_PATH=... node codemod-b.cjs <tree> <roadB-sites.json> [--dry-run]
const ts = require('typescript'), fs = require('node:fs'), path = require('node:path');
const [tree, sitesPath, mode] = process.argv.slice(2);
// ROAD_B_ONLY / ROAD_B_EXCEPT (a file path) narrow the run to one file or to every other, to bisect a census change.
const sites = JSON.parse(fs.readFileSync(sitesPath, 'utf8')).filter(site =>
    (!process.env.ROAD_B_ONLY || site.site.startsWith(process.env.ROAD_B_ONLY + ':')) && (!process.env.ROAD_B_EXCEPT || !site.site.startsWith(process.env.ROAD_B_EXCEPT + ':')));
const arrowMembers = new Set(['System.directoryExists', 'System.createDirectory']);
const staticFieldMembers = new Set(['IdentifierNameMap.toKey']);
// Helpers run once per call on a hot path, where an arrow before the statement would be made every call.
const perCallHelpers = new Set(['toPath', 'relName']);

const config = ts.getParsedCommandLineOfConfigFile(path.join(tree, 'src/compiler/tsconfig.json'), {}, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: () => {} });
config.options.typeRoots = [path.join(process.env.NODE_PATH.split(':')[0], '@types')];
const program = ts.createProgram(config.fileNames, { ...config.options, noEmit: true });
const checker = program.getTypeChecker();
const refusals = [], edits = new Map(); // file -> [{ start, end, text }]
const addEdit = (file, edit) => { if (!edits.has(file)) edits.set(file, []); edits.get(file).push(edit); };

function nodeAt(sourceFile, line, column) {
    const position = sourceFile.getPositionOfLineAndCharacter(line - 1, column - 1);
    let found;
    (function visit(node) {
        if (node.getStart(sourceFile) === position && ts.isPropertyAccessExpression(node)) found = found || node;
        if (node.pos <= position && position < node.end) ts.forEachChild(node, visit);
    })(sourceFile);
    return found;
}
function isStableReceiver(node) {
    return ts.isIdentifier(node) || (ts.isPropertyAccessExpression(node) && ts.isIdentifier(node.name) && isStableReceiver(node.expression));
}
function rootOf(node) { return ts.isIdentifier(node) ? node : rootOf(node.expression); }
function statementOf(node) {
    for (let current = node; current.parent; current = current.parent)
        if (ts.isStatement(current) && (ts.isBlock(current.parent) || ts.isSourceFile(current.parent) || ts.isModuleBlock(current.parent) || ts.isCaseClause(current.parent) || ts.isDefaultClause(current.parent))) return current;
    return undefined;
}
function writesRoot(statement, root) {
    let writes = false;
    (function visit(node) {
        if (ts.isBinaryExpression(node) && node.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && node.operatorToken.kind <= ts.SyntaxKind.LastAssignment && ts.isIdentifier(node.left) && node.left.text === root.text) writes = true;
        ts.forEachChild(node, visit);
    })(statement);
    return writes;
}
function memberSignature(symbol) {
    const declaration = (symbol.declarations || []).find(ts.isMethodSignature) || (symbol.declarations || []).find(ts.isMethodDeclaration);
    return declaration;
}
function parametersText(declaration) {
    const source = declaration.getSourceFile();
    return { list: declaration.parameters.map(parameter => parameter.getText(source)).join(', '), names: declaration.parameters.map(parameter => (parameter.dotDotDotToken ? '...' : '') + parameter.name.getText(source)).join(', ') };
}
function nameInScope(location, name) {
    return checker.getSymbolsInScope(location, ts.SymbolFlags.Value).some(symbol => symbol.name === name);
}

// Every symbol the program writes: an assignment's target, compound or not, ++ and --, through an identifier or a
// property read. A view reads its receiver before its statement's earlier operands run, so it's value-identical only
// when nothing can reassign the receiver in between; a receiver whose root and every link of its chain are never
// written anywhere in the program can't be. (Writes by element access or Object.assign aren't seen; the census and a
// lane count are the second line for those.)
const written = new Set();
for (const sourceFile of program.getSourceFiles()) {
    if (sourceFile.isDeclarationFile) continue;
    (function visit(node) {
        let target;
        if (ts.isBinaryExpression(node) && node.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && node.operatorToken.kind <= ts.SyntaxKind.LastAssignment) target = node.left;
        if ((ts.isPrefixUnaryExpression(node) || ts.isPostfixUnaryExpression(node)) && (node.operator === ts.SyntaxKind.PlusPlusToken || node.operator === ts.SyntaxKind.MinusMinusToken)) target = node.operand;
        if (target && (ts.isIdentifier(target) || ts.isPropertyAccessExpression(target))) {
            const symbol = checker.getSymbolAtLocation(ts.isPropertyAccessExpression(target) ? target.name : target);
            if (symbol) written.add(symbol);
        }
        ts.forEachChild(node, visit);
    })(sourceFile);
}
function receiverEvidence(receiver) {
    const links = [];
    for (let current = receiver; ; current = current.expression) {
        const symbol = checker.getSymbolAtLocation(ts.isPropertyAccessExpression(current) ? current.name : current);
        const name = ts.isPropertyAccessExpression(current) ? current.name.text : current.text;
        if (!symbol) return { ok: false, why: `${name} has no symbol` };
        const declaration = symbol.valueDeclaration;
        const constant = declaration && ts.isVariableDeclaration(declaration) && (ts.getCombinedNodeFlags(declaration) & ts.NodeFlags.Const);
        if (written.has(symbol)) return { ok: false, why: `${name} is written in the program` };
        links.push(`${name} ${constant ? 'const' : ts.isParameter(declaration || receiver) ? 'parameter, never written' : 'never written'}`);
        if (!ts.isPropertyAccessExpression(current)) break;
    }
    return { ok: true, why: links.reverse().join(', ') };
}
const views = new Map(); // statement -> { file, receiverText, viewName, members: Set, statement }
const plannedNames = new Map(); // block-scope container -> names planned in it
function blockOf(statement) {
    for (let current = statement.parent; current; current = current.parent)
        if (ts.isBlock(current) || ts.isCaseBlock(current) || ts.isModuleBlock(current) || ts.isSourceFile(current)) return current;
    return statement.getSourceFile();
}
for (const site of sites) {
    const [file, line, column] = site.site.split(':');
    const sourceFile = program.getSourceFile(path.join(tree, file));
    const node = sourceFile && nodeAt(sourceFile, Number(line), Number(column));
    if (!node) { refusals.push(`${site.site}: no property read found there`); continue; }
    const member = site.member, name = node.name.text;
    if (staticFieldMembers.has(member)) {
        // An inline arrow whose parameters take the types the callee's own parameter declares, so the value is seen as
        // exactly what it's handed to (a wider parameter is "a function taking A seen as one taking B"). A static field
        // would be read before its initializer runs. It's made only on the generated-name path that reads it.
        const declaration = memberSignature(checker.getSymbolAtLocation(node.name));
        const contextual = checker.getContextualType(node);
        const signature = contextual && checker.getSignaturesOfType(contextual, ts.SignatureKind.Call)[0];
        if (!signature) { refusals.push(`${site.site}: no contextual call signature for ${member}`); continue; }
        const list = signature.getParameters().map(parameter => `${parameter.name}: ${checker.typeToString(checker.getTypeOfSymbolAtLocation(parameter, node), node, ts.TypeFormatFlags.NoTruncation)}`).join(', ');
        const names = signature.getParameters().map(parameter => parameter.name).join(', ');
        addEdit(file, { start: node.getStart(sourceFile), end: node.end, text: `((${list}) => ${node.expression.getText(sourceFile)}.${name}(${names}))` });
        console.log(`arrow ${site.site} ${member}: inline, typed as its callee's parameter (${list})`);
        continue;
    }
    if (!isStableReceiver(node.expression)) { refusals.push(`${site.site}: receiver ${node.expression.getText(sourceFile)} isn't an identifier or a chain of reads on one`); continue; }
    const receiverText = node.expression.getText(sourceFile);
    const intersection = Boolean(checker.getTypeAtLocation(node.expression).flags & ts.TypeFlags.Intersection);
    if (arrowMembers.has(member) || intersection) {
        // An exact-parameter arrow through the receiver. For a member the object holds as a literal method a view would
        // stop at the read (erasedMethodField); for an intersection-typed receiver lowering can't hold the value a view
        // needs. The arrow is one const, made once: at the top of the function that declares the receiver's root for
        // a parameter root, otherwise just before the site's statement. A save-then-override site is left: an arrow
        // reads the member at call time and would call the override. So is a per-call helper, where the arrow would be
        // made on every call.
        const enclosing = ts.findAncestor(node, ts.isFunctionLike);
        const enclosingName = enclosing && enclosing.name ? enclosing.name.getText(sourceFile) : '';
        if (site.use === 'variable initializer' || perCallHelpers.has(enclosingName)) { console.log(`left ${site.site} ${member}: ${site.use === 'variable initializer' ? 'save-then-override, where an arrow would call the override' : `in ${enclosingName}, a per-call helper`}; it keeps its refusal`); continue; }
        const declaration = memberSignature(checker.getSymbolAtLocation(node.name));
        const parameters = parametersText(declaration);
        const constName = receiverText.split('.').map((part, index) => index ? part[0].toUpperCase() + part.slice(1) : part).join('') + name[0].toUpperCase() + name.slice(1);
        const rootDeclaration = checker.getSymbolAtLocation(rootOf(node.expression)).valueDeclaration;
        const body = rootDeclaration && ts.isParameter(rootDeclaration) && rootDeclaration.parent.body && ts.isBlock(rootDeclaration.parent.body) ? rootDeclaration.parent.body : undefined;
        const anchor = body ? body.statements[0] : statementOf(node);
        if (!anchor) { refusals.push(`${site.site}: nowhere to make the arrow`); continue; }
        const key = `${file}|arrow|${anchor.getStart(sourceFile)}|${constName}`;
        if (!views.has(key)) {
            views.set(key, true);
            if (nameInScope(anchor, constName)) { refusals.push(`${site.site}: ${constName} is already in scope`); continue; }
            const at = anchor.getStart(sourceFile, true);
            const indentation = ' '.repeat(sourceFile.getLineAndCharacterOfPosition(at).character);
            const returnType = declaration.type ? ': ' + declaration.type.getText(declaration.getSourceFile()) : '';
            addEdit(file, { start: at, end: at, text: `const ${constName} = (${parameters.list})${returnType} => ${receiverText}.${name}(${parameters.names});\n${indentation}` });
        }
        addEdit(file, { start: node.getStart(sourceFile), end: node.end, text: constName });
        console.log(`arrow ${site.site} ${member}: ${constName}, made once ${body ? 'at the top of the function declaring ' + rootOf(node.expression).text : 'before its statement'}`);
        continue;
    }
    const statement = statementOf(node);
    if (!statement) { refusals.push(`${site.site}: no statement it can be placed before`); continue; }
    if (writesRoot(statement, rootOf(node.expression))) { refusals.push(`${site.site}: its statement writes ${rootOf(node.expression).text}`); continue; }
    const evidence = receiverEvidence(node.expression);
    if (!evidence.ok) { refusals.push(`${site.site}: the view would read ${receiverText} early, and ${evidence.why}`); continue; }
    console.log(`evidence ${site.site} ${receiverText}: ${evidence.why}`);
    const key = `${file}|${statement.getStart(sourceFile)}|${receiverText}`;
    if (!views.has(key)) {
        // A view's name is free in its statement's scope and unique in its block: case clauses share one block scope.
        const container = blockOf(statement), base = receiverText.split('.').pop() + 'View';
        const planned = plannedNames.get(container) || new Set();
        let viewName = base;
        for (let suffix = 2; nameInScope(statement, viewName) || planned.has(viewName); suffix++) viewName = base + suffix;
        planned.add(viewName);
        plannedNames.set(container, planned);
        views.set(key, { file, sourceFile, statement, receiverText, viewName, members: new Map() });
    }
    const view = views.get(key);
    const optional = checker.getSymbolAtLocation(node.name).flags & ts.SymbolFlags.Optional ? '?' : '';
    view.members.set(name, `readonly ${name}${optional}: (typeof ${receiverText})["${name}"]`);
    addEdit(file, { start: node.expression.getStart(sourceFile), end: node.expression.end, text: view.viewName });
    // A presence test in an if's condition narrows the member for the block it guards, so every read of the same
    // member on the same receiver in that block goes through the view too, calls included: the view is the receiver's
    // own object, so a call through it passes the same this, and the narrowing holds.
    if (ts.isIfStatement(statement) && node.pos >= statement.expression.pos && node.end <= statement.expression.end) {
        const symbol = checker.getSymbolAtLocation(node.name);
        (function visit(inner) {
            if (ts.isPropertyAccessExpression(inner) && inner.name.text === name && inner.expression.getText(sourceFile) === receiverText && checker.getSymbolAtLocation(inner.name) === symbol) {
                addEdit(file, { start: inner.expression.getStart(sourceFile), end: inner.expression.end, text: view.viewName });
                console.log(`guarded ${file}:${sourceFile.getLineAndCharacterOfPosition(inner.getStart(sourceFile)).line + 1} ${receiverText}.${name} reads through ${view.viewName}`);
            }
            ts.forEachChild(inner, visit);
        })(statement.thenStatement);
    }
}
// Two views of different receivers before one statement need different names.
const namesByStatement = new Map();
for (const view of views.values()) {
    if (view === true) continue;
    const statementKey = `${view.file}|${view.statement.getStart(view.sourceFile)}`;
    const taken = namesByStatement.get(statementKey) || new Set();
    if (taken.has(view.viewName)) refusals.push(`${view.file}: two views named ${view.viewName} before one statement`);
    taken.add(view.viewName);
    namesByStatement.set(statementKey, taken);
    const start = view.statement.getStart(view.sourceFile);
    const indentation = ' '.repeat(view.sourceFile.getLineAndCharacterOfPosition(start).character);
    addEdit(view.file, { start, end: start, text: `const ${view.viewName}: { ${[...view.members.values()].join('; ')} } = ${view.receiverText};\n${indentation}` });
}
if (refusals.length) { for (const refusal of refusals) console.log('refused ' + refusal); console.log(`refused ${refusals.length}, nothing written`); process.exit(1); }
let total = 0;
for (const [file, fileEdits] of edits) {
    const name = path.join(tree, file), text = fs.readFileSync(name, 'utf8');
    // The file keeps its own line ending: positions are the checker's, on the file's own bytes, and inserted lines end
    // the way the file's lines do. A file mixing endings is refused.
    const crlf = text.includes('\r\n');
    if (crlf ? /(?<!\r)\n/.test(text) : text.includes('\r')) throw new Error(`${file}: mixed line endings`);
    let result = text;
    for (const edit of fileEdits.sort((a, b) => b.start - a.start || b.end - a.end)) result = result.slice(0, edit.start) + (crlf ? edit.text.replace(/\n/g, '\r\n') : edit.text) + result.slice(edit.end);
    const parsed = ts.createSourceFile(file, result, ts.ScriptTarget.Latest, true);
    if (parsed.parseDiagnostics.length) throw new Error(`${file}: the rewrite doesn't parse: ${parsed.parseDiagnostics[0].messageText}`);
    total += fileEdits.length;
    if (mode !== '--dry-run') fs.writeFileSync(name, result);
    console.log(`codemod ${file}: ${fileEdits.length} edits`);
}
console.log(`codemod: ${sites.length} sites, ${[...views.values()].filter(view => view !== true).length} views, ${total} edits in ${edits.size} files${mode === '--dry-run' ? ' (dry run)' : ''}`);
