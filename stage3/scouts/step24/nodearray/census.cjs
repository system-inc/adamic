// Stock checker census of array layout and source sites, pinned to TypeScript 6.0.3.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const root = path.resolve(process.argv[2]);
const out = path.resolve(process.argv[3] || __dirname);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath, ts.sys.readFile);
assert.equal(read.error, undefined);
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
assert.equal(config.errors.length, 0);
const program = ts.createProgram(config.fileNames, config.options);
const checker = program.getTypeChecker();
assert.equal(ts.getPreEmitDiagnostics(program).length, 0);
console.log('stock diagnostics: 0');
const files = program.getSourceFiles().filter(f => f.fileName.startsWith(path.join(root, 'src/compiler') + '/') && !f.fileName.includes('.generated.')).sort((a,b)=>a.fileName.localeCompare(b.fileName));
function location(n) {
    const f = n.getSourceFile(), pos = f.getLineAndCharacterOfPosition(n.getStart(f));
    return { file: path.relative(root, f.fileName), line: pos.line + 1, column: pos.character + 1, start: n.getStart(f), end: n.end };
}
const textType = t => checker.typeToString(t, undefined, ts.TypeFormatFlags.NoTruncation);
function walk(n, visit) { visit(n); ts.forEachChild(n, child => walk(child, visit)); }
const declared = [];
for (const f of files) walk(f, n => { if (ts.isInterfaceDeclaration(n) && ['NodeArray','MutableNodeArray'].includes(n.name.text)) declared.push(n); });
assert.equal(declared.length, 2);
const arrayTypes = new Map();
function isNodeArray(t, seen = new Set()) {
    if (!t || seen.has(t)) return false;
    seen.add(t);
    if (['NodeArray','MutableNodeArray'].includes(t.symbol?.name) || ['NodeArray','MutableNodeArray'].includes(t.target?.symbol?.name)) return true;
    if ((t.objectFlags & ts.ObjectFlags.Mapped) && t.aliasTypeArguments?.some(x=>isNodeArray(x, seen))) return true;
    if (t.isUnionOrIntersection() && t.types.some(x=>isNodeArray(x, seen))) return true;
    const constraint = checker.getBaseConstraintOfType(t);
    if (constraint && constraint !== t && isNodeArray(constraint, seen)) return true;
    if (t.flags & ts.TypeFlags.Object) {
        const target = t.target || t;
        if (target.objectFlags & (ts.ObjectFlags.Class | ts.ObjectFlags.Interface)) {
            for (const base of checker.getBaseTypes(target) || []) if (isNodeArray(base, seen)) return true;
        }
    }
    return false;
}
function nodeArray(t) { if (!arrayTypes.has(t)) arrayTypes.set(t, isNodeArray(t)); return arrayTypes.get(t); }
function arrayLike(t) {
    if (nodeArray(t) || checker.isArrayType(t) || checker.isTupleType(t)) return true;
    return t.isUnionOrIntersection() ? t.types.some(arrayLike) : false;
}
function symbols(s) { return s ? [s, ...checker.getRootSymbols(s)] : []; }
function declarations(s) { return [...new Set(symbols(s).flatMap(s=>s.declarations || []))].map(location); }
const layouts = declared.map(n => {
    const type = checker.getTypeAtLocation(n);
    return { ...location(n), name: n.name.text, fields: checker.getPropertiesOfType(type).filter(p => p.declarations?.some(d=>!d.getSourceFile().isDeclarationFile)).map(p => ({ name:p.name, type:textType(checker.getTypeOfSymbolAtLocation(p,n)), optional:!!(p.flags & ts.SymbolFlags.Optional), declarations:declarations(p) })) };
});
const fields = new Set(layouts.flatMap(l=>l.fields.map(f=>f.name)));
const standardArrayFields = new Set(declared.flatMap(n=>checker.getPropertiesOfType(checker.getTypeAtLocation(n)).map(p=>p.name)).filter(name=>!fields.has(name)));
const reflection=[];
assert.deepEqual([...fields].sort(), ['end','hasTrailingComma','pos','transformFlags']);
const accesses=[], creations=[], candidates=[], dynamic=[], rangeCalls=[], rangeHelpers=[];
const rangeNames=new Set(["setTextRange","setTextRangePosEnd","setTextRangePos","setTextRangeEnd","setTextRangePosWidth"]);
function unwrap(n) { while (ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)||ts.isParenthesizedExpression(n)||ts.isNonNullExpression(n)) n=n.expression; return n; }
function accessMode(n) {
    let p=n.parent;
    if(ts.isBinaryExpression(p)&&p.left===n&&p.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&p.operatorToken.kind<=ts.SyntaxKind.LastAssignment) return p.operatorToken.kind===ts.SyntaxKind.EqualsToken?'write':'read/write';
    if((ts.isPrefixUnaryExpression(p)||ts.isPostfixUnaryExpression(p))&&[ts.SyntaxKind.PlusPlusToken,ts.SyntaxKind.MinusMinusToken].includes(p.operator)) return 'read/write';
    if(ts.isDeleteExpression(p)) return 'delete';
    return 'read';
}
function names(n) {
    if(ts.isPropertyAccessExpression(n)) return [n.name.text];
    const t=checker.getTypeAtLocation(n.argumentExpression);
    return (t.isUnion()?t.types:[t]).filter(t=>t.flags & (ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral)).map(t=>String(t.value));
}
function sourceSymbol(n) {
    let s=checker.getSymbolAtLocation(n);
    if(s?.flags&ts.SymbolFlags.Alias)s=checker.getAliasedSymbol(s);
    return s;
}
function calleeName(n) {
    const sig=checker.getResolvedSignature(n), d=sig?.declaration;
    return d?.name?.getText() || sourceSymbol(ts.isPropertyAccessExpression(n.expression)?n.expression.name:n.expression)?.name;
}
for(const f of files) walk(f, n=>{
    // Shared TextRange helpers are not typed NodeArray, but can read/write its layout.
    if(ts.isPropertyAccessExpression(n) && ['pos','end'].includes(n.name.text)) {
        let owner=n.parent;
        while(owner&&!ts.isFunctionDeclaration(owner))owner=owner.parent;
        if(owner && rangeNames.has(owner.name?.text)) {
            const receiverType=checker.getTypeAtLocation(n.expression);
            rangeHelpers.push({...location(n),function:owner.name.text,field:n.name.text,mode:accessMode(n),expression:n.getText(),receiver_type:textType(receiverType),declarations:declarations(checker.getPropertyOfType(receiverType,n.name.text))});
        }
    }
    if((ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n)) && nodeArray(checker.getTypeAtLocation(n.expression))) {
        const ns=names(n),rt=checker.getTypeAtLocation(n.expression);
        if(!ns.length && ts.isElementAccessExpression(n)) dynamic.push({...location(n),expression:n.getText(),mode:accessMode(n),receiver_type:textType(rt),key_type:textType(checker.getTypeAtLocation(n.argumentExpression))});
        for(const name of ns) {
            const symbol=checker.getPropertyOfType(rt,name);
            if(standardArrayFields.has(name))continue;
            if(/^\d+$/.test(name))continue;
            const mode=accessMode(n),p=n.parent;
            accesses.push({...location(n),field:name,mode,expression:n.getText(),receiver:n.expression.getText(),receiver_type:textType(rt),declarations:declarations(symbol),outside_declared_layout:!fields.has(name),value_type:mode.includes('write')&&ts.isBinaryExpression(p)?textType(checker.getTypeAtLocation(p.right)):undefined});
        }
    }
    if(ts.isCallExpression(n)) {
        const name=calleeName(n),rt=checker.getTypeAtLocation(n),arg=n.arguments[0];
        if(rangeNames.has(name)) {
            const arrayArguments=n.arguments.map((a,index)=>({a,index,type:checker.getTypeAtLocation(a)})).filter(x=>arrayLike(x.type));
            if(arrayArguments.length) rangeCalls.push({...location(n),callee:name,expression:n.getText(),signature_declaration:checker.getResolvedSignature(n)?.declaration?location(checker.getResolvedSignature(n).declaration):null,arrays:arrayArguments.map(x=>({argument:x.index+1,expression:x.a.getText(),type:textType(x.type),role:x.index===0?'range target: writes pos/end as selected by helper':'range provider: reads pos/end'}))});
        }
        if(name==='hasProperty' && arg && arrayLike(checker.getTypeAtLocation(arg)) && n.arguments[1]) {
            const key=checker.getTypeAtLocation(n.arguments[1]);
            if(key.flags & ts.TypeFlags.StringLiteral && fields.has(key.value))reflection.push({...location(n),field:key.value,mode:'own-presence query',expression:n.getText(),receiver:arg.getText()});
        }
        if(name==='defineProperties' && arg && nodeArray(checker.getTypeAtLocation(arg)) && n.arguments[1] && ts.isObjectLiteralExpression(n.arguments[1])) {
            for(const p of n.arguments[1].properties)if(p.name)reflection.push({...location(p),field:p.name.getText(),mode:'descriptor write',expression:p.getText(),receiver:arg.getText(),outside_declared_layout:!fields.has(p.name.getText())});
        }
        let category;
        if(name==='createNodeArray' && nodeArray(rt))category='createNodeArray call (may reuse)';
        else if(['setTextRange','setTextRangePosEnd','setTextRangePos','setTextRangeEnd','setTextRangePosWidth'].includes(name) && arg && arrayLike(checker.getTypeAtLocation(arg))) category='setTextRange on array (range initialization/update)';
        else if(ts.isPropertyAccessExpression(n.expression)&&name==='slice'&&nodeArray(checker.getTypeAtLocation(n.expression.expression)))category='NodeArray.slice (ordinary array result)';
        else if(ts.isPropertyAccessExpression(n.expression)&&name==='slice') {
            for(let p=n.parent;p&&(ts.isParenthesizedExpression(p)||ts.isAsExpression(p)||ts.isConditionalExpression(p));p=p.parent)if((ts.isAsExpression(p)||ts.isConditionalExpression(p))&&nodeArray(checker.getTypeAtLocation(p))){category='slice promoted to NodeArray';break;}
        }
        if(category)creations.push({...location(n),category,callee:name,expression:n.getText(),result_type:textType(rt),signature_declaration:checker.getResolvedSignature(n)?.declaration?location(checker.getResolvedSignature(n).declaration):null});
        else if(nodeArray(rt))candidates.push({...location(n),category:'other NodeArray-returning call (not necessarily an allocation)',callee:name,expression:n.getText(),result_type:textType(rt)});
    }
    if(ts.isBinaryExpression(n) && n.operatorToken.kind===ts.SyntaxKind.InKeyword && nodeArray(checker.getTypeAtLocation(n.right))) {
        const key=checker.getTypeAtLocation(n.left);
        if(key.flags & ts.TypeFlags.StringLiteral)reflection.push({...location(n),field:key.value,mode:'presence query (includes prototype)',expression:n.getText(),receiver:n.right.getText(),outside_declared_layout:!fields.has(key.value)});
    }
    if((ts.isAsExpression(n)||ts.isTypeAssertionExpression(n))&&nodeArray(checker.getTypeAtLocation(n))&&!nodeArray(checker.getTypeAtLocation(n.expression)))creations.push({...location(n),category:'array/other value promoted by assertion',expression:n.getText(),source_type:textType(checker.getTypeAtLocation(n.expression)),result_type:textType(checker.getTypeAtLocation(n))});
    if(ts.isArrayLiteralExpression(n) && (nodeArray(checker.getTypeAtLocation(n)) || (checker.getContextualType(n)&&nodeArray(checker.getContextualType(n)))))creations.push({...location(n),category:'NodeArray-context array literal',expression:n.getText(),result_type:textType(checker.getTypeAtLocation(n))});
});
const counts=rows=>rows.reduce((c,r)=>(c[r]=(c[r]||0)+1,c),{});
const result={typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',stock_diagnostics:0,files:files.map(f=>({file:path.relative(root,f.fileName),sha256:crypto.createHash('sha256').update(fs.readFileSync(f.fileName)).digest('hex')})),layouts,accesses,reflection,range_helpers:rangeHelpers,range_calls:rangeCalls,creations,other_returning_calls:candidates,dynamic_accesses:dynamic,summary:{range_calls:rangeCalls.length,range_array_arguments:counts(rangeCalls.flatMap(x=>x.arrays.map(a=>a.role))),range_helper_accesses:rangeHelpers.length,accesses:accesses.length,reflection:reflection.length,access_modes:counts(accesses.map(x=>x.mode)),access_fields:counts(accesses.map(x=>x.field)),creation_forms:counts(creations.map(x=>x.category)),other_returning_calls:candidates.length,dynamic_accesses:dynamic.length}};
fs.mkdirSync(out,{recursive:true});
fs.writeFileSync(path.join(out,'census.json'),JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify(result.summary,null,2));
