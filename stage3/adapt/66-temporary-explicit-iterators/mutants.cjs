'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),cp=require('node:child_process'),assert=require('node:assert/strict');
const {nodes,ts}=require('./census.cjs');
const [before,after,output]=process.argv.slice(2);
assert(before&&after&&output,'usage: mutants.cjs BEFORE AFTER OUTPUT-DIRECTORY');
fs.mkdirSync(output,{recursive:true});
const reports=[];
for(const name of ['remaining-generator','skipped-element','eager-iterator']) {
    const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'adapt66-mutant-'));
    try {
        fs.cpSync(path.join(after,'src/compiler'),path.join(scratch,'src/compiler'),{recursive:true});
        const file=path.join(scratch,'src/compiler/core.ts');let text=fs.readFileSync(file,'utf8');
        const parsed=ts.createSourceFile(file,text,99,true);
        const functionName=name==='eager-iterator'?'arrayReverseIterator':'singleIterator';
        const found=nodes(parsed,n=>ts.isFunctionDeclaration(n)&&n.name?.text===functionName);assert.equal(found.length,1);
        const node=found[0];let replacement=node.getText(parsed);
        if(name==='remaining-generator') {
            const original=ts.createSourceFile(file,fs.readFileSync(path.join(before,'src/compiler/core.ts'),'utf8'),99,true);
            replacement=nodes(original,n=>ts.isFunctionDeclaration(n)&&n.name?.text===functionName)[0].getText(original);
        } else if(name==='skipped-element') {
            assert(replacement.includes('let pending = true;'));replacement=replacement.replace('let pending = true;','let pending = false;');
        } else {
            assert(replacement.includes('let index: number | undefined;'));replacement=replacement.replace('let index: number | undefined;','let index: number | undefined = array.length - 1;');
        }
        assert.notEqual(replacement,node.getText(parsed));
        text=text.slice(0,node.getStart(parsed))+replacement+text.slice(node.end);fs.writeFileSync(file,text);
        const args=name==='remaining-generator'?[path.join(__dirname,'census.cjs'),scratch,'--check']:[path.join(__dirname,'verify.cjs'),before,scratch,path.join(output,name+'.json'),functionName];
        const start=Date.now();const result=cp.spawnSync(process.execPath,args,{encoding:'utf8',timeout:60000});
        fs.writeFileSync(path.join(output,name+'.log'),result.stdout+result.stderr);
        assert.equal(result.status,1,name+' survived or timed out');
        const catcher=name==='remaining-generator'?'census failed: generators or yield delegation remain':name==='eager-iterator'?'laziness: array length read before next':'singleIterator changed values or evaluation timing';
        assert((result.stdout+result.stderr).includes(catcher),name+' failed for the wrong reason');
        reports.push({name,caught:true,exit:result.status,catcher,seconds:(Date.now()-start)/1000});
    } finally {fs.rmSync(scratch,{recursive:true,force:true});}
}
fs.writeFileSync(path.join(output,'report.json'),JSON.stringify({status:'pass',mutants:reports},null,2)+'\n');console.log(JSON.stringify(reports));
