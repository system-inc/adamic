'use strict';
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const ts=require('typescript');
const control=path.resolve(process.argv[2]),probe=path.resolve(process.argv[3]);
const hash=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const diagnostics=tree=>{const file=path.join(tree,'built/local/typescript.d.ts');const program=ts.createProgram([file],{strict:true,noEmit:true,skipLibCheck:false,target:ts.ScriptTarget.ES2024,types:[]});return ts.getPreEmitDiagnostics(program).map(d=>({file:d.file&&path.relative(tree,d.file.fileName),line:d.file&&d.file.getLineAndCharacterOfPosition(d.start).line+1,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,' ')}));};
function files(dir){return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(dir,e.name)):[path.join(dir,e.name)]);}
const javascript=files(path.join(control,'built')).filter(f=>/\.(js|mjs|cjs)$/.test(f)).map(f=>{const relative=path.relative(control,f),other=path.join(probe,relative);return {file:relative,control:hash(f),probe:fs.existsSync(other)?hash(other):null};});
const result={typescript:ts.version,controlDiagnostics:diagnostics(control),probeDiagnostics:diagnostics(probe),javascript,javascriptIdentical:javascript.every(r=>r.control===r.probe)};
if(result.controlDiagnostics.length!==0||!result.probeDiagnostics.some(d=>d.code===2304&&d.message.includes('AdamicJsonRecoveryObject'))||!result.javascriptIdentical)throw Error(JSON.stringify(result));
console.log(JSON.stringify(result,null,2));
