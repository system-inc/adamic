'use strict';
const crypto=require('node:crypto');
const e={fs:require('node:fs'),path:require('node:path'),assert:require('node:assert/strict'),ts:require('typescript'),digest:s=>crypto.createHash('sha256').update(s).digest('hex')};
const {fs,path,assert,ts} = e;
const targetKeys=new Set(require('./maybebind-sites.json').map(s=>JSON.stringify([s.owner,s.after.match(/\.([A-Za-z0-9_]+)\?\.bind/)[1]])));
const wanted=s=>targetKeys.has(JSON.stringify([s.owner,s.member]));
function unwrap(n) { while (ts.isParenthesizedExpression(n)) n=n.expression; return n; }
function inspect(file,text,side) {
    const options={allowJs:true,noLib:true,noResolve:true,noEmit:true,target:ts.ScriptTarget.ESNext};
    const host=ts.createCompilerHost(options);
    host.getSourceFile=(name,version)=>name===file?ts.createSourceFile(name,text,version,true,ts.ScriptKind.JS):undefined;
    host.fileExists=name=>name===file;host.readFile=name=>name===file?text:undefined;
    const p=ts.createProgram([file],options,host),source=p.getSourceFile(file),checker=p.getTypeChecker();
    assert.equal(source.parseDiagnostics.length,0);
    const symbol=n=>{
        if(!ts.isIdentifier(n)||!(/^_[a-z]\d*$/.test(n.text)||/^maybeBind\d*$/.test(n.text)))return undefined;
        const parent=n.parent;
        if((ts.isPropertyAccessExpression(parent)&&parent.name===n)||((ts.isPropertyAssignment(parent)||ts.isPropertySignature(parent))&&parent.name===n))return undefined;
        return checker.getSymbolAtLocation(n);
    };
    const canonical=new Map(), introduced=new Set(), binders=new Set(), coreHelpers=new Set(), sites=[];
    function key(n,member,receiver,form) {
        const owners=[];
        for(let a=n.parent;a;a=a.parent)if(ts.isFunctionLike(a)&&a.name&&ts.isIdentifier(a.name))owners.unshift(a.name.text);
        return {owner:owners.join('.'),receiver:receiver.text,member:member.name.text,form};
    }
    function discover(n) {
        if(ts.isFunctionDeclaration(n)&&n.name&&/^maybeBind\d*$/.test(n.name.text)) {
            if(n.parameters.length===1)binders.add(symbol(n.name));
            if(n.parameters.length===2)coreHelpers.add(n);
        }
        if(ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&/^maybeBind\d*$/.test(n.expression.text)&&n.arguments.length===2) {
            const [receiver,member]=n.arguments;
            if(ts.isIdentifier(receiver)&&ts.isPropertyAccessExpression(member)&&ts.isIdentifier(member.expression)&&receiver.text===member.expression.text) {
                const site=key(n,member,receiver,'original');
                if(wanted(site)){canonical.set(n,{receiver,member});sites.push(site);}
            }
        }
        if(ts.isConditionalExpression(n)) {
            const condition=unwrap(n.condition),yes=unwrap(n.whenTrue),no=unwrap(n.whenFalse);
            if(ts.isBinaryExpression(condition)&&condition.operatorToken.kind===ts.SyntaxKind.EqualsEqualsToken&&condition.right.kind===ts.SyntaxKind.NullKeyword&&ts.isVoidExpression(yes)&&ts.isCallExpression(no)&&ts.isPropertyAccessExpression(no.expression)&&no.expression.name.text==='bind'&&no.arguments.length===1) {
                const assign=unwrap(condition.left),receiver=no.arguments[0];
                if(ts.isBinaryExpression(assign)&&assign.operatorToken.kind===ts.SyntaxKind.EqualsToken&&ts.isIdentifier(assign.left)&&ts.isPropertyAccessExpression(assign.right)&&ts.isIdentifier(receiver)&&ts.isIdentifier(assign.right.expression)&&receiver.text===assign.right.expression.text&&ts.isIdentifier(no.expression.expression)&&symbol(assign.left)===symbol(no.expression.expression)) {
                    const site=key(n,assign.right,receiver,'bound');
                    if(wanted(site)){canonical.set(n,{receiver,member:assign.right});introduced.add(symbol(assign.left));sites.push(site);}
                }
            }
        }
        ts.forEachChild(n,discover);
    }
    discover(source);
    const tempNames=new Map(), scopeCounts=new Map(), bindings=[];
    function temps(n,scope) {
        if(ts.isFunctionLike(n))scope=n;
        if(ts.isVariableDeclaration(n)&&ts.isIdentifier(n.name)&&/^_[a-z]\d*$/.test(n.name.text)) {
            const s=symbol(n.name);
            if(!introduced.has(s)) {
                const count=scopeCounts.get(scope)||0;scopeCounts.set(scope,count+1);tempNames.set(s,'__32_temp_'+count);bindings.push({symbol:s,canonical:'__32_temp_'+count,name:n.name.text});
            }
        }
        ts.forEachChild(n,child=>temps(child,scope));
    }
    temps(source,source);
    for(const s of binders) {tempNames.set(s,'__32_binder');bindings.push({symbol:s,canonical:'__32_binder',name:s.name});}
    const audits=new Map(bindings.map(b=>[b.symbol,{name:b.name,canonical:b.canonical,occurrences:0,mentions:[]}]));
    const spellings=new Map();
    for(const a of audits.values()){if(!spellings.has(a.name))spellings.set(a.name,[]);spellings.get(a.name).push(a);}
    function audit(n) {
        if(ts.isIdentifier(n)&&spellings.has(n.text)) {
            const a=audits.get(symbol(n));
            if(a) {
                a.occurrences++;
                if(ts.isPropertyAccessExpression(n.parent)&&n.parent.expression===n&&n.parent.name.text==='name') a.mentions.push(n.parent.getText(source));
                if(ts.isElementAccessExpression(n.parent)&&n.parent.expression===n) a.mentions.push(n.parent.getText(source));
            }
        }
        if(ts.isStringLiteralLike(n))for(const [name,rows] of spellings)if(n.text.includes(name))for(const a of rows)a.mentions.push(n.getText(source));
        const property=ts.isPropertyAccessExpression(n)?n.name:((ts.isPropertyAssignment(n)||ts.isMethodDeclaration(n))?n.name:undefined);
        if(property)for(const a of spellings.get(property.getText(source))||[])a.mentions.push(n.getText(source));
        ts.forEachChild(n,audit);
    }
    audit(source);
    const renameAudit=[...audits.values()];
    const transformed=ts.transform(source,[context=>{
        function visit(n) {
            // All handed-off original calls remain, so both bundles retain the core helper.
            if(ts.isParenthesizedExpression(n)&&canonical.has(unwrap(n)))return visit(unwrap(n));
            const bound=canonical.get(n);
            if(bound) {
                // A remaining original after call is intentionally retained for the negative witness.
                if(side==='after'&&ts.isCallExpression(n))return ts.visitEachChild(n,visit,context);
                return ts.factory.createCallExpression(ts.factory.createIdentifier('__32_bound'),undefined,[ts.visitNode(bound.receiver,visit),ts.visitNode(bound.member,visit)]);
            }
            if(ts.isVariableStatement(n)) {
                const kept=n.declarationList.declarations.filter(d=>!ts.isIdentifier(d.name)||!introduced.has(symbol(d.name)));
                if(!kept.length)return undefined;
                return ts.visitEachChild(ts.factory.updateVariableStatement(n,n.modifiers,ts.factory.updateVariableDeclarationList(n.declarationList,kept)),visit,context);
            }
            if(ts.isIdentifier(n)) {
                const s=symbol(n),name=tempNames.get(s);
                if(name)return ts.factory.createIdentifier(name);
                // A renamed binder is mapped by its checker symbol and audited above.
            }
            return ts.visitEachChild(n,visit,context);
        }
        return root=>ts.visitNode(root,visit);
    }]);
    const normalized=ts.createPrinter({removeComments:true,newLine:ts.NewLineKind.LineFeed}).printFile(transformed.transformed[0]);
    transformed.dispose();
    return {normalized,sites,introducedTemps:introduced.size,helperDeclarations:coreHelpers.size,renameAudit};
}
function compare(beforeTree,afterTree,out) {
    fs.mkdirSync(out,{recursive:true});
    const b=path.join(beforeTree,'built/local'),a=path.join(afterTree,'built/local');
    const names=fs.readdirSync(b).filter(n=>n.endsWith('.js')).sort();assert.equal(names.length,9);
    assert.deepEqual(fs.readdirSync(a).filter(n=>n.endsWith('.js')).sort(),names);
    const rows=[],unexpected=new Map();
    for(const name of names) {
        const before=fs.readFileSync(path.join(b,name),'utf8'),after=fs.readFileSync(path.join(a,name),'utf8');
        if(before===after){rows.push({file:name,byteIdentical:true,pass:true});continue;}
        const bi=inspect(path.join(b,name),before,'before'),ai=inspect(path.join(a,name),after,'after');
        const original=ai.sites.filter(s=>s.form==='original');
        for(const site of original) {
            const pin=require('./maybebind-sites.json').find(r=>r.owner===site.owner&&r.receiver===site.receiver&&r.member===site.member);
            const source=fs.readFileSync(path.join(beforeTree,'src/compiler/watch.ts'),'utf8');
            const at=source.indexOf(pin.before);assert(at>=0&&source.indexOf(pin.before,at+1)===-1,'source site attribution');
            const prefix=source.slice(0,at);
            unexpected.set(JSON.stringify([site.owner,site.receiver,site.member]),{file:'watch.ts',line:prefix.split('\n').length,column:at-prefix.lastIndexOf('\n'),...site});
        }
        const strip=sites=>sites.map(({owner,receiver,member})=>JSON.stringify([owner,receiver,member])).sort();
        const inventoryEqual=JSON.stringify(strip(bi.sites))===JSON.stringify(strip(ai.sites));
        const same=bi.normalized===ai.normalized;
        assert.equal(bi.renameAudit.length,ai.renameAudit.length,'remaining binding declaration count');
        const renames=bi.renameAudit.map((b,i)=>{
            const a=ai.renameAudit[i];
            assert.equal(b.canonical,a.canonical,'binding correspondence');
            assert.equal(b.occurrences,a.occurrences,'binding occurrence count: '+b.name);
            if(b.name!==a.name)assert(!b.mentions.length&&!a.mentions.length,'observable renamed spelling');
            return {before:b.name,after:a.name,canonical:b.canonical,occurrencesBefore:b.occurrences,occurrencesAfter:a.occurrences,mentionsBefore:b.mentions,mentionsAfter:a.mentions};
        });
        if(!same) {
            fs.writeFileSync(path.join(out,name+'.before-normalized.txt'),bi.normalized);
            fs.writeFileSync(path.join(out,name+'.after-normalized.txt'),ai.normalized);
        }
        rows.push({file:name,byteIdentical:false,renames:renames.filter(r=>r.before!==r.after),expectedSites:bi.sites.length,boundSites:ai.sites.filter(s=>s.form==='bound').length,remainingOriginal:original,generatedTemps:ai.introducedTemps,inventoryEqual,normalizedEqual:same,pass:inventoryEqual&&same&&!original.length&&bi.sites.length===12,beforeSha256:e.digest(before),afterSha256:e.digest(after)});
    }
    const result={pass:rows.every(r=>r.pass),files:rows,unexpectedSites:[...unexpected.values()]};
    fs.writeFileSync(path.join(out,'comparison.json'),JSON.stringify(result,null,2)+'\n');
    return result;
}
module.exports={compare,inspect};
if(require.main===module) {
    assert.equal(process.argv.length,5,'usage: node compare.cjs BEFORE AFTER OUTPUT');
    const result=compare(...process.argv.slice(2).map(p=>path.resolve(p)));
    console.log(JSON.stringify(result,null,2));process.exitCode=result.pass?0:1;
}
