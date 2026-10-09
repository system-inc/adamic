// Stock TypeScript's checker is the meter, never a text pattern matcher.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const cp = require('node:child_process');
const ts = require(process.env.STAGE3_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw Error(`Expected TypeScript 6.0.3, got ${ts.version}`);
const tree = path.resolve(process.argv[2]);
const out = path.resolve(process.argv[3] || __dirname);
const pin = '050880ce59e30b356b686bd3144efe24f875ebc8';
if (cp.execFileSync('git', ['-C', tree, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== pin) throw Error('Wrong upstream commit');
const configPath = path.join(tree, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configPath, ts.sys.readFile);
if (config.error) throw Error(ts.flattenDiagnosticMessageText(config.error.messageText, '\n'));
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath));
// Inspect upstream under its own options. Adamic's fixed options are used on fixtures.
parsed.options.configFilePath = configPath;
const program = ts.createProgram(parsed.fileNames, parsed.options);
const checker = program.getTypeChecker();
const typeText = t => checker.typeToString(t, undefined, ts.TypeFormatFlags.NoTruncation);
const location = n => {
    const sf = n.getSourceFile();
    return `${path.relative(tree, sf.fileName).split(path.sep).join('/')}:${sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1}`;
};
const walk = (n, f) => { f(n); ts.forEachChild(n, c => walk(c, f)); };
const families = t => {
    if (t.isUnion()) return [...new Set(t.types.flatMap(families))].sort();
    const f = ts.TypeFlags;
    if (t.flags & f.BooleanLike) return ['boolean'];
    if (t.flags & f.NumberLike) return ['number'];
    if (t.flags & f.StringLike) return ['string'];
    if (t.flags & (f.Undefined | f.Void)) return ['undefined'];
    if (t.flags & f.Null) return ['null'];
    if (t.flags & f.Never) return [];
    if (t.flags & f.Object) return ['object'];
    return ['unresolved'];
};
const uses = n => {
    const found = [];
    let child = n;
    let p = n.parent;
    // Follow the containing expression to its first statement; don't guess from strings.
    while (p && !ts.isStatement(p)) {
        if (ts.isPrefixUnaryExpression(p) && p.operator === ts.SyntaxKind.ExclamationToken) found.push('truthiness (!result)');
        if (ts.isBinaryExpression(p)) {
            const op = p.operatorToken.kind;
            if ([ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.BarBarToken].includes(op) && p.left === child) found.push('truthiness (short circuit)');
            if ([ts.SyntaxKind.EqualsEqualsEqualsToken, ts.SyntaxKind.ExclamationEqualsEqualsToken, ts.SyntaxKind.LessThanToken, ts.SyntaxKind.GreaterThanToken].includes(op)) {
                const other = p.left === child ? p.right : p.left;
                found.push(other.kind === ts.SyntaxKind.Identifier && other.text === 'undefined' ? 'undefined comparison' : `comparison: ${p.getText().replaceAll('\n',' ').replaceAll('\r','')}`);
            }
        }
        if (ts.isConditionalExpression(p) && p.condition === child) found.push('truthiness (conditional)');
        child = p;
        p = p.parent;
    }
    if (!found.some(t=>t.startsWith('comparison') || t.startsWith('undefined')) && p && (ts.isIfStatement(p) || ts.isWhileStatement(p) || ts.isDoStatement(p)) && p.expression === child) found.push('truthiness (condition)');
    if (p && ts.isSwitchStatement(p) && p.expression === child) found.push('comparison (switch case labels)');
    if (p && ts.isReturnStatement(p)) found.push('returned to caller');
    if (!found.length) found.push('value stored/passed or discarded; see source context');
    return [...new Set(found)];
};
const callNames = new Set(['lookAhead','tryParse','speculationHelper','tryScan','scanRange','resetTokenState','setTextPos']);
const rows = [], rescans = [], files = [], textResets = [];
for (const name of ['parser.ts','scanner.ts']) {
    const sf = program.getSourceFile(path.join(tree,'src/compiler',name));
    if (!sf || sf.parseDiagnostics.length) throw Error(`Cannot parse ${name}`);
    if (sf.text !== cp.execFileSync('git',['-C',tree,'show',`${pin}:src/compiler/${name}`],{encoding:'utf8',maxBuffer:4000000})) throw Error(`Modified pinned source: ${name}`);
    files.push({file:location(sf).split(':')[0], bytes:Buffer.byteLength(sf.text), sha256:crypto.createHash('sha256').update(sf.text).digest('hex')});
    // Both complete files are traversed, including nested JSDoc parsers and scanner callbacks.
    walk(sf, n => {
        if (!ts.isCallExpression(n)) return;
        const callee = ts.isPropertyAccessExpression(n.expression) ? n.expression.name.text : ts.isIdentifier(n.expression) ? n.expression.text : '';
        if (!callNames.has(callee) && !callee.startsWith('reScan') && callee !== 'setText') return;
        const row = {site:location(n), call:callee, expression:n.getText(), result:typeText(checker.getTypeAtLocation(n)), tests:uses(n)};
        if (callee === 'setText') { textResets.push(row); return; }
        if (!callNames.has(callee)) { rescans.push(row); return; }
        const callback = ['lookAhead','tryParse','speculationHelper','tryScan'].includes(callee) ? n.arguments[0] : callee === 'scanRange' ? n.arguments[2] : undefined;
        if (callback) {
            const sig = checker.getSignaturesOfType(checker.getTypeAtLocation(callback),ts.SignatureKind.Call)[0];
            if (!sig) throw Error(`Missing callback signature at ${row.site}`);
            const decl = sig.declaration;
            const result = checker.getReturnTypeOfSignature(sig);
            row.callback = callback.getText();
            row.callbackSite = decl ? location(decl) : null;
            row.declared = decl?.type ? decl.type.getText() : '(inferred, no return annotation)';
            row.inferred = typeText(result);
            row.families = families(result);
            row.bodyReturns = [];
            if (decl?.body) {
                const returns = a => {
                    if (a !== decl.body && ts.isFunctionLike(a)) return;
                    if (ts.isReturnStatement(a)) row.bodyReturns.push(a.expression ? typeText(checker.getTypeAtLocation(a.expression)) : 'undefined');
                    ts.forEachChild(a, returns);
                };
                if (ts.isBlock(decl.body)) returns(decl.body);
                else row.bodyReturns.push(typeText(checker.getTypeAtLocation(decl.body)));
            }
            row.assertions = [];
            if (decl?.body) walk(decl.body, a => {
                if (ts.isAsExpression(a) || ts.isTypeAssertionExpression(a)) {
                    row.assertions.push({site:location(a), before:typeText(checker.getTypeAtLocation(a.expression)), after:typeText(checker.getTypeAtLocation(a)), beforeFamilies:families(checker.getTypeAtLocation(a.expression))});
                }
            });
            const losesFamily = row.assertions.some(a => a.beforeFamilies.some(f => !row.families.includes(f)));
            row.classification = row.families.length === 1 && !row.families.includes('unresolved') && !losesFamily ? 'one proven runtime family' : 'checked result specialization';
            row.reason = losesFamily ? 'callback body contains an assertion dropping a runtime family; inspect before projection' : row.classification === 'checked result specialization' ? 'unresolved generic or multiple runtime families; keep truthiness separate from result projection' : 'checker establishes one representation family; literals/enums retain exact values';
            row.helperTest = callee === 'scanRange' ? 'unconditional rewind, no result test' : 'scanner !result || isLookahead; parser !result || kind !== TryParse (when used)';
        } else { row.declared = row.inferred = '(no callback)'; row.classification = 'rewind, void result'; }
        let statement = n; while(statement.parent && !ts.isStatement(statement)) statement = statement.parent;
        row.context = statement.getText();
        // Record all reads of a directly assigned local and their immediate consumers.
        let binding = n.parent;
        const assigned = ts.isBinaryExpression(binding) && binding.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isIdentifier(binding.left) ? binding.left : undefined;
        while(binding && !ts.isStatement(binding) && !ts.isVariableDeclaration(binding)) binding = binding.parent;
        if (binding && ts.isVariableDeclaration(binding) && ts.isIdentifier(binding.name)) {
            const symbol = checker.getSymbolAtLocation(binding.name);
            row.localUses = [];
            walk(sf, id => {
                if(ts.isIdentifier(id) && id !== binding.name && checker.getSymbolAtLocation(id) === symbol) row.localUses.push({site:location(id),tests:uses(id),context:id.parent.getText()});
            });
        }
        if (assigned) {
            const symbol = checker.getSymbolAtLocation(assigned);
            row.assignedUses = [];
            let scope=n.parent; while(scope.parent && !ts.isFunctionLike(scope)) scope=scope.parent;
            walk(scope,id=> {if(ts.isIdentifier(id) && id.getStart()>n.getStart() && checker.getSymbolAtLocation(id)===symbol) row.assignedUses.push({site:location(id),tests:uses(id),context:id.parent.getText()});});
        }
        if(row.site === 'src/compiler/parser.ts:5238') row.forwardedTests = ['parser.ts:5223 comparison triState === Tristate.False (0)', 'parser.ts:5227 comparison triState === Tristate.True (1)'];
        rows.push(row);
    });
}
const counts = {};
for (const r of rows) { const f = r.site.split(':')[0]; counts[f] ||= {}; counts[f][r.call] = (counts[f][r.call]||0)+1; }
const classifications = {};
for (const r of rows) classifications[r.classification] = (classifications[r.classification]||0)+1;
const diagnostics = program.getSemanticDiagnostics();
if (diagnostics.length) throw Error(`Inventory must have zero semantic diagnostics: ${diagnostics.length}`);
const data = {pin,typescript:ts.version,files,semanticDiagnostics:diagnostics.map(d=>({site:d.file ? location(d.file):null,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,' ')})), counts,classifications,rows,rescans,textResets};
fs.writeFileSync(path.join(out,'inventory.json'),JSON.stringify(data,null,2)+'\n');
const escape = s => String(s).replaceAll('|','\\|').replaceAll('\n',' ').replaceAll('\r','');
const lines = ['# Speculative parsing inventory','',`TypeScript ${ts.version}, commit ${pin}. All calls in the two complete ASTs; declarations and exported aliases are not calls.`, '', '“One” means one runtime representation family, not one literal value or one object shape. Boolean literals and numeric enums retain their exact values. “Checked” is a conservative specialization boundary, not an instruction to coerce the callback result. Generic forwarding must be instantiated at its callers. Known unions can remain tagged unions; a narrowed result projection must be checked when proof is missing. Assertions are not proof.','', 'Every lookAhead/tryParse/tryScan result is also tested by the helper using JavaScript `!result`; lookAhead always rewinds, tryParse/tryScan commit only truthy results. scanRange always restores. Full expressions, contexts, assigned-local uses, body return expression types (bodyReturns) and assertion input types are in inventory.json.','', '## Counts','', '```json',JSON.stringify({counts,classifications,rescans:rescans.length,textResets:textResets.length,semanticDiagnostics:diagnostics.length},null,2),'```','', '## Calls','', '| Site | Call | Callback declaration / checker result (annotation honored) | Result consumer | Decision |','|---|---|---|---|---|'];
for(const r of rows) lines.push(`| ${r.site} | ${r.call} | ${escape(r.declared)} / ${escape(r.inferred)}${r.callbackSite ? ' at '+r.callbackSite : ''} | ${escape([...r.tests,...(r.localUses||[]).flatMap(u=>u.tests),...(r.assignedUses||[]).flatMap(u=>u.tests),...(r.forwardedTests||[])].join('; '))} | ${r.classification}${(r.assertions||[]).some(a=>a.beforeFamilies.some(f=>!r.families.includes(f))) ? '; assertion hides a runtime alternative' : ''} |`);
lines.push('','## Supplemental scanner rescans','','These rewind/reinterpret the current token; they take no speculative callback. This includes parser wrappers and calls to them. `setTextPos` has zero calls (only an interface declaration and an exported alias of resetTokenState). setText initialization is not a speculative callback.','', '| Site | Call | Result type | Consumer |','|---|---|---|---|');
for(const r of rescans) lines.push(`| ${r.site} | ${r.call} | ${escape(r.result)} | ${escape(r.tests.join('; '))} |`);
lines.push('','## Text reset calls','','These can reposition the scanner but are initialization/range setup, not speculative callbacks.','', '| Site | Call | Result |','|---|---|---|');
for(const r of textResets) lines.push(`| ${r.site} | ${escape(r.expression)} | ${r.result} |`);
fs.writeFileSync(path.join(out,'inventory.md'),lines.join('\n')+'\n');
console.log(JSON.stringify({counts,classifications,rescans:rescans.length,textResets:textResets.length,semanticDiagnostics:diagnostics.length}));
