'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('typescript');
const assert=require('node:assert/strict'),inspector=require('node:inspector'),vm=require('node:vm');
const tree=process.argv[2];if(!tree)throw Error('usage: debugger-proof.cjs TREE');
const name=path.join(tree,'src/compiler/debug.ts'),source=ts.createSourceFile(name,fs.readFileSync(name,'utf8'),ts.ScriptTarget.Latest,true);
let declaration;
function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='fail')declaration=n;ts.forEachChild(n,visit);}visit(source);
assert(declaration,'actual Debug.fail owner');
const original=declaration.getText(source).replace(/^export /,''),mutant=original.replace('debugger;',';');
assert.notEqual(mutant,original,'one real debugger statement removed');
const session=new inspector.Session();session.connect();let paused=0;
session.on('Debugger.paused',()=>{paused++;session.post('Debugger.resume');});session.post('Debugger.enable');
for(const [name,text,expected] of [['original',original,1],['deleted-debugger-mutant',mutant,0]]){
    const before=paused;
    const script='(function(){'+text+'\ntry { fail("probe"); } catch (error) { if(error.message !== "Debug Failure. probe") throw error; } })();';
    vm.runInThisContext(ts.transpileModule(script,{compilerOptions:{target:ts.ScriptTarget.ESNext}}).outputText,{filename:name+'.js'});
    assert.equal(paused-before,expected,'actual Debug.fail '+name+' inspector pauses');
}
session.post('Debugger.disable');session.disconnect();
console.log(JSON.stringify({owner:'src/compiler/debug.ts:fail',originalPauses:1,removedPauses:0,errorMessageUnchanged:true,mutant:'deleting actual Debug.fail debugger loses its observable breakpoint',decision:'retain until behavior policy is decided'}));
