'use strict';
// Evidence only. No source mutation or compiler change.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const options = { strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    verbatimModuleSyntax: true, noEmit: true, allowImportingTsExtensions: true,
    module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [] };
function createAudit(program, root) {
    const checker = program.getTypeChecker();
    const references = new Map();
    const sourceFiles = program.getSourceFiles().filter(f => !f.isDeclarationFile && f.fileName.startsWith(root + path.sep));
    function where(n) { const f = n.getSourceFile(), p = f.getLineAndCharacterOfPosition(n.getStart(f));
        return `${path.relative(root, f.fileName)}:${p.line + 1}:${p.character + 1}`; }
    function text(n) { return n.getText().slice(0, 240); }
    function symbol(n) { let s = ts.isShorthandPropertyAssignment(n.parent) ? checker.getShorthandAssignmentValueSymbol(n.parent) : checker.getSymbolAtLocation(n);
        if (s?.flags & ts.SymbolFlags.Alias) s = checker.getAliasedSymbol(s); return s; }
    function type(t) { return checker.typeToString(t, undefined, ts.TypeFormatFlags.NoTruncation); }
    function primitive(t) { return t.isUnion() ? t.types.every(primitive) :
        !!(t.flags & (ts.TypeFlags.StringLike | ts.TypeFlags.NumberLike | ts.TypeFlags.BooleanLike |
            ts.TypeFlags.BigIntLike | ts.TypeFlags.ESSymbolLike | ts.TypeFlags.Null | ts.TypeFlags.Undefined | ts.TypeFlags.Never)); }
    function fn(n) { for (let p = n.parent; p; p = p.parent) if (ts.isFunctionLike(p)) return p; }
    function wrapper(n) { return ts.isParenthesizedExpression(n) || ts.isNonNullExpression(n) ||
        ts.isAsExpression(n) || ts.isTypeAssertionExpression(n) || ts.isSatisfiesExpression(n); }
    function assignment(n) { return ts.isBinaryExpression(n) && n.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && n.operatorToken.kind <= ts.SyntaxKind.LastAssignment; }
    function visit(n) {
        if (ts.isTypeNode(n) || ts.isInterfaceDeclaration(n) || ts.isTypeAliasDeclaration(n)) return;
        if (ts.isIdentifier(n)) {
            const p = n.parent;
            const declaration = p.name === n && (ts.isVariableDeclaration(p) || ts.isParameter(p) || ts.isBindingElement(p) || ts.isFunctionDeclaration(p) || ts.isFunctionExpression(p) || ts.isClassDeclaration(p) || ts.isMethodDeclaration(p) || ts.isPropertyDeclaration(p));
            const key = (ts.isPropertyAccessExpression(p) && p.name === n) || ts.isImportSpecifier(p) || ts.isExportSpecifier(p);
            if (!declaration && !key) { const s = symbol(n); if (s) { if (!references.has(s)) references.set(s, []); references.get(s).push(n); } }
        }
        ts.forEachChild(n, visit);
    }
    sourceFiles.forEach(visit);
    function locate(site) {
        const [file, l, c] = site.where.split(':'); const sf = program.getSourceFile(path.join(root, file));
        if (!sf) return { failure: 'source file missing' };
        const position = sf.getPositionOfLineAndCharacter(Number(l) - 1, Number(c) - 1);
        const wanted = site.reason.split(' seen as ')[0].replace(/^a value of type /, '');
        const candidates = [];
        function scan(n) {
            if (position < n.pos || position >= n.end) return;
            if (ts.isExpressionNode(n) && n.getStart(sf) <= position) candidates.push(n);
            ts.forEachChild(n, scan);
        }
        scan(sf);
        const normalize = s => s.replace(/\s+/g, '').replace(/readonly/g, 'readonly');
        const matching = candidates.filter(n => [checker.typeToString(checker.getTypeAtLocation(n)), type(checker.getTypeAtLocation(n))].some(t => normalize(t) === normalize(wanted)));
        matching.sort((a,b) => (a.end-a.getStart(sf)) - (b.end-b.getStart(sf)));
        if (!matching.length) return { failure: 'census source type does not map uniquely to a stock-checker expression',
            candidates: candidates.slice(-5).map(n => ({ where: where(n), expression: text(n), type: type(checker.getTypeAtLocation(n)) })) };
        return { node: matching[0], original: checker.getTypeAtLocation(matching[0]) };
    }
    function inspect(node, original) {
        const reads = [], writes = [], unresolved = [], calls = [], aliases = [];
        const visited = new Set(); let steps = 0;
        const unknown = (n, reason) => unresolved.push({ where: where(n), reason, expression: text(n) });
        function domain(t, fields) {
            if (!fields.length) return t;
            if (t.isUnion()) {
                const members = t.types.filter(v => !(v.flags & (ts.TypeFlags.Null | ts.TypeFlags.Undefined)));
                const results = members.map(v => domain(v, fields));
                return results.every(Boolean) ? results : undefined;
            }
            const [key, ...rest] = fields;
            const next = key === '[]' ? checker.getIndexTypeOfType(t, ts.IndexKind.Number) : checker.getTypeOfPropertyOfType(t, key);
            return next && domain(next, rest);
        }
        function flatten(v) { return Array.isArray(v) ? v.flatMap(flatten) : v ? [v] : []; }
        function conditional(n, scope) {
            for (let p=n.parent; p && p!==scope; p=p.parent) if (ts.isIfStatement(p) || ts.isConditionalExpression(p) || ts.isSwitchStatement(p) ||
                ts.isForStatement(p) || ts.isWhileStatement(p) || ts.isForOfStatement(p) || ts.isTryStatement(p)) return true;
            return false;
        }
        function disjoint(actual, expected) {
            if (expected.isUnion()) return expected.types.every(t => disjoint(actual, t));
            if (actual.isUnion()) return actual.types.every(t => disjoint(t, expected));
            if (expected.flags & ts.TypeFlags.Never) return !(actual.flags & ts.TypeFlags.Never);
            if (actual.flags & ts.TypeFlags.Undefined) return !(expected.flags & (ts.TypeFlags.Undefined | ts.TypeFlags.Void | ts.TypeFlags.Any | ts.TypeFlags.Unknown));
            if (actual.flags & ts.TypeFlags.Null) return !(expected.flags & (ts.TypeFlags.Null | ts.TypeFlags.Any | ts.TypeFlags.Unknown));
            if (actual.isLiteral() && expected.isLiteral()) return actual.value !== expected.value;
            const kinds = [ts.TypeFlags.NumberLike, ts.TypeFlags.StringLike, ts.TypeFlags.BooleanLike, ts.TypeFlags.BigIntLike];
            return kinds.some(k => actual.flags & k && kinds.some(q => q !== k && expected.flags & q));
        }
        function write(n, fields, rhs, env, kind, scope) {
            const expected = flatten(domain(original, fields));
            let actual;
            const generic = !!(original.flags & ts.TypeFlags.TypeParameter);
            const bound = generic && checker.getBaseConstraintOfType(original);
            const bounded = generic ? flatten(bound && domain(bound, fields)) : expected;
            if (generic) expected.splice(0, expected.length, ...bounded);
            while (rhs && wrapper(rhs)) rhs = rhs.expression;
            if (rhs && ts.isIdentifier(rhs) && env.has(symbol(rhs))) actual = checker.getTypeAtLocation(env.get(symbol(rhs)));
            else if (rhs) actual = checker.getTypeAtLocation(rhs);
            if (!expected.length || !actual || actual.flags & (ts.TypeFlags.Any | ts.TypeFlags.Unknown | ts.TypeFlags.TypeParameter | ts.TypeFlags.IndexedAccess | ts.TypeFlags.Conditional) ||
                expected.some(t => t.flags & (ts.TypeFlags.Any | ts.TypeFlags.TypeParameter | ts.TypeFlags.IndexedAccess | ts.TypeFlags.Conditional))) {
                unknown(n, 'write domain unresolved, generic, or unchecked'); return;
            }
            const relation = expected.map(t => checker.isTypeAssignableTo(actual, t));
            const excludesAll = expected.every(t => disjoint(actual, t));
            let counterexample;
            if (generic && !excludesAll && fields.length === 1 && fields[0] === 'flags' && kind === '=') {
                const property = checker.getPropertyOfType(bound, 'flags');
                const readonly = property?.declarations?.some(d => ts.getCombinedModifierFlags(d) & ts.ModifierFlags.Readonly);
                const leaves = t => t.isUnion() ? t.types.flatMap(leaves) : [t];
                const read = leaves(expected[0]).find(t => t.isLiteral() && t.value === 16);
                const value = leaves(actual).find(t => t.isLiteral() && t.value === 0);
                let feasible = true;
                for (let p=n.parent; p && p!==scope; p=p.parent) {
                    if (ts.isIfStatement(p)) {
                        if (!ts.isIdentifier(p.expression) || !ts.isIdentifier(node) || symbol(p.expression)!==symbol(node) || n.getStart()<p.thenStatement.getStart() || n.end>p.thenStatement.end) feasible=false;
                    } else if (ts.isConditionalExpression(p) || ts.isSwitchStatement(p) || ts.isForStatement(p) || ts.isWhileStatement(p) || ts.isForOfStatement(p) || ts.isTryStatement(p)) feasible=false;
                }
                if (readonly && read && value && feasible) counterexample = {kind:'readonly generic-field instantiation', holder_field:type(read), written_field:type(value), witness:'controls/tsc-writers.a#flags', guard:'object-presence only'};
            }
            if (generic && !excludesAll && !counterexample) { unknown(n, 'generic original holder may narrow the constraint field'); return; }
            const guarded = conditional(n, scope);
            const record = { where: where(n), path: fields.join('.'), operation: kind,
                original_domains: expected.map(type), written_type: type(actual), value: rhs && text(rhs),
                compatible: counterexample ? false : relation.every(Boolean), definitely_outside: excludesAll || !!counterexample, conditional: guarded,
                original_domain_is_generic_upper_bound: generic, counterexample, feasible_counterexample: !!counterexample };
            if (expected.some(t => t.flags & ts.TypeFlags.EnumLike) && actual.flags & ts.TypeFlags.Number && !(actual.flags & ts.TypeFlags.NumberLiteral)) {
                unknown(n, 'numeric enum write requires a value-domain proof'); record.compatible = null;
            }
            if (!record.compatible && !counterexample && (guarded || expected.length > 1 || !record.definitely_outside)) unknown(n, 'incompatible write needs value/branch/union-variant feasibility proof');
            writes.push(record);
        }
        function trackBinding(binding, fields, env, callStack) {
            if (!ts.isIdentifier(binding.name)) { unknown(binding, 'destructured alias'); return; }
            const s = symbol(binding.name), scope = ts.isParameter(binding) ? binding.parent : fn(binding);
            if (!s || !scope) { unknown(binding, 'global or unresolved alias'); return; }
            const key = `${where(binding)}|${fields.join('.')}|${callStack.map(c=>where(c)).join('>')}`;
            if (visited.has(key)) { unknown(binding, 'recursive or cyclic alias graph'); return; } visited.add(key);
            const uses = references.get(s) || [];
            if (uses.some(n => assignment(n.parent) && n.parent.left===n || (ts.isPrefixUnaryExpression(n.parent) || ts.isPostfixUnaryExpression(n.parent)) && [ts.SyntaxKind.PlusPlusToken,ts.SyntaxKind.MinusMinusToken].includes(n.parent.operator))) {
                unknown(binding, 'reassigned alias requires flow-sensitive identity tracking'); return;
            }
            aliases.push({ where: where(binding), name: binding.name.text, path: fields.join('.'), references: uses.length });
            for (const n of uses) {
                if (fn(n)!==scope) { unknown(n, 'alias captured by a nested callback/closure'); continue; }
                use(n, fields, env, callStack, scope);
            }
        }
        function consume(call, argument, fields, env, callStack, scope) {
            const index = call.arguments.indexOf(argument);
            if (index<0 || ts.isSpreadElement(argument)) { unknown(call, 'spread or non-positional argument'); return; }
            const signature = checker.getResolvedSignature(call);
            const callee = ts.isPropertyAccessExpression(call.expression) ? symbol(call.expression.name) : symbol(call.expression);
            const declarations = callee?.declarations || [];
            const bodies = declarations.filter(d => (ts.isFunctionDeclaration(d) || ts.isFunctionExpression(d) || ts.isArrowFunction(d)) && d.body);
            for (const d of declarations) if (ts.isVariableDeclaration(d) && d.initializer && (ts.isArrowFunction(d.initializer) || ts.isFunctionExpression(d.initializer))) bodies.push(d.initializer);
            if (bodies.length!==1 || !signature || bodies[0].getSourceFile().isDeclarationFile) {
                unknown(call, 'unresolved, external, virtual, or callback consumer'); return;
            }
            const body=bodies[0], parameter=body.parameters[index];
            if (!parameter || parameter.dotDotDotToken) { unknown(call, 'rest/container argument or missing implementation parameter'); return; }
            if (callStack.length>=16 || callStack.some(c=>checker.getResolvedSignature(c)?.declaration===signature.declaration)) { unknown(call, 'recursive call or depth bound'); return; }
            const next = new Map(env);
            body.parameters.forEach((p,i) => { if (ts.isIdentifier(p.name) && call.arguments[i]) next.set(symbol(p.name),call.arguments[i]); });
            calls.push({ where: where(call), target: where(body), parameter: where(parameter), path: fields.join('.') });
            trackBinding(parameter, fields, next, [...callStack,call]);
        }
        function use(n, fields, env, callStack, scope) {
            if (++steps>600) { if (steps===601) unknown(n,'per-site traversal bound'); return; }
            const p=n.parent;
            if (!p) { unknown(n,'no receiving syntax'); return; }
            if (wrapper(p)) { use(p, fields, env, callStack, scope); return; }
            if ((ts.isPropertyAccessExpression(p) || ts.isElementAccessExpression(p)) && p.expression===n) {
                let field;
                if (ts.isPropertyAccessExpression(p)) {
                    field=p.name.text; const s=checker.getSymbolAtLocation(p.name);
                    if (s?.declarations?.some(d=>ts.isGetAccessorDeclaration(d) || ts.isSetAccessorDeclaration(d))) { unknown(p,'accessor effects'); return; }
                } else field=ts.isStringLiteral(p.argumentExpression) ? p.argumentExpression.text : '[]';
                const parent=p.parent;
                if (ts.isCallExpression(parent) && parent.expression===p) {
                    const s=ts.isPropertyAccessExpression(p) && checker.getSymbolAtLocation(p.name);
                    const stock=s?.declarations?.some(d=>d.getSourceFile().isDeclarationFile && path.basename(d.getSourceFile().fileName).startsWith('lib.'));
                    if (stock && ['push','unshift','fill','splice'].includes(field)) {
                        const values=field==='splice'?parent.arguments.slice(2):field==='fill'?parent.arguments.slice(0,1):parent.arguments;
                        values.forEach(v=>write(parent,[...fields,'[]'],v,env,field,scope));
                        if (!values.length) reads.push({where:where(parent),operation:'removal-only array operation'});
                        if (!ts.isExpressionStatement(parent.parent)) unknown(parent,'array mutator result or element aliases escape');
                        return;
                    }
                    unknown(parent,'method receiver or element/callback effects'); return;
                }
                if (assignment(parent) && parent.left===p) {
                    write(parent,[...fields,field],parent.operatorToken.kind===ts.SyntaxKind.EqualsToken?parent.right:parent,env,ts.tokenToString(parent.operatorToken.kind),scope); return;
                }
                if (ts.isDeleteExpression(parent)) { unknown(parent,'delete needs required-property and alias proof'); return; }
                if ((ts.isPrefixUnaryExpression(parent)||ts.isPostfixUnaryExpression(parent)) && [ts.SyntaxKind.PlusPlusToken,ts.SyntaxKind.MinusMinusToken].includes(parent.operator)) {
                    write(parent,[...fields,field],parent,env,'increment/decrement',scope);return;
                }
                const t=checker.getTypeAtLocation(p);
                if (primitive(t)) { reads.push({where:where(p),operation:'primitive field read',path:[...fields,field].join('.')}); return; }
                use(p,[...fields,field],env,callStack,scope); return;
            }
            if (ts.isVariableDeclaration(p) && p.initializer===n) { trackBinding(p,fields,env,callStack); return; }
            if (ts.isCallExpression(p) && p.arguments.includes(n)) { consume(p,n,fields,env,callStack,scope); return; }
            if (ts.isReturnStatement(p) || ts.isArrowFunction(p) && p.body===n) {
                if (callStack.length && fn(n)===scope) {
                    const caller=callStack[callStack.length-1];use(caller,fields,env,callStack.slice(0,-1),fn(caller));
                } else unknown(p,'returned alias escapes receiving scope'); return;
            }
            if (assignment(p) && p.right===n) {
                if (ts.isIdentifier(p.left)) unknown(p,'assignment into an existing binding needs flow-sensitive alias tracking');
                else unknown(p,'stored into a field/container');return;
            }
            if (ts.isBinaryExpression(p)) {
                const op=p.operatorToken.kind;
                if ([ts.SyntaxKind.AmpersandAmpersandToken,ts.SyntaxKind.BarBarToken,ts.SyntaxKind.QuestionQuestionToken].includes(op)) {
                    const objectOnly = (t=>t.isUnion()?t.types.every(v=>primitive(v)||!!(v.flags & ts.TypeFlags.Object)):primitive(t)||!!(t.flags&ts.TypeFlags.Object))(checker.getTypeAtLocation(n));
                    if (p.left===n && op===ts.SyntaxKind.AmpersandAmpersandToken && objectOnly && !fields.length &&
                        (t=>t.isUnion()?t.types.every(v=>!!(v.flags&(ts.TypeFlags.Object|ts.TypeFlags.Undefined|ts.TypeFlags.Null))):!!(t.flags&ts.TypeFlags.Object))(checker.getTypeAtLocation(n))) {
                        reads.push({where:where(n),operation:'truthiness only; object is not the && result'});return;
                    }
                    use(p,fields,env,callStack,scope);return;
                }
                if ([ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken,ts.SyntaxKind.EqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsToken,
                    ts.SyntaxKind.InKeyword,ts.SyntaxKind.InstanceOfKeyword].includes(op)) { reads.push({where:where(n),operation:'identity/presence test'});return; }
                unknown(p,'unsupported binary use');return;
            }
            if (ts.isConditionalExpression(p)) {
                if (p.condition===n) reads.push({where:where(n),operation:'condition read'});
                else use(p,fields,env,callStack,scope);return;
            }
            if ((ts.isIfStatement(p)||ts.isWhileStatement(p)||ts.isDoStatement(p)) && p.expression===n ||
                ts.isTypeOfExpression(p) || ts.isVoidExpression(p) || ts.isPrefixUnaryExpression(p)&&p.operator===ts.SyntaxKind.ExclamationToken || ts.isExpressionStatement(p)) {
                reads.push({where:where(n),operation:'discard/control read'}); return;
            }
            unknown(p,ts.isArrayLiteralExpression(p)||ts.isObjectLiteralExpression(p)||ts.isPropertyAssignment(p)||ts.isSpreadElement(p)?'container construction or shared element alias':`unsupported receiving syntax ${ts.SyntaxKind[p.kind]}`);
        }
        use(node,[],new Map(),[],fn(node));
        const bad=writes.filter(w=>w.compatible===false && w.definitely_outside && (!w.conditional || w.feasible_counterexample) && w.original_domains.length===1);
        const classification = bad.length ? 'c' : unresolved.length ? 'd' : writes.length ? 'b' : 'a';
        return { classification, source_expression: text(node), source_expression_where: where(node), original_type: type(original),
            receiving_syntax: ts.SyntaxKind[node.parent.kind], reads, writes, aliases, calls, unresolved, steps };
    }
    return { checker, sourceFiles, inspect, locate, where, type, references };
}
module.exports = { createAudit, options, ts };
if (require.main===module) {
    const [treeArg,sitesArg,outputArg]=process.argv.slice(2); if (!outputArg) throw Error('usage: node audit.cjs <tree> <sites.json> <output.json>');
    if(ts.version!=='6.0.3')throw Error('expected TypeScript 6.0.3');
    const tree=path.resolve(treeArg),sites=JSON.parse(fs.readFileSync(sitesArg,'utf8'));
    const program=ts.createProgram(ts.sys.readDirectory(path.join(tree,'src'),['.ts'],['**/lib/**']),options);
    const audit=createAudit(program,tree);
    const records=sites.map((site,index)=>{
        const sourceFamily=site.reason.split(' seen as ')[0].replace(/^a value of type /,'');
        const family=sourceFamily.startsWith('DiagnosticWith')?'diagnostics':sourceFamily==='never[]'?'shared-never':'other';
        const found=audit.locate(site);
        const result=found.node?audit.inspect(found.node,found.original):{ classification:'d',unresolved:[{where:site.where,reason:found.failure,candidates:found.candidates}],reads:[],writes:[],calls:[],aliases:[] };
        return { id:index+1,where:site.where,reason:site.reason,family,source_family:sourceFamily,...result };
    });
    const counts={a:0,b:0,c:0,d:0},families={},sourceFamilies={};
    for(const r of records){counts[r.classification]++;for(const [dict,key] of [[families,r.family],[sourceFamilies,r.source_family]]){dict[key]??={a:0,b:0,c:0,d:0,total:0};dict[key][r.classification]++;dict[key].total++;}}
    const hashes={};for(const sf of audit.sourceFiles)hashes[path.relative(tree,sf.fileName)]=crypto.createHash('sha256').update(sf.text).digest('hex');
    const output={schema_version:1,typescript:ts.version,source_commit:'12fef29628bfda6cf8bec4ba2a326fa4ec4a66d5',
        scope:'whole src program; exact 615 retained table-selected observations; no runtime reachability claim',
        source_files:audit.sourceFiles.length,root_files:program.getRootFileNames().length,program_files:program.getSourceFiles().length,declaration_files:program.getSourceFiles().filter(f=>f.isDeclarationFile).length,indexed_symbols:audit.references.size,counts,families,source_families:sourceFamilies,
        class_definitions:{a:'closed receiver/alias graph; no observed write',b:'closed graph; every observed write fits the original static read domains',c:'known identity path to a disjoint write or a validated legal generic instantiation; object-presence guards allowed; other escapes may remain',d:'unresolved identity, escape, callback, branch/variant feasibility, generic domain, or analysis bound'},
        examples:Object.fromEntries(['a','b','c','d'].map(k=>[k,records.filter(r=>r.classification===k).sort((a,b)=>k==='a'?b.calls.length-a.calls.length:0).slice(0,3).map(r=>r.id)])),
        c_sites:records.filter(r=>r.classification==='c').map(r=>r.id),source_sha256:hashes,records};
    fs.writeFileSync(outputArg,JSON.stringify(output,null,2)+'\n');console.log(JSON.stringify({counts,families,source_files:audit.sourceFiles.length,indexed_symbols:audit.references.size},null,2));
}
