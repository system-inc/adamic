'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const {census, nodes, ts} = require('./census.cjs');
const bodies = require('./bodies.cjs');
const runtime = require('./runtime.cjs');
function adapt(tree) {
    const before = census(tree);
    if (!before.generatorCount) {
        for (const name of ['core.ts','checker.ts']) assert(fs.readFileSync(path.join(tree,'src/compiler',name),'utf8').includes(runtime.replaceAll('\n','\r\n')), 'missing iterator fallback');
        return {files:0, generators:0};
    }
    assert.equal(before.generatorCount,13,'generator census drift');
    assert.equal(before.delegationCount,2,'delegation census drift');
    const plans=[];
    for (const name of ['core.ts','checker.ts']) {
        const file=path.join(tree,'src/compiler',name), original=fs.readFileSync(file,'utf8');
        const source=ts.createSourceFile(file,original,99,true);
        const generators=nodes(source,n=>ts.isFunctionLike(n)&&!!n.asteriskToken);
        assert.equal(generators.length,name==='core.ts'?7:6);
        const edits=[];
        for(const node of generators) {
            const key=node.name?.getText(source)||'<anonymous>';
            let body=bodies[key];
            assert(body,'unreviewed generator '+key);
            const expected=require('./reviewed.json').find(row=>row.file===name&&row.name===key);
            assert(expected,'missing reviewed body');
            // Compare token streams, retaining literals and every operator; earlier adapters may reflow whitespace.
            function tokens(text) { const scanner=ts.createScanner(99,true,ts.LanguageVariant.Standard,text); const result=[]; for(let kind=scanner.scan();kind!==ts.SyntaxKind.EndOfFileToken;kind=scanner.scan())result.push([kind,scanner.getTokenText()]);return result; }
            assert.deepEqual(tokens(node.body.getText(source)),tokens(expected.body),'reviewed generator drift: '+key);
            edits.push({start:node.asteriskToken.getStart(source),end:node.asteriskToken.end,text:''});
            edits.push({start:node.body.getStart(source),end:node.body.end,text:'{'+body.replaceAll('\r\n','\n').replaceAll('\n','\r\n')+'}'});
        }
        let after=original;
        for(const edit of edits.sort((a,b)=>b.start-a.start))after=after.slice(0,edit.start)+edit.text+after.slice(edit.end);
        // Module-local helpers introduce no namespace exports or public declarations.
        after+='\r\n'+runtime.replaceAll('\n','\r\n');
        if(name==='core.ts') after+='\r\n'+bodies.flatMapRuntime.replaceAll('\n','\r\n');
        const parsed=ts.createSourceFile(file,after,99,true);
        assert.equal(parsed.parseDiagnostics.length,0,'replacement parse');
        assert.equal(nodes(parsed,n=>ts.isYieldExpression(n)||ts.isFunctionLike(n)&&!!n.asteriskToken).length,0);
        plans.push({file,original,after});
    }
    for(const plan of plans)assert.equal(fs.readFileSync(plan.file,'utf8'),plan.original);
    for(const plan of plans)fs.writeFileSync(plan.file,plan.after);
    const after=census(tree);assert.equal(after.generatorCount,0);assert.equal(after.delegationCount,0);
    return {files:plans.length,generators:before.generatorCount,delegations:before.delegationCount};
}
if(require.main===module)console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
module.exports={adapt};
