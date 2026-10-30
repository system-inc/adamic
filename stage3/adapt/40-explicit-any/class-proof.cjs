'use strict';
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const assert = require('node:assert/strict'), ts = require('typescript');
assert.equal(ts.version,'6.0.3');
const [mode, tree, evidence, className] = process.argv.slice(2);
const rules = require('./class-rules.json')[className];
assert(rules && tree && evidence,'usage: mode tree proof class');
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const walk = d => fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(d,e.name)):[path.join(d,e.name)]).sort();
function enclosing(node) {
    for(let p=node.parent;p;p=p.parent) if ((ts.isInterfaceDeclaration(p)||ts.isTypeAliasDeclaration(p)||ts.isClassDeclaration(p)||ts.isFunctionDeclaration(p) || ts.isVariableDeclaration(p))&&p.name) return p.name.text;
}
function declarationProjection(text, contracts) {
    const src=ts.createSourceFile('api.d.ts',text,ts.ScriptTarget.Latest,true), edits=[];let aliasNeeded=false;
    function visit(n) {
        if (((ts.isPropertyDeclaration(n)||ts.isPropertySignature(n))&&className==='brands'||ts.isParameter(n)&&className==='enum-display')&&n.type?.kind===ts.SyntaxKind.AnyKeyword&&contracts.some(c=>c.key===n.name.text&&c.owner===enclosing(n))) edits.push([n.type.getStart(src),n.type.end,className==='brands'?'undefined':'Record<string, string | number>']);
        if(className==='diagnostic'&&ts.isParameter(n)&&contracts.some(c=>c.key===n.name.text&&c.owner===enclosing(n))) {
            if(n.type?.kind===ts.SyntaxKind.AnyKeyword)edits.push([n.type.getStart(src),n.type.end,'string | number']);
            else if(ts.isArrayTypeNode(n.type)&&n.type.elementType.kind===ts.SyntaxKind.AnyKeyword){edits.push([n.type.elementType.getStart(src),n.type.elementType.end,'DiagnosticArguments[number]']);aliasNeeded=true;}
        }
        if(className==='diagnostic'&&ts.isTypeAliasDeclaration(n)&&n.name.text==='DiagnosticArguments')edits.push([n.type.getStart(src),n.type.end,'(string | number | boolean | readonly string[] | SourceFile | undefined)[]']);
        ts.forEachChild(n,visit);
    }
    visit(src);
    if(aliasNeeded)for(const n of src.statements) {
        if(ts.isImportDeclaration(n)&&n.moduleSpecifier.text==='./_namespaces/ts.js'&&ts.isNamedImports(n.importClause?.namedBindings)) {
            const elements=n.importClause.namedBindings.elements;
            if(!elements.some(e=>e.name.text==='DiagnosticArguments')) {
                const message=elements.find(e=>e.name.text==='DiagnosticMessage');assert(message);
                edits.push([message.getStart(src),message.getStart(src),'type DiagnosticArguments, ']);
            }
        }
    }
    for(const [a,b,value] of edits.sort((a,b)=>b[0]-a[0]))text=text.slice(0,a)+value+text.slice(b);
    return text;
}
function expectedSource(text,file) {
    if(className==='diagnostic')for(const r of require('./diagnostic-declarations.json').filter(r=>r.file===file)){assert.equal(text.split(r.before).length,2,'consumer owner reconstruction');text=text.replace(r.before,r.after);}
    const lines=text.split('\n');
    for(const r of rules.filter(r=>r.file===file)) {
        if(className!=='brands') {
            assert.equal(lines[r.line-1].replace(/\r$/,''),r.originalLine,'independent enum owner reconstruction');
            const at=lines[r.line-1].indexOf('any',r.column-1);assert.equal(at,r.column-1);
            lines[r.line-1]=lines[r.line-1].slice(0,at)+r.replacement+lines[r.line-1].slice(at+3);continue;
        }
        const key=r.key.replace(/[.*+?^${}()|[\]\\]/g,'\\$&');
        const expression=new RegExp('('+ (r.key.startsWith(' ')?'"'+key+'"':key) +'\\??:\\s*)any\\b','g');
        let count=0;
        lines[r.line-1]=lines[r.line-1].replace(expression,(_,prefix)=>{count++;return prefix+'undefined';});
        assert.equal(count,1,'independent owner reconstruction: '+r.id);
    }
    return lines.join('\n');
}
function equalJS(before,after) { assert.deepEqual(after,before,'entire emitted JavaScript file set and bytes'); }
function snapshot() {
    const texts={}, sources={}, census=[];
    for(const file of walk(path.join(tree,'src/compiler')).filter(f=>f.endsWith('.ts')&&!f.includes('.generated.'))) {
        const text=fs.readFileSync(file,'utf8'), relative=path.relative(tree,file);texts[relative]=text;sources[relative]=hash(text);
        const src=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
        function visit(n){if(n.kind===ts.SyntaxKind.AnyKeyword){const p=src.getLineAndCharacterOfPosition(n.getStart(src));census.push({file:relative,line:p.line+1,column:p.character+1,source:text.split(/\r?\n/)[p.line]});}ts.forEachChild(n,visit);}visit(src);
    }
    const js={}, declarations={};
    for(const file of walk(path.join(tree,'built'))) {const name=path.relative(path.join(tree,'built'),file);if(/\.(js|mjs|cjs)$/.test(file))js[name]=hash(fs.readFileSync(file));if(/\.d\.ts$/.test(file))declarations[name]=fs.readFileSync(file,'utf8');}
    const configPath=path.join(tree,'src/compiler/tsconfig.json');
    const read=ts.readConfigFile(configPath,ts.sys.readFile);assert(!read.error);
    const config=ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configPath),undefined,configPath);assert.equal(config.errors.length,0);
    const diagnostics=ts.getPreEmitDiagnostics(ts.createProgram(config.fileNames,config.options)).map(d=>{const p=d.file?.getLineAndCharacterOfPosition(d.start);return{code:d.code,file:d.file?path.relative(tree,d.file.fileName):null,line:p?p.line+1:null,column:p?p.character+1:null,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};});
    return {texts,sources,census,js,declarations,diagnostics};
}
fs.mkdirSync(evidence,{recursive:true});
const statePath=path.join(evidence,'state-before.json');
const api=path.join(tree,'tests/baselines/reference/api/typescript.d.ts');
if(mode==='before') {
    if(className==='brands') {
    const keys=new Set(rules.map(r=>r.key));
    function audit(text,file) {
        const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);let declarations=0;
        function visit(n) {
            if((ts.isIdentifier(n)||ts.isStringLiteral(n))&&keys.has(n.text)) {
                const p=n.parent;
                assert((ts.isPropertySignature(p)||ts.isPropertyDeclaration(p))&&p.name===n,'runtime marker reference: '+file+':'+n.text);
                if(ts.isPropertyDeclaration(p))assert(p.modifiers?.some(m=>m.kind===ts.SyntaxKind.DeclareKeyword)&&!p.initializer,'emitted marker field');
                declarations++;
            }
            ts.forEachChild(n,visit);
        }
        visit(source);return declarations;
    }
    let markerDeclarations=0,filesAudited=0;
    for(const f of walk(path.join(tree,'src')).filter(f=>f.endsWith('.ts'))) {markerDeclarations+=audit(fs.readFileSync(f,'utf8'),f);filesAudited++;}
    assert.throws(()=>audit('obj.__pathBrand;','real-reference-mutant.ts'),/runtime marker reference/);
    fs.writeFileSync(path.join(evidence,'marker-audit.json'),JSON.stringify({filesAudited,markerDeclarations,runtimeReferences:0,mutant:'added marker property read caught'},null,2)+'\n');
    }
    const state=snapshot();
    state.reference=fs.readFileSync(api,'utf8');state.contracts=[];
    if(className==='enum-display')state.contracts=[{owner:'formatEnum',key:'enumObject'}];
    else if(className==='diagnostic')state.contracts=rules.map(r=>({owner:r.owner,key:r.key}));
    else for(const r of rules){const src=ts.createSourceFile(r.file,state.texts[r.file],ts.ScriptTarget.Latest,true);function visit(n){if((ts.isPropertyDeclaration(n)||ts.isPropertySignature(n))&&n.name.text===r.key&&src.getLineAndCharacterOfPosition(n.getStart(src)).line+1===r.line)state.contracts.push({key:r.key,owner:enclosing(n)});ts.forEachChild(n,visit);}visit(src);}
    assert.equal(state.contracts.length,className==='enum-display'?1:rules.length);
    fs.writeFileSync(statePath,JSON.stringify(state));
    const {texts,declarations,reference,contracts,...summary}=state;
    fs.writeFileSync(path.join(evidence,'before.json'),JSON.stringify({...summary,contracts},null,2)+'\n');
    console.log(JSON.stringify({mode,tokens:state.census.length,diagnostics:state.diagnostics.length,jsFiles:Object.keys(state.js).length}));
} else if(mode==='api') {
    const before=JSON.parse(fs.readFileSync(statePath));
    const expected=declarationProjection(before.reference,before.contracts);
    assert([before.reference,expected].includes(fs.readFileSync(api,'utf8')),'public API reference differs only by reviewed edits');
    const actual=fs.readFileSync(path.join(tree,'built/local/typescript.d.ts'),'utf8').replace(/\r\n/g,'\n');
    assert.equal(actual,expected,'public API differs only by reviewed brand declarations');
    fs.writeFileSync(api,expected);
    console.log(JSON.stringify({mode,referenceEdits:before.reference===expected?0:ts.createSourceFile('api',before.reference,ts.ScriptTarget.Latest,true).text.split(': any').length-expected.split(': any').length}));
} else if(mode==='after') {
    const before=JSON.parse(fs.readFileSync(statePath)), after=snapshot();
    equalJS(before.js,after.js);
    assert.deepEqual(Object.keys(after.texts),Object.keys(before.texts));
    for(const file of Object.keys(before.texts))assert.equal(after.texts[file],expectedSource(before.texts[file],file),'only reviewed declaration tokens: '+file);
    assert.equal(before.census.length-after.census.length,rules.length,'exact class census reduction');
    assert.deepEqual(Object.keys(after.declarations),Object.keys(before.declarations));
    const changedDeclarations=[];
    for(const file of Object.keys(before.declarations)) {const expected=declarationProjection(before.declarations[file],before.contracts);assert.equal(after.declarations[file],expected,'mechanically explained declarations: '+file);if(before.declarations[file]!==expected)changedDeclarations.push(file);}
    assert.equal(fs.readFileSync(api,'utf8'),declarationProjection(before.reference,before.contracts),'exact public API exception');
    const consumerErrors=after.diagnostics.filter(d=>!before.diagnostics.some(p=>JSON.stringify(p)===JSON.stringify(d)));
    fs.writeFileSync(path.join(evidence,'consumer-errors.json'),JSON.stringify(consumerErrors,null,2)+'\n');
    const {texts,declarations,...summary}=after;fs.writeFileSync(path.join(evidence,'after.json'),JSON.stringify(summary,null,2)+'\n');
    assert.equal(consumerErrors.length,0,'new consumer diagnostics');
    const {spawnSync}=require('node:child_process');
    const second=spawnSync(process.execPath,[path.join(__dirname,'adapt.cjs'),tree],{encoding:'utf8'});assert.equal(second.status,0,second.stderr);
    for(const file of Object.keys(after.texts))assert.equal(fs.readFileSync(path.join(tree,file),'utf8'),after.texts[file],'adapter idempotence');
    const mutants=[];
    function killed(name,fn){assert.throws(fn);mutants.push(name);}
    const jsMutant=structuredClone(after.js);jsMutant[Object.keys(jsMutant)[0]]=hash('changed output');killed('JavaScript bytes',()=>equalJS(before.js,jsMutant));
    const missing=structuredClone(after.js);delete missing[Object.keys(missing)[0]];killed('JavaScript file set',()=>equalJS(before.js,missing));
    const first=rules[0];killed('unreviewed source byte',()=>assert.equal(after.texts[first.file]+'\n',expectedSource(before.texts[first.file],first.file)));
    killed('census dropped site',()=>assert.equal(before.census.length-(after.census.length-1),rules.length));
    killed('unrelated API edit',()=>assert.equal(declarationProjection(before.reference,before.contracts)+'\ntype Unreviewed = string;\n',fs.readFileSync(api,'utf8')));
    killed('idempotence changed byte',()=>assert.equal(after.texts[first.file]+'\n',fs.readFileSync(path.join(tree,first.file),'utf8')));
    if(className==='brands'){const bad=before.texts[first.file].replace('__incrementalBuildInfoFileIdBrand: any','__incrementalBuildInfoFileIdBrand: string');killed('unreviewed owner type',()=>require('./classes.cjs').plan(bad,first.file));}
    else {const invalid=className==='enum-display'?first.originalLine.replace('enumObject: any','enumObject: string'):first.originalLine.replace('args: any[]','args: boolean[]');const bad=before.texts[first.file].replace(first.originalLine,invalid);killed('unreviewed class owner type',()=>require('./classes.cjs').planEnum(bad,first.file,className));}
    const configPath=path.join(tree,'src/compiler/tsconfig.json');
    const config=ts.parseJsonConfigFileContent(ts.readConfigFile(configPath,ts.sys.readFile).config,ts.sys,path.dirname(configPath),undefined,configPath);
    const host=ts.createCompilerHost(config.options), originalRead=host.readFile;
    host.readFile=file=>{
        const text=originalRead(file);
        if(className==='brands'&&file===path.join(tree,'src/compiler/checker.ts'))return text.replace('declare _symbolLinksBrand: undefined;','declare _symbolLinksBrand: string;');
        if(className==='enum-display'&&file===path.join(tree,'src/compiler/debug.ts'))return text.replace('enumObject: Record<string, string | number>','enumObject: Record<string, boolean>');
        if(className==='diagnostic'&&file===path.join(tree,'src/compiler/commandLineParser.ts'))return text.replace('...args: DiagnosticArguments[number][]','...args: boolean[]');
        return text;
    };
    const mutantCode=className==='brands'?2322:2345;
    assert(ts.getPreEmitDiagnostics(ts.createProgram(config.fileNames,config.options,host)).some(d=>d.code===mutantCode),'checker catches conflicting class contract');
    mutants.push(className==='brands'?'real brand implementation mismatch caught by stock checker TS2322':className==='enum-display'?'real enum map domain mutation caught by stock checker TS2345':'real diagnostic argument domain mutation caught by stock checker TS2345');
    const proof={class:className,before:before.census.length,after:after.census.length,removed:rules.length,jsFiles:Object.keys(after.js).length,declarationFiles:Object.keys(after.declarations).length,changedDeclarations,consumerErrors,idempotent:true,mutants};
    fs.writeFileSync(path.join(evidence,'proof.json'),JSON.stringify(proof,null,2)+'\n');console.log(JSON.stringify(proof,null,2));
} else throw Error('unknown proof mode');
