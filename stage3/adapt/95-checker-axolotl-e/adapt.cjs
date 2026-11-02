#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require(process.env.CENSUS_TYPESCRIPT||'typescript');
const {compiler,nodes}=require('../47-host-errors/adapt.cjs');
const inventory=require('./sites.json');
const hash=text=>crypto.createHash('sha256').update(text).digest('hex');
function adapt(tree,check=false){
    assert.equal(ts.version,'6.0.3');
    const program=compiler(tree,ts),checker=program.getTypeChecker();
    const file=program.getSourceFile(path.join(tree,'src/compiler/types.ts'));
    assert(file&&!file.parseDiagnostics.length,'types must parse');
    const owners=nodes(ts,file,n=>ts.isInterfaceDeclaration(n)&&n.name.text==='TypeChecker');
    assert.equal(owners.length,1,'TypeChecker owner');
    const members=owners[0].members.filter(n=>ts.isMethodSignature(n)&&n.name.getText(file)==='signatureToSignatureDeclaration');
    assert.equal(members.length,2,'both reviewed overloads');
    const symbol=checker.getSymbolAtLocation(members[0].name);
    assert(symbol&&symbol.declarations.length===2,'overload symbol');
    assert.equal(checker.getTypeOfSymbolAtLocation(symbol,members[0]).getCallSignatures().length,2);
    const edits=[];
    for(const [index,member] of members.entries()){
        const row=inventory[index];
        assert.equal(checker.getSymbolAtLocation(member.name),symbol,'shared overload symbol');
        assert.equal(ts.getJSDocTags(member).some(t=>t.tagName.text==='internal'),index===1,'public/internal overload order');
        const normalized=member.getText(file).replace(row.after,row.before);
        assert.equal(hash(normalized),row.ownerBeforeSha256,'overload owner shape');
        const slots=nodes(ts,member.type,n=>ts.isPropertySignature(n)&&n.name.getText(file)==='typeArguments');
        assert.equal(slots.length,1,'intersection slot count');
        const slot=slots[0];
        assert(slot.questionToken&&ts.isTypeLiteralNode(slot.parent)&&ts.isIntersectionTypeNode(slot.parent.parent),'optional intersection slot');
        const text=slot.type.getText(file);
        assert([row.before,row.after].includes(text),'slot type shape');
        const reference=ts.isUnionTypeNode(slot.type)?slot.type.types[0]:slot.type;
        assert(ts.isTypeReferenceNode(reference)&&reference.typeName.getText(file)==='NodeArray');
        assert.equal(reference.typeArguments.length,1);
        assert.equal(reference.typeArguments[0].getText(file),'TypeNode');
        assert.equal(checker.getSymbolAtLocation(reference.typeName)?.name,'NodeArray','bound NodeArray type');
        assert.equal(checker.typeToString(checker.getTypeFromTypeNode(slot.type)),text===row.before?row.typeBefore:row.typeAfter,'slot type');
        assert(checker.getSymbolAtLocation(slot.name)?.declarations.includes(slot),'slot symbol');
        if(text===row.before){assert(!check,'missing public contract repair: overload '+index);edits.push({start:slot.type.getStart(file),end:slot.type.end,text:row.after});}
    }
    let changed=file.text;
    for(const edit of edits.sort((a,b)=>b.start-a.start))changed=changed.slice(0,edit.start)+edit.text+changed.slice(edit.end);
    assert.equal(ts.createSourceFile(file.fileName,changed,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
    for(const removeComments of [false,true]){
        const options={fileName:file.fileName,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
        assert.equal(ts.transpileModule(file.text,options).outputText,ts.transpileModule(changed,options).outputText,'JavaScript changed');
    }
    assert.equal(fs.readFileSync(file.fileName,'utf8'),file.text,'source changed while planning');
    if(!check&&changed!==file.text)fs.writeFileSync(file.fileName,changed);
    return {files:Number(changed!==file.text),declarations:edits.length,javascriptIdentical:true,declaredPublicApiChanges:1};
}
module.exports={adapt};
if(require.main===module){const check=process.argv[2]==='--check';assert.equal(process.argv.length,check?4:3,'usage: adapt.cjs [--check] TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[check?3:2]),check),null,2));}
