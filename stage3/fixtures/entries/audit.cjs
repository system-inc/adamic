// Stock compiler audit. Virtual .ts names exist only in the compiler host, never on disk.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
assert.equal(ts.version,'6.0.3');
const here=__dirname,upstream=path.resolve(process.argv[2]),output=process.argv[3];
const provenance=JSON.parse(fs.readFileSync(path.join(here,'scanner-source.json'),'utf8'));
for(const pin of provenance.files)assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(upstream,pin.file))).digest('hex'),pin.sha256);
const rows=[];
for(const contract of JSON.parse(fs.readFileSync(path.join(here,'expectations.json'),'utf8'))) {
    const name=path.join(here,contract.file),source=fs.readFileSync(name,'utf8'),virtual=name+'.ts';
    const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,lib:['lib.es2024.d.ts'],types:[]};
    const host=ts.createCompilerHost(options),load=host.getSourceFile.bind(host);
    host.getSourceFile=(file,language,onError,createNew)=>file===virtual?ts.createSourceFile(virtual,'declare const console: { log(message: string): void };\n'+source+'\nexport {};',language,true):load(file,language,onError,createNew);
    const program=ts.createProgram([virtual],options,host),checker=program.getTypeChecker(),tree=program.getSourceFile(virtual);
    const diagnostics=ts.getPreEmitDiagnostics(program);
    assert.equal(diagnostics.length,0,ts.formatDiagnostics(diagnostics,{getCanonicalFileName:x=>x,getCurrentDirectory:()=>here,getNewLine:()=> '\n'}));
    const facts={file:contract.file,diagnostics:0,sha256:crypto.createHash('sha256').update(source).digest('hex'),entries:[]};
    function type(n){return checker.typeToString(checker.getTypeAtLocation(n),undefined,ts.TypeFormatFlags.NoTruncation);}
    function visit(n){
        if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&n.expression.expression.getText(tree)==='Object'&&n.expression.name.text==='entries')facts.entries.push({argument_type:type(n.arguments[0]),stock_result_type:type(n),contextual_type:checker.getContextualType(n)?checker.typeToString(checker.getContextualType(n)):null});
        if(ts.isVariableDeclaration(n)&&n.name.getText(tree)==='source') {
            assert.ok(ts.isObjectLiteralExpression(n.initializer));
            facts.allocation_properties=n.initializer.properties.map(p=>({name:p.name.getText(tree),kind:ts.SyntaxKind[p.kind],value_type:ts.isPropertyAssignment(p)?type(p.initializer):null}));
        }
        if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken&&ts.isElementAccessExpression(n.left))facts.alias_write={receiver:n.left.expression.getText(tree),key:n.left.argumentExpression.getText(tree),value_type:type(n.right)};
        ts.forEachChild(n,visit);
    }
    visit(tree);assert.equal(facts.entries.length,1);
    rows.push(facts);
}
const expected=['proven','proven','checked','checked','refused','refused'];
const contracts=JSON.parse(fs.readFileSync(path.join(here,'expectations.json'),'utf8'));
assert.deepEqual(contracts.map(x=>x.expected.reflection),expected);
assert.equal(rows[1].entries[0].argument_type,'Visible');
assert.ok(rows[1].allocation_properties.some(p=>p.name==='hidden'&&p.value_type==='2'));
assert.ok(rows[2].allocation_properties.some(p=>p.name==='hidden'&&p.value_type==='"wrong"'));
assert.equal(rows[3].alias_write.receiver,'alias');
assert.equal(rows[3].alias_write.value_type,'2');
assert.ok(rows[4].allocation_properties.some(p=>p.kind==='GetAccessor'));
assert.ok(rows[5].allocation_properties.some(p=>p.name.includes('Symbol')));
const keywordMap=new Map(Object.entries(ts.textToKeywordObj));
const golden='count:'+keywordMap.size+'\n'+[...keywordMap].map(([key,value])=>key+':'+value+'\n').join('');
assert.equal(golden,JSON.parse(fs.readFileSync(path.join(here,'status.json'),'utf8'))[0].node.stdout);
const report={typescript:ts.version,stock_diagnostics:0,fixtures:rows,scanner_keywords:keywordMap.size,scanner_stdout_sha256:crypto.createHash('sha256').update(golden).digest('hex')};
if(output)fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');
console.log('PASS pinned source hashes, six stock checker programs, structural views, hidden value types, alias write, getter, symbol key, and full stock scanner golden');
