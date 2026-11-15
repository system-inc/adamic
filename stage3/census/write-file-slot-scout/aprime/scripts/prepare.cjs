const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto');const ts=require('/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/typescript');
const r='/workspace/scratch/write-file-slot-aprime',base='/workspace/scratch/write-file-slot-measure/main';
assert(JSON.parse(fs.readFileSync(r+'/audit.json')).gatePassed);
const gate=JSON.parse(fs.readFileSync(r+'/namespace-summary-scoped.json'));assert(gate.adaptedReads===0&&gate.adaptedControls>0&&gate.lanePass,'count gate incomplete');
const tree=r+'/candidate';assert(!fs.existsSync(tree));
fs.cpSync(base,tree,{recursive:true,filter:p=>!['node_modules','built'].some(n=>p===path.join(base,n))});
const sysFile=tree+'/src/compiler/sys.ts',harnessFile=tree+'/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts';
const raw=fs.readFileSync(sysFile,'utf8'),text=raw.replace(/\r\n/g,'\n');const source=ts.createSourceFile('sys.ts',text,ts.ScriptTarget.Latest,true);const owner=source.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='patchWriteFileEnsuringDirectory');assert(owner&&owner.jsDoc.length===1);
const pinned=JSON.parse(fs.readFileSync('/workspace/scratch/write-file-slot-scout/stage3/adapt/99-correctness-write-file-view/owners.json'))[0];assert.equal(owner.getText(source),pinned.after);
const start=owner.jsDoc[0].getStart(source),end=owner.end;let changed=text.slice(0,start)+text.slice(end+2);assert(!changed.includes('export function patchWriteFileEnsuringDirectory'));
const before='        const nodeSystem: NodeResolutionSystem = {';
const wrapper=`        const patchedWriteFile: System["writeFile"] = (path, data, writeBom) =>
            writeFileEnsuringDirectories(
                path,
                data,
                !!writeBom,
                (path, data, writeByteOrderMark) => writeFile(path, data, writeByteOrderMark),
                path => nodeSystem.createDirectory(path),
                path => nodeSystem.directoryExists(path),
            );
`;
assert.equal(changed.split(before).length,2);changed=changed.replace(before,wrapper+before);assert.equal(changed.split('            writeFile,').length,2);changed=changed.replace('            writeFile,','            writeFile: patchedWriteFile,');
const call=`    if (sys) {
        // patch writefile to create folder before writing the file
        patchWriteFileEnsuringDirectory(sys);
    }
`;assert.equal(changed.split(call).length,2);changed=changed.replace(call,'');assert(!changed.includes('patchWriteFileEnsuringDirectory'));
const hraw=fs.readFileSync(harnessFile,'utf8'),htext=hraw.replace(/\r\n/g,'\n');assert.equal(htext.split('    patchWriteFileEnsuringDirectory,\n').length,2);let h=htext.replace('    patchWriteFileEnsuringDirectory,\n','');
h='import { System, writeFileEnsuringDirectories } from "../../_namespaces/ts.js";\n'+h;
const local=owner.getText(source).replace(/^export /,'');h=h.replace(/\n*$/,'\n\n')+local+'\n';
const hs=ts.createSourceFile('h.ts',h,ts.ScriptTarget.Latest,true),localNode=hs.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='patchWriteFileEnsuringDirectory');assert.equal(localNode.body.getText(hs),owner.body.getText(source));assert.equal(localNode.getText(hs),local);assert(!localNode.modifiers?.some(m=>m.kind===ts.SyntaxKind.ExportKeyword));
assert.equal(hs.parseDiagnostics.length,0);assert.equal(ts.createSourceFile('s.ts',changed,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
fs.writeFileSync(sysFile,changed.replace(/\n/g,'\r\n'));fs.writeFileSync(harnessFile,h.replace(/\n/g,'\r\n'));
const sha=s=>crypto.createHash('sha256').update(s).digest('hex');const receipt={base:'fe5a1a54f045e9bce4e5aa32ca74885a317fbebd',changedFiles:['src/compiler/sys.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts'],helperBodyByteIdentical:true,oldBodySha256:sha(owner.body.getText(source)),movedBodySha256:sha(localNode.body.getText(hs)),localDeclaration:local,wrapper,sourceHashes:Object.fromEntries([sysFile,harnessFile].map(f=>[path.relative(tree,f),sha(fs.readFileSync(f))]))};fs.writeFileSync(r+'/prepare.json',JSON.stringify(receipt,null,2)+'\n');console.log(JSON.stringify(receipt,null,2));
