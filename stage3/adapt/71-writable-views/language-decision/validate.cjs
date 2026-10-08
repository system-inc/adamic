'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),vm=require('node:vm'),child=require('node:child_process');
const {ts,options,createAudit}=require('./audit.cjs');
const mutation=process.argv[2];
const cases={
    receivers:[['read','a'],['writeSame','b'],['writeOther','c'],['store','d'],['forward','a']],
    arrays:[['read','a'],['same','b'],['other','c'],['conditional','d']],
    nested:[['read','a'],['same','b'],['other','c']],
    escapes:[['stored','d'],['captured','d'],['unknown','d'],['generic','c']],
};
const outputs={receivers:'1 1 2 2 2',arrays:'1 1,1,2,2',nested:'1 2'};
const reports=[];
for(const [name,expected] of Object.entries(cases)){
    let source=fs.readFileSync(path.join(__dirname,'controls',name+'.a'),'utf8');
    if(name==='receivers'){
        if(mutation==='reader-writes')source=source.replace('const alias = view; return alias.value;', 'const alias = view; alias.value = 2; return alias.value;');
        if(mutation==='compatible-writes-other')source=source.replace('alias.value = 1;', 'alias.value = 2;');
        if(mutation==='unsound-writes-same')source=source.replace('alias.value = 2;', 'alias.value = 1;');
        if(mutation==='shorthand-no-escape')source=source.replace('return { view };','return { value: view.value };');
    }
    const root='/unit71-audit-controls',file=root+'/'+name+'.ts';
    const settings={...options,noEmit:false,allowImportingTsExtensions:false,moduleDetection:ts.ModuleDetectionKind.Legacy};
    const host=ts.createCompilerHost(settings),get=host.getSourceFile;
    host.getSourceFile=(f,v,e,n)=>f===file?ts.createSourceFile(f,source,v,true,ts.ScriptKind.TS):get(f,v,e,n);
    let javascript;host.writeFile=(f,s)=>{if(f.endsWith(name+'.js'))javascript=s;};
    const program=ts.createProgram([file],settings,host),diagnostics=ts.getPreEmitDiagnostics(program);
    assert.equal(diagnostics.length,0,diagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n')).join('\n'));
    const audit=createAudit(program,root),sf=program.getSourceFile(file);
    const calls=[];function visit(n){if(ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&n.arguments[0]?.getText(sf)==='holder')calls.push(n);ts.forEachChild(n,visit);}visit(sf);
    for(const [callee,classification] of expected){
        const selected=calls.filter(c=>c.expression.text===callee);assert.equal(selected.length,1);
        const node=selected[0].arguments[0],result=audit.inspect(node,audit.checker.getTypeAtLocation(node));
        assert.equal(result.classification,classification,`${name}/${callee}: ${JSON.stringify(result)}`);
        if(classification==='a')assert.equal(result.unresolved.length+result.writes.length,0);
        if(classification==='b')assert(result.writes.length && result.writes.every(w=>w.compatible) && !result.unresolved.length);
        if(classification==='c')assert(result.writes.some(w=>w.definitely_outside && !w.conditional));
        if(classification==='d')assert(result.unresolved.length);
        reports.push({control:name,callee,classification,reads:result.reads.length,writes:result.writes,unresolved:result.unresolved});
    }
    if(name==='escapes'){
        let generic;function find(n){if(ts.isAsExpression(n)&&n.expression.getText(sf)==='view')generic=n.expression;ts.forEachChild(n,find);}find(sf);
        assert.equal(audit.inspect(generic,audit.checker.getTypeAtLocation(generic)).classification,'d');
    }
    if(outputs[name]){assert.equal(program.emit().emitSkipped,false);const lines=[];vm.runInNewContext(javascript,{console:{log:v=>lines.push(v)}});assert.deepEqual(lines,[outputs[name]]);}
}
if(!mutation){
    const mutants=[];
    for(const name of ['reader-writes','compatible-writes-other','unsound-writes-same','shorthand-no-escape']){
        const result=child.spawnSync(process.execPath,[__filename,name],{encoding:'utf8'});
        assert.equal(result.status,1,name);assert(result.stderr.includes('AssertionError'),result.stderr);assert(result.stderr.includes('receivers/'),result.stderr);
        mutants.push({name,exit:result.status,caught_by:'receiver classification assertion; stock TypeScript diagnostics remain zero'});
    }
    console.log(JSON.stringify({typescript:ts.version,controls:reports,external_node_outputs:outputs,mutants},null,2));
}
