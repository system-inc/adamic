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
                else if (node.type.kind !== ts.SyntaxKind.UndefinedKeyword) throw Error('unreviewed brand type: ' + key);
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    if (found.size !== selected.length) throw Error('missing reviewed brands: ' + file);
    for (const e of edits.sort((a,b) => b.start-a.start)) text = text.slice(0,e.start)+e.value+text.slice(e.end);
    return {text, removed: edits.length, contracts};
}
function apply(tree) {
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
            else if(n.type.kind!==ts.SyntaxKind.UndefinedKeyword)throw Error('unreviewed public marker type');
        }
        ts.forEachChild(n,visitReference);
    }
    visitReference(source);
    if(referenceMarkers!==27)throw Error('public brand census changed');
    let referenceText=beforeReference;
    for(const e of referenceEdits.sort((a,b)=>b.start-a.start))referenceText=referenceText.slice(0,e.start)+'undefined'+referenceText.slice(e.end);
    for (const p of plans) if (fs.readFileSync(p.name,'utf8') !== p.before) throw Error('concurrent source change');
    for (const p of plans) if (p.before !== p.text) fs.writeFileSync(p.name,p.text);
    if(referenceText!==beforeReference)fs.writeFileSync(reference,referenceText);
    console.log(JSON.stringify({class:'brands', removed: plans.reduce((n,p) => n+p.removed,0)}));
}
module.exports = {plan,apply,owner};

function planEnum(text,file) {
    const selected=(rules['enum-display']||[]).filter(r=>r.file===file);
    const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true), lines=text.split('\n');
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
function applyEnums(tree) {
    const plans=[...new Set((rules['enum-display']||[]).map(r=>r.file))].map(file=>{const name=path.join(tree,file),before=fs.readFileSync(name,'utf8');return{name,before,...planEnum(before,file)};});
    for(const p of plans)if(fs.readFileSync(p.name,'utf8')!==p.before)throw Error('concurrent enum source change');
    for(const p of plans)if(p.before!==p.text)fs.writeFileSync(p.name,p.text);
    console.log(JSON.stringify({class:'enum-display',removed:plans.reduce((n,p)=>n+p.removed,0)}));
}
module.exports.planEnum=planEnum;
module.exports.applyEnums=applyEnums;
