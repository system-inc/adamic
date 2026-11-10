'use strict';
// Compare the complete composed pipeline, independently of adapter line offsets.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const cp = require('node:child_process'), assert = require('node:assert/strict');
const ts = require('typescript');
const {plan, owner, auditBrands} = require('./classes.cjs');
const rules = require('./class-rules.json').brands;
const [mode, tree, proof] = process.argv.slice(2);
const resolved = require('./resolved-config-brand.cjs');
assert.equal(ts.version, '6.0.3');
assert(tree && proof && ['prepare', 'before', 'after', 'audit', 'resolved-before', 'resolved-after'].includes(mode), 'usage: prepare|before|after|audit tree proof');
const walk = d => fs.readdirSync(d, {withFileTypes:true}).flatMap(e =>
    e.isDirectory() ? walk(path.join(d, e.name)) : [path.join(d, e.name)]).sort();
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const files = [...new Set(rules.map(r => r.file))];
const contracts = files.flatMap(file => plan(cp.execFileSync('git', ['-C', tree, 'show', 'HEAD:' + file], {encoding:'utf8', maxBuffer:32*1024*1024}), file).contracts);
assert.equal(contracts.length, 36);
function project(text, file, preparing = false) {
    const src = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true), edits = [];
    function visit(n) {
        if ((ts.isPropertySignature(n) || ts.isPropertyDeclaration(n)) && n.type &&
            contracts.some(c => c.key === n.name.text && c.owner === owner(n)) &&
            (preparing ? n.type.kind === ts.SyntaxKind.UndefinedKeyword :
                n.type.kind === ts.SyntaxKind.VoidKeyword)) edits.push([n.type.getStart(src), n.type.end]);
        ts.forEachChild(n, visit);
    }
    visit(src);
    for (const [a,b] of edits.sort((a,b) => b[0]-a[0])) text = text.slice(0,a) + (preparing ? 'void' : 'undefined') + text.slice(b);
    return {text, count:edits.length};
}
function snapshot() {
    const sources = {}, js = {}, declarations = {};
    for (const f of walk(path.join(tree, 'src/compiler')).filter(f => f.endsWith('.ts')))
        sources[path.relative(tree,f)] = fs.readFileSync(f,'utf8');
    for (const f of walk(path.join(tree,'built'))) {
        const name = path.relative(path.join(tree,'built'),f);
        if (/\.(js|mjs|cjs)$/.test(f)) js[name] = hash(fs.readFileSync(f));
        if (/\.d\.ts$/.test(f)) declarations[name] = fs.readFileSync(f,'utf8');
    }
    assert(Object.keys(js).length > 0);
    return {sources, js, declarations, reference:fs.readFileSync(path.join(tree,'tests/baselines/reference/api/typescript.d.ts'),'utf8')};
}
fs.mkdirSync(proof,{recursive:true});
if (mode === 'resolved-before' || mode === 'resolved-after') {
    const state = snapshot();
    if (mode === 'resolved-before') {
        assert.equal(resolved.plan(state.sources[resolved.rule.file]).state, 'before');
        fs.writeFileSync(path.join(proof, 'resolved-state-before.json'), JSON.stringify(state));
    } else {
        const before = JSON.parse(fs.readFileSync(path.join(proof, 'resolved-state-before.json')));
        assert.deepEqual(state.js, before.js, 'every emitted JavaScript artifact');
        assert.deepEqual(Object.keys(state.sources), Object.keys(before.sources));
        for (const [file, text] of Object.entries(before.sources))
            assert.equal(state.sources[file], file === resolved.rule.file ? resolved.plan(text).text : text, file);
        const project = text => text.replace('_isResolvedConfigFileName: never;', '_isResolvedConfigFileName: undefined;');
        assert.notEqual(project(before.reference), before.reference, 'one public brand line');
        assert.equal(state.reference, before.reference, 'the reference remains upstream for the sanctions path');
        assert.deepEqual(Object.keys(state.declarations), Object.keys(before.declarations));
        const changedDeclarations = [];
        for (const [file, text] of Object.entries(before.declarations)) {
            assert.equal(state.declarations[file], project(text), 'only resolved-config brand: ' + file);
            if (state.declarations[file] !== text) changedDeclarations.push(file);
        }
        const source = before.sources[resolved.rule.file];
        assert.throws(() => resolved.plan(source.replace('_isResolvedConfigFileName: never', '_isResolvedConfigFileName: string')));
        assert.equal(resolved.plan(resolved.plan(source).text).files, 0);
        const report = {class: resolved.rule.class, sourceSites: 1, publicLines: 1,
            jsFiles: Object.keys(state.js).length, emittedJavaScriptByteIdentical: true,
            declarationFiles: Object.keys(state.declarations).length, changedDeclarations,
            referenceUnchanged: true, ownerDriftRejected: true, idempotent: true};
        fs.writeFileSync(path.join(proof, 'resolved-proof.json'), JSON.stringify(report, null, 2) + '\n');
        fs.writeFileSync(path.join(proof, 'resolved-before-js.json'), JSON.stringify(before.js, null, 2) + '\n');
        fs.writeFileSync(path.join(proof, 'resolved-after-js.json'), JSON.stringify(state.js, null, 2) + '\n');
        console.log(JSON.stringify(report));
    }
    process.exit(0);
}
if (mode === 'audit') {
    const state = {sources:Object.fromEntries(walk(path.join(tree,'src/compiler')).filter(f => f.endsWith('.ts')).map(f => [path.relative(tree,f),fs.readFileSync(f,'utf8')]))};
    const keys = [...new Set(rules.map(r => r.key))];
    const allDeclarations = Object.entries(state.sources).flatMap(([file,text]) =>
        auditBrands(text,file).map(d => ({file,...d})));
    const audit = keys.map(key => {
        const search = cp.execFileSync('rg',['-n','-F',key,path.join(tree,'src/compiler')],{encoding:'utf8'});
        const declarations = allDeclarations.filter(d => d.key === key);
        assert.equal(declarations.length,search.trim().split('\n').length);
        return {key,search,declarations,reads:0,writes:0,tests:0,decision:'undefined'};
    });
    fs.writeFileSync(path.join(proof,'member-audit.json'),JSON.stringify(audit,null,2)+'\n');
    assert.equal(audit.length,31);
    assert.equal(audit.reduce((n,r)=>n+r.declarations.length,0),36);
    // Audit this amendment throughout the checkout, including tests and scripts.
    // Ignore installed dependencies, Git metadata and generated build products.
    function checkoutFiles(directory) {
        return fs.readdirSync(directory, {withFileTypes:true}).flatMap(entry => {
            const file = path.join(directory, entry.name);
            if (entry.isDirectory()) return ['.git', 'node_modules', 'built'].includes(entry.name) ? [] : checkoutFiles(file);
            return entry.isFile() ? [file] : [];
        });
    }
    const resolvedDeclarations = [];
    for (const file of checkoutFiles(tree)) {
        const text = fs.readFileSync(file, 'utf8');
        if (!text.includes(resolved.rule.key)) continue;
        assert(/\.(?:[cm]?[jt]sx?)$/.test(file), 'unparsed marker occurrence: ' + file);
        const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
        let count = 0;
        function visit(node) {
            if ((ts.isIdentifier(node) || ts.isStringLiteral(node)) && node.text === resolved.rule.key) {
                const member = node.parent;
                assert(ts.isPropertySignature(member) && member.name === node, 'runtime resolved-config marker: ' + file);
                assert.equal(owner(member), resolved.rule.owner);
                count++;
                resolvedDeclarations.push({file:path.relative(tree,file), owner:owner(member),
                    line:source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1});
            }
            ts.forEachChild(node, visit);
        }
        visit(source);
        assert.equal(count, text.split(resolved.rule.key).length - 1, 'every textual occurrence parsed');
    }
    assert.deepEqual(resolvedDeclarations.map(d=>d.file).sort(), ['lib/typescript.d.ts', resolved.rule.file, 'tests/baselines/reference/api/typescript.d.ts']);
    const amendment = {key:resolved.rule.key, declarations:resolvedDeclarations, sourceDeclarations:1,
        publicDeclarations:2, reads:0, writes:0, tests:0, scope:'whole checkout except dependencies, Git metadata and generated builds'};
    fs.writeFileSync(path.join(proof,'resolved-member-audit.json'),JSON.stringify(amendment,null,2)+'\n');
    console.log(JSON.stringify({members:audit.length + 1,sites:37,reads:0,writes:0,tests:0,amendment}));
    process.exit(0);
}
if (mode === 'prepare') {
    let count = 0;
    for (const file of files) {
        const name = path.join(tree,file), text = fs.readFileSync(name,'utf8');
        const result = project(text,file,true); count += result.count;
        if (result.text !== text) fs.writeFileSync(name,result.text);
    }
    const name = path.join(tree,'tests/baselines/reference/api/typescript.d.ts');
    const result = project(fs.readFileSync(name,'utf8'),'api',true);
    assert.equal(count,36);assert.equal(result.count,27);
    fs.writeFileSync(name,result.text);
    console.log(JSON.stringify({restoredSourceSites:count,restoredPublicLines:result.count}));
} else if (mode === 'before') {
    const state = snapshot();
    assert.equal(Object.entries(state.sources).reduce((n,[file,text]) => n + project(text,file).count,0),36);
    assert.equal(project(state.reference,'api').count,27);
    fs.writeFileSync(path.join(proof,'state-before.json'),JSON.stringify(state));
    fs.writeFileSync(path.join(proof,'before-js.json'),JSON.stringify(state.js,null,2)+'\n');
    console.log(JSON.stringify({sources:Object.keys(state.sources).length,jsFiles:Object.keys(state.js).length,brandSites:36,publicLines:27}));
} else {
    const before = JSON.parse(fs.readFileSync(path.join(proof,'state-before.json'))), after = snapshot();
    assert.deepEqual(after.js,before.js,'all emitted JavaScript bytes and file names');
    assert.deepEqual(Object.keys(after.sources),Object.keys(before.sources));
    for (const [file,text] of Object.entries(before.sources))
        assert.equal(after.sources[file],project(text,file).text,'only brand annotations: '+file);
    assert.deepEqual(Object.keys(after.declarations),Object.keys(before.declarations));
    const changedDeclarations = [];
    for (const [file,text] of Object.entries(before.declarations)) {
        assert.equal(after.declarations[file],project(text,file).text,'declaration projection: '+file);
        if (after.declarations[file] !== text) changedDeclarations.push(file);
    }
    assert.equal(after.reference,project(before.reference,'api').text);
    const mutants = [];
    const first = files[0];
    assert.throws(() => assert.equal(after.sources[first]+'\n',project(before.sources[first],first).text));
    mutants.push('unreviewed source byte');
    function killed(name, fn) { assert.throws(fn); mutants.push(name); }
    for (const [name,text] of [
        ['runtime read','value.__pathBrand;'],
        ['runtime write','value.__pathBrand = undefined;'],
        ['runtime test','"__pathBrand" in value;'],
        ['materialized field','class Links { _symbolLinksBrand: void; }']
    ]) killed(name,() => auditBrands(text,'mutant.a'));
    const changed = {...after.js, [Object.keys(after.js)[0]]:hash('mutant')};
    killed('changed JavaScript bytes',() => assert.deepEqual(changed,before.js));
    const missing = {...after.js};delete missing[Object.keys(missing)[0]];
    killed('missing JavaScript file',() => assert.deepEqual(missing,before.js));
    killed('unrelated public API line',() => assert.equal(after.reference+'\ntype Mutant = string;\n',project(before.reference,'api').text));
    const original = cp.execFileSync('git',['-C',tree,'show','HEAD:src/compiler/builder.ts'],{encoding:'utf8'});
    killed('unreviewed owner type',() => plan(original.replace('__incrementalBuildInfoFileIdBrand: any','__incrementalBuildInfoFileIdBrand: string'),'src/compiler/builder.ts'));
    const touched = files.map(file => ({file, before:hash(before.sources[file]),after:hash(after.sources[file]),
        standaloneJavaScriptIdentical:ts.transpileModule(before.sources[file],{compilerOptions:{target:ts.ScriptTarget.ESNext}}).outputText ===
            ts.transpileModule(after.sources[file],{compilerOptions:{target:ts.ScriptTarget.ESNext}}).outputText}));
    assert(touched.every(f => f.standaloneJavaScriptIdentical));
    for (const file of files) {
        const original = cp.execFileSync('git',['-C',tree,'show','HEAD:'+file],{encoding:'utf8',maxBuffer:32*1024*1024});
        const first = plan(original,file);
        assert.equal(plan(first.text,file).text,first.text,'brand plan idempotence');
    }
    const sorted = cp.execFileSync('git',['-C',tree,'show','HEAD:src/compiler/corePublic.ts'],{encoding:'utf8'}).replaceAll(': any;',': undefined;');
    assert.equal(plan(sorted,'src/compiler/corePublic.ts').text,sorted,'required undefined sorted marker left alone');
    const report = {brandSites:36,publicLines:27,jsFiles:Object.keys(after.js).length,
        declarationFiles:Object.keys(after.declarations).length,changedDeclarations,touched,mutants};
    fs.writeFileSync(path.join(proof,'after-js.json'),JSON.stringify(after.js,null,2)+'\n');
    fs.writeFileSync(path.join(proof,'proof.json'),JSON.stringify(report,null,2)+'\n');
    console.log(JSON.stringify(report,null,2));
}
