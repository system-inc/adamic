const fs=require('fs'),path=require('path'),cp=require('child_process');
const ts=require('../../../api/node_modules/typescript');
const norm=require('../../../lane/normalize-api.cjs');
const label=process.argv[2],baseLabel=process.argv[3]||'original',out=path.join(__dirname,label);
function read(dir){let p=path.join(__dirname,dir,'typescript.d.ts');return fs.existsSync(p)?fs.readFileSync(p,'utf8'):require('zlib').gunzipSync(fs.readFileSync(p+'.gz')).toString();}
const text=read(label),before=read(baseLabel),sf=ts.createSourceFile('typescript.d.ts',text,ts.ScriptTarget.Latest,true);
fs.writeFileSync(path.join(out,'typescript.d.ts'),text);
const selected=[];
function visit(n){if((ts.isInterfaceDeclaration(n)&&['LineMapSource','SourceFileLike'].includes(n.name.text))||(ts.isFunctionDeclaration(n)&&['getLineStarts','getLineAndCharacterOfPosition'].includes(n.name?.text)))selected.push({name:n.name.text,kind:ts.SyntaxKind[n.kind],text:n.getText(sf)});n.forEachChild(visit);}visit(sf);
const consumer=path.join(out,'public-sourcefile-consumer.ts');fs.writeFileSync(consumer,'import ts = require("./typescript");\ndeclare const file: ts.SourceFile;\nts.getLineAndCharacterOfPosition(file, 0);\n');
const options={strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,noEmit:true,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.CommonJS,types:[],skipLibCheck:false};
const program=ts.createProgram([consumer],options);
const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>{const pos=d.file&&d.start!=null?d.file.getLineAndCharacterOfPosition(d.start):null;return {code:d.code,file:d.file?path.basename(d.file.fileName):null,line:pos?pos.line+1:null,column:pos?pos.character+1:null,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};});
const baselineFile=path.join(out,'comparison-baseline.d.ts');fs.writeFileSync(baselineFile,before);
const diff=cp.spawnSync('diff',['-u','--label',baseLabel+'/typescript.d.ts','--label',label+'/typescript.d.ts',baselineFile,path.join(out,'typescript.d.ts')],{encoding:'utf8',maxBuffer:16*1024*1024});
fs.writeFileSync(path.join(out,'typescript-public.diff'),diff.stdout||'');fs.writeFileSync(path.join(out,'public-diff.status.json'),JSON.stringify({child_exit:diff.status,stderr:diff.stderr},null,2)+'\n');if(![0,1].includes(diff.status))throw Error('diff failed');fs.unlinkSync(baselineFile);
const changes=norm.changes(before,text);fs.writeFileSync(path.join(out,'normalized-api-changes.json'),JSON.stringify(changes,null,2)+'\n');
const result={label,baseline:baseLabel,typescript:ts.version,selected_declarations:selected,options,diagnostics,normalized_changes:changes.length};fs.writeFileSync(path.join(out,'api-analysis.json'),JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({label,diagnostics:diagnostics.length,selected,normalized_changes:changes.length}));
