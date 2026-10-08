'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require(process.env.CENSUS_TYPESCRIPT || 'typescript');
assert.equal(ts.version,'6.0.3');
assert.equal(process.argv.length,3,'usage: adapt.cjs <tree>');
const tree=path.resolve(process.argv[2]),file=path.join(tree,'src/compiler/debug.ts');
const original='type AssertionKeys = MatchingKeys<typeof Debug, AnyFunction>;';
const keys=['assertEachNode','assertNode','assertNotNode','assertOptionalNode','assertOptionalToken','assertMissingNode'];
const truthful='type AssertionKeys = '+keys.map(k=>JSON.stringify(k)).join(' | ')+';';
function plan(text,core) {
    const source=ts.createSourceFile('debug.ts',text,ts.ScriptTarget.Latest,true);
    const owners=source.statements.filter(n=>ts.isModuleDeclaration(n)&&n.name.text==='Debug');
    assert.equal(owners.length,1,'Debug owner drift');
    const calls=[],cache=[],writers=[],helpers=[];
    function visit(n) {
        if(ts.isFunctionDeclaration(n)&&n.name?.text==='shouldAssertFunction') helpers.push(n);
        if(ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&n.expression.text==='shouldAssertFunction') {
            assert.equal(n.arguments.length,2,'assertion caller drift');
            assert(ts.isStringLiteral(n.arguments[1]),'nonliteral assertion key');
            calls.push(n.arguments[1].text);
        }
        if(ts.isIdentifier(n)&&n.text==='assertionCache') {
            cache.push(n);
            const p=n.parent;
            if(ts.isVariableDeclaration(p)&&p.name===n) assert.equal(p.initializer?.getText(source),'{}','cache initializer drift');
            else if(ts.isCallExpression(p)) assert.equal(p.expression.getText(source),'getOwnKeys','cache escape');
            else {
                assert(ts.isElementAccessExpression(p)&&p.expression===n,'cache escape');
                const outer=p.parent;
                if(ts.isBinaryExpression(outer)&&outer.left===p) {
                    assert.equal(outer.operatorToken.kind,ts.SyntaxKind.EqualsToken,'cache update drift');
                    writers.push(outer.getText(source));
                }
            }
        }
        ts.forEachChild(n,visit);
    }
    visit(owners[0]);
    assert.deepEqual(calls.slice().sort(),keys.slice().sort(),'finite assertion key set changed');
    assert.equal(helpers.length,1,'assertion helper drift');
    assert.equal(cache.length,5,'cache ownership/reference drift');
    assert.deepEqual(writers.slice().sort(),['assertionCache[key] = undefined','assertionCache[name] = { level, assertion: Debug[name] }'].sort(),'cache write drift');
    assert.equal(helpers[0].parameters[1].name.getText(source),'name');
    assert.equal(helpers[0].typeParameters?.[0].constraint.getText(source),'AssertionKeys');
    const c=ts.createSourceFile('core.ts',core,ts.ScriptTarget.Latest,true);
    const enumerators=c.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='getOwnKeys');
    assert.equal(enumerators.length,1,'key enumerator drift');
    assert.equal(enumerators[0].body.getText(c).replace(/\r\n/g,'\n'),'{\n    const keys: string[] = [];\n    for (const key in map) {\n        if (hasOwnProperty.call(map, key)) {\n            keys.push(key);\n        }\n    }\n\n    return keys;\n}','key enumeration changed');
    assert.equal(text.split(original).length-1+text.split(truthful).length-1,1,'assertion type drift');
    const changed=text.includes(original)?text.replace(original,truthful):text;
    for(const removeComments of [false,true]) {
        const compilerOptions={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments};
        assert.equal(ts.transpileModule(changed,{compilerOptions}).outputText,ts.transpileModule(text,{compilerOptions}).outputText,'runtime JavaScript changed');
    }
    return {changed,keys,already_adapted:changed===text};
}
const before=fs.readFileSync(file,'utf8'),core=fs.readFileSync(path.join(tree,'src/compiler/core.ts'),'utf8');
const result=plan(before,core);
if(result.changed!==before)fs.writeFileSync(file,result.changed);
console.log(JSON.stringify({file:'src/compiler/debug.ts',keys:result.keys,already_adapted:result.already_adapted,javascript_identical:true,sameMap:'pending: honest union exposes builder.ts:564 TS2322; no compensating cast applied'}));
module.exports={plan};
