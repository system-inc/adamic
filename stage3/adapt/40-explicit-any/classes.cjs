'use strict';
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const rules = require('./class-rules.json');
function owner(node) {
    for (let p = node.parent; p; p = p.parent) {
        if ((ts.isInterfaceDeclaration(p) || ts.isTypeAliasDeclaration(p) || ts.isClassDeclaration(p) || ts.isFunctionDeclaration(p) || ts.isVariableDeclaration(p)) && p.name) return p.name.text;
    }
    throw Error('brand has no named owner');
}
function plan(text, file) {
    const selected = rules.brands.filter(r => r.file === file);
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const found = new Set(), edits = [], contracts = [];
    function visit(node) {
        if ((ts.isPropertySignature(node) || ts.isPropertyDeclaration(node)) && node.name && node.type) {
            const key = node.name.text;
            const line = source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1;
            const rule = selected.find(r => r.key === key && r.line === line);
            if (rule) {
                if (found.has(rule.id) || node.initializer) throw Error('unreviewed brand initialization');
                found.add(rule.id);
                contracts.push({key, owner: owner(node)});
                if (node.type.kind === ts.SyntaxKind.AnyKeyword) edits.push({start: node.type.getStart(source), end: node.type.end, value: rule.replacement});
                else if (node.type.kind !== ts.SyntaxKind.VoidKeyword &&
                    !(key === ' __sortedArrayBrand' && node.type.kind === ts.SyntaxKind.UndefinedKeyword)) throw Error('unreviewed brand type: ' + key);
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    if (found.size !== selected.length) throw Error('missing reviewed brands: ' + file);
    for (const e of edits.sort((a,b) => b.start-a.start)) text = text.slice(0,e.start)+e.value+text.slice(e.end);
    return {text, removed: edits.length, contracts};
}
// Phantom markers may occur only in erased declarations, never as values.
function auditBrands(text, file) {
    const keys = new Set(rules.brands.map(r => r.key));
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const declarations = [];
    function visit(n) {
        if ((ts.isIdentifier(n) || ts.isStringLiteral(n)) && keys.has(n.text)) {
            const p = n.parent;
            if (!(ts.isPropertySignature(p) || ts.isPropertyDeclaration(p)) || p.name !== n)
                throw Error('runtime marker reference: ' + file + ':' + n.text);
            if (ts.isPropertyDeclaration(p) &&
                (!p.modifiers?.some(m => m.kind === ts.SyntaxKind.DeclareKeyword) || p.initializer))
                throw Error('emitted marker field: ' + file + ':' + n.text);
            declarations.push({key: n.text, owner: owner(n),
                line: source.getLineAndCharacterOfPosition(n.getStart(source)).line + 1});
        }
        ts.forEachChild(n, visit);
    }
    visit(source);
    return declarations;
}
function auditTree(tree) {
    function walk(directory) {
        for (const entry of fs.readdirSync(directory, {withFileTypes: true})) {
            const file = path.join(directory, entry.name);
            if (entry.isDirectory()) walk(file);
            else if (file.endsWith('.ts')) auditBrands(fs.readFileSync(file, 'utf8'), file);
        }
    }
    walk(path.join(tree, 'src/compiler'));
}
function apply(tree) {
    auditTree(tree);
    const plans = [...new Set(rules.brands.map(r => r.file))].map(file => {
        const name = path.join(tree,file), before = fs.readFileSync(name,'utf8');
        return {name,before,...plan(before,file)};
    });
    // The public API golden may differ only by these reviewed marker types.
    const reference = path.join(tree,'tests/baselines/reference/api/typescript.d.ts');
    const beforeReference = fs.readFileSync(reference,'utf8');
    const source = ts.createSourceFile(reference,beforeReference,ts.ScriptTarget.Latest,true);
    const contracts = plans.flatMap(p=>p.contracts), referenceEdits=[];
    let referenceMarkers=0;
    function visitReference(n) {
        if((ts.isPropertySignature(n)||ts.isPropertyDeclaration(n))&&n.type&&contracts.some(c=>c.key===n.name.text&&c.owner===owner(n))) {
            referenceMarkers++;
            if(n.type.kind===ts.SyntaxKind.AnyKeyword)referenceEdits.push({start:n.type.getStart(source),end:n.type.end});
            else if(n.type.kind!==ts.SyntaxKind.VoidKeyword &&
                !(n.name.text===' __sortedArrayBrand' && n.type.kind===ts.SyntaxKind.UndefinedKeyword))throw Error('unreviewed public marker type');
        }
        ts.forEachChild(n,visitReference);
    }
    visitReference(source);
    if(referenceMarkers!==27)throw Error('public brand census changed');
    let referenceText=beforeReference;
    for(const e of referenceEdits.sort((a,b)=>b.start-a.start))referenceText=referenceText.slice(0,e.start)+'void'+referenceText.slice(e.end);
    for (const p of plans) if (fs.readFileSync(p.name,'utf8') !== p.before) throw Error('concurrent source change');
    for (const p of plans) if (p.before !== p.text) fs.writeFileSync(p.name,p.text);
    if(referenceText!==beforeReference)fs.writeFileSync(reference,referenceText);
    console.log(JSON.stringify({class:'brands', removed: plans.reduce((n,p) => n+p.removed,0)}));
}
module.exports = {plan,apply,owner,auditBrands};

function planEnum(text,file,className='enum-display') {
    const selected=(rules[className]||[]).filter(r=>r.file===file);
    const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true), lines=text.split('\n');
    if (className === 'filesystem') {
        const edits = []; let found = 0;
        function visit(node) {
            if (ts.isVariableDeclaration(node) && node.name.getText(source) === 'stat') {
                let enclosing = node.parent;
                while (enclosing && !ts.isFunctionDeclaration(enclosing)) enclosing = enclosing.parent;
                if (enclosing?.name?.text === 'getAccessibleFileSystemEntries') {
                    const rule = selected[0];
                    if (!rule || node.initializer || !node.type) throw Error('unreviewed filesystem owner');
                    found++;
                    const line = source.getLineAndCharacterOfPosition(node.getStart(source)).line;
                    const original = rule.originalLine;
                    const adapted = original.replace('any', rule.replacement);
                    const current = lines[line].replace(/\r$/, '');
                    if (current === original && node.type.kind === ts.SyntaxKind.AnyKeyword) edits.push([node.type.getStart(source), node.type.end, rule.replacement]);
                    else if (current !== adapted || node.type.getText(source) !== rule.replacement) throw Error('unreviewed filesystem annotation');
                }
            }
            ts.forEachChild(node, visit);
        }
        visit(source);
        if (found !== selected.length) throw Error('missing or duplicate filesystem stat owner');
        for (const [start, end, value] of edits.sort((a,b) => b[0]-a[0])) text = text.slice(0,start)+value+text.slice(end);
        return {text, removed: edits.length};
    }
    let removed=0;
    for(const r of selected) {
        const before=r.originalLine, after=before.slice(0,r.column-1)+r.replacement+before.slice(r.column+2);
        const current=lines[r.line-1]?.replace(/\r$/,'');
        if(current===after)continue;
        if(current!==before)throw Error('unreviewed enum owner/use: '+r.id);
        let matched=false;
        function visit(n) {
            if(n.kind===ts.SyntaxKind.AnyKeyword) {
                const p=source.getLineAndCharacterOfPosition(n.getStart(source));
                if(p.line+1===r.line&&p.character+1===r.column) {
                    if(ts.SyntaxKind[n.parent.kind]!==r.parentKind)throw Error('enum edit is not its reviewed type position');
                    matched=true;
                }
            }
            ts.forEachChild(n,visit);
        }
        visit(source);if(!matched)throw Error('missing reviewed enum any');
        lines[r.line-1]=after+(lines[r.line-1].endsWith('\r')?'\r':'');removed++;
    }
    return {text:lines.join('\n'),removed};
}
function applyEnums(tree,className='enum-display') {
    const plans=[...new Set((rules[className]||[]).map(r=>r.file))].map(file=>{const name=path.join(tree,file),before=fs.readFileSync(name,'utf8');return{name,before,...planEnum(before,file,className)};});
    for(const p of plans)if(fs.readFileSync(p.name,'utf8')!==p.before)throw Error('concurrent enum source change');
    for(const p of plans)if(p.before!==p.text)fs.writeFileSync(p.name,p.text);
    console.log(JSON.stringify({class:className,removed:plans.reduce((n,p)=>n+p.removed,0)}));
}
module.exports.planEnum=planEnum;
module.exports.applyEnums=applyEnums;

function applyDiagnosticReference(tree) {
    const file=path.join(tree,'tests/baselines/reference/api/typescript.d.ts'),text=fs.readFileSync(file,'utf8');
    const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),edits=[];let matched=0;
    function visit(n) {
        if(ts.isParameter(n)&&n.name.text==='arg0'&&owner(n)==='ErrorCallback') {
            matched++;
            if(n.type.kind===ts.SyntaxKind.AnyKeyword)edits.push([n.type.getStart(source),n.type.end]);
            else if(n.type.getText(source)!=='string | number')throw Error('unreviewed public diagnostic type');
        }
        ts.forEachChild(n,visit);
    }
    visit(source);if(matched!==1)throw Error('public diagnostic callback count changed');
    let after=text;for(const [a,b]of edits.sort((a,b)=>b[0]-a[0]))after=after.slice(0,a)+'string | number'+after.slice(b);
    if(after!==text)fs.writeFileSync(file,after);
}
module.exports.applyDiagnosticReference=applyDiagnosticReference;

function applyDiagnosticDeclarations(tree) {
    const extras=require('./diagnostic-declarations.json');
    const texts=new Map();
    for(const r of extras) {
        const file=path.join(tree,r.file);let text=texts.get(file)??fs.readFileSync(file,'utf8');
        if(!text.includes(r.after)) {
            if(text.split(r.before).length!==2)throw Error('diagnostic declaration drift: '+r.before);
            text=text.replace(r.before,r.after);texts.set(file,text);
        }
    }
    for(const [file,text]of texts)fs.writeFileSync(file,text);
}
module.exports.applyDiagnosticDeclarations=applyDiagnosticDeclarations;
