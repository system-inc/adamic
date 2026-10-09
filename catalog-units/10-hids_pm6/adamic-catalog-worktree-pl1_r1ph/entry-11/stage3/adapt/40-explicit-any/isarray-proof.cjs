'use strict';
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const assert = require('node:assert/strict'), ts = require('typescript');
const cp = require('node:child_process'), os = require('node:os');
const [mode,tree,proof] = process.argv.slice(2);
assert.equal(ts.version,'6.0.3');
assert(['before','after'].includes(mode) && tree && proof,'usage: before|after tree proof');
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const walk = d => fs.readdirSync(d,{withFileTypes:true}).flatMap(e => e.isDirectory()?walk(path.join(d,e.name)):[path.join(d,e.name)]).sort();
function program(replacement) {
    const file = path.join(tree,'src/compiler/tsconfig.json');
    const config = ts.parseJsonConfigFileContent(ts.readConfigFile(file,ts.sys.readFile).config,ts.sys,path.dirname(file),undefined,file);
    assert.equal(config.errors.length,0);
    const host = ts.createCompilerHost(config.options), read = host.readFile;
    if(replacement) host.readFile = name => {
        const text=read(name);
        return name === path.join(tree,'src/compiler/core.ts') ?
            text.replace(/export function isArray\(value: (any|unknown)\)/,'export function isArray(value: '+replacement+')') : text;
    };
    return ts.createProgram(config.fileNames,config.options,host);
}
function diagnostics(p) {
    return ts.getPreEmitDiagnostics(p).map(d => ({file:d.file?path.relative(tree,d.file.fileName):null,
        start:d.start,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
}
function snapshot() {
    const sources={}, js={};
    for(const file of walk(path.join(tree,'src/compiler')).filter(f=>f.endsWith('.ts')))
        sources[path.relative(tree,file)]=fs.readFileSync(file,'utf8');
    for(const file of walk(path.join(tree,'built')).filter(f=>/\.(js|mjs|cjs)$/.test(f)))
        js[path.relative(path.join(tree,'built'),file)]=hash(fs.readFileSync(file));
    assert(Object.keys(js).length>0);
    return {sources,js,publicAPI:hash(fs.readFileSync(path.join(tree,'built/local/typescript.d.ts')))};
}
fs.mkdirSync(proof,{recursive:true});
const file=path.join(proof,'state-before.json');
if(mode==='before') {
    const p=program(), checker=p.getTypeChecker(), callers=[];
    for(const source of p.getSourceFiles().filter(f=>f.fileName.startsWith(path.join(tree,'src/compiler')))) {
        function visit(n) {
            if(ts.isCallExpression(n)) {
                let symbol=checker.getSymbolAtLocation(n.expression);
                if(symbol?.flags & ts.SymbolFlags.Alias)symbol=checker.getAliasedSymbol(symbol);
                if(symbol?.declarations?.some(d=>ts.isFunctionDeclaration(d)&&d.name?.text==='isArray'&&d.getSourceFile().fileName===path.join(tree,'src/compiler/core.ts'))) {
                    const pos=source.getLineAndCharacterOfPosition(n.getStart(source));
                    callers.push({file:path.relative(tree,source.fileName),line:pos.line+1,expression:n.getText(source)});
                }
            }
            ts.forEachChild(n,visit);
        }
        visit(source);
    }
    assert(callers.length>0);
    const state={...snapshot(),diagnostics:diagnostics(p),callers};
    fs.writeFileSync(file,JSON.stringify(state));
    fs.writeFileSync(path.join(proof,'callers.json'),JSON.stringify(callers,null,2)+'\n');
    console.log(JSON.stringify({callerSites:callers.length,diagnostics:state.diagnostics.length,jsFiles:Object.keys(state.js).length}));
} else {
    const before=JSON.parse(fs.readFileSync(file)),after=snapshot();
    assert.deepEqual(after.js,before.js,'every emitted JavaScript file and byte');
    assert.equal(after.publicAPI,before.publicAPI,'internal predicate has no public API delta');
    assert.deepEqual(Object.keys(after.sources),Object.keys(before.sources));
    for(const [name,text] of Object.entries(before.sources)) {
        const expected=name==='src/compiler/core.ts' ? text.replace('export function isArray(value: any)','export function isArray(value: unknown)') : text;
        assert.equal(after.sources[name],expected,'unchanged caller source: '+name);
    }
    const actual=diagnostics(program());
    // Positions after the longer annotation shift; compare diagnostic identities.
    const identity=d=>JSON.stringify({file:d.file,code:d.code,message:d.message});
    assert.deepEqual(actual.map(identity),before.diagnostics.map(identity),'all compiler callers type-check unchanged');
    const mutant=diagnostics(program('string'));
    const errors=mutant.filter(d=>d.code===2345&&!actual.some(a=>identity(a)===identity(d)));
    assert(errors.length>0,'string input mutant must reject a real compiler caller');
    const input=before.sources['src/compiler/core.ts'],output=after.sources['src/compiler/core.ts'];
    assert.equal(ts.transpileModule(input,{compilerOptions:{target:ts.ScriptTarget.ESNext}}).outputText,
        ts.transpileModule(output,{compilerOptions:{target:ts.ScriptTarget.ESNext}}).outputText);
    const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'isarray-body-mutant-'));
    fs.mkdirSync(path.join(scratch,'src/compiler'),{recursive:true});
    fs.writeFileSync(path.join(scratch,'src/compiler/core.ts'),output.replace('return Array.isArray(value);','return false;'));
    const bodyMutant=cp.spawnSync(process.execPath,[path.join(__dirname,'adapt.cjs'),scratch],{encoding:'utf8'});
    assert.notEqual(bodyMutant.status,0);
    assert(bodyMutant.stderr.includes('unreviewed uses: isArray'),'adapter rejects changed predicate body');
    fs.writeFileSync(path.join(proof,'body-mutant.log'),bodyMutant.stderr);
    const report={site:58,class:'Generic array predicate input',replacement:'unknown',callerSites:before.callers.length,
        beforeDiagnostics:before.diagnostics.length,afterDiagnostics:actual.length,newDiagnostics:0,
        jsFiles:Object.keys(after.js).length,publicAPIIdentical:true,bodyMutantCaught:true,stringMutantErrors:errors};
    fs.writeFileSync(path.join(proof,'proof.json'),JSON.stringify(report,null,2)+'\n');
    console.log(JSON.stringify(report,null,2));
}
