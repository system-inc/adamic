'use strict';
const fs = require('node:fs'), path = require('node:path'), cp = require('node:child_process'), crypto = require('node:crypto');
const ts = require(process.env.SCANNER_TYPESCRIPT);
if (ts.version !== '6.0.3' || process.argv.length < 3) throw Error('usage: SCANNER_TYPESCRIPT=... node escape-proof.cjs NEW_OUTPUT [COMPILER COMPILER_CWD]');
const out = path.resolve(process.argv[2]);fs.mkdirSync(out,{recursive:false});
const source = fs.readFileSync(path.join(__dirname,'main.a'),'utf8'), sf = ts.createSourceFile('main.a',source,ts.ScriptTarget.Latest,true);
const names = ['escapeJsonString','escapeOptionalString','formatDiagnosticArgument'];
const functions = sf.statements.filter(n=>ts.isFunctionDeclaration(n) && names.includes(n.name?.text));
if (functions.length !== names.length) throw Error('missing or duplicate driver helpers');
const helpers=functions.map(n=>n.getText(sf)).join('\n\n');
const fixture=fs.readFileSync(path.join(__dirname,'escape-fixture.a'),'utf8');
const emit=s=>ts.transpileModule(s,{compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.CommonJS}}).outputText;
function run(command,stem,cwd=out) {
 const stdout=fs.openSync(path.join(out,stem+'.stdout'),'w'),stderr=fs.openSync(path.join(out,stem+'.stderr'),'w');
 try {return cp.spawnSync(command[0],command.slice(1),{cwd,stdio:['ignore',stdout,stderr]}).status;}finally{fs.closeSync(stdout);fs.closeSync(stderr);}
}
const actual=helpers+'\n'+fixture+'\nexerciseEscaper(escapeJsonString, formatDiagnosticArgument);\n';
fs.writeFileSync(path.join(out,'actual.a'),actual);fs.writeFileSync(path.join(out,'actual.cjs'),emit(actual));
fs.writeFileSync(path.join(out,'oracle.cjs'),emit(fixture+'\nexerciseEscaper(text => JSON.stringify(text), value => JSON.stringify(value ?? null));\n'));
if(run(['node',path.join(out,'oracle.cjs')],'oracle')!==0 || run(['node',path.join(out,'actual.cjs')],'actual')!==0)throw Error('fixture execution failed');
if(run(['diff','-u','oracle.stdout','actual.stdout'],'comparison')!==0)throw Error('escaper differs from Node JSON.stringify');
const branch = actual.split('\n').find(line => line.includes('else if (code === 10)'));
if(!branch || actual.split(branch).length!==2)throw Error('newline escape mutant site changed');
const mutant=actual.replace(branch,"else if (code === 10) result += text.charAt(index);");
fs.writeFileSync(path.join(out,'mutant.a'),mutant);fs.writeFileSync(path.join(out,'mutant.cjs'),emit(mutant));
if(run(['node',path.join(out,'mutant.cjs')],'mutant')!==0 || run(['diff','-u','oracle.stdout','mutant.stdout'],'mutant-diff')!==1)throw Error('dropped escape mutant escaped comparison');
const report={asciiCodeUnits:128,stringCases:140,payloadCases:5,nodeComparisonExit:0,droppedNewlineEscapeMutant:{nodeExit:0,diffExit:1},driverHelperSha256:crypto.createHash('sha256').update(helpers).digest('hex')};
if(process.argv[3]) {
 const compiler=path.resolve(process.argv[3]),cwd=path.resolve(process.argv[4]);
 report.nativeBuildExit=run([compiler,'build',path.join(out,'actual.a'),'-o',path.join(out,'native')],'native-build',cwd);
 if(report.nativeBuildExit!==0) throw Error('native fixture build failed; see native-build.stderr');
 if(report.nativeBuildExit===0){report.nativeExit=run([path.join(out,'native')],'native');report.nativeDiffExit=run(['diff','-u','oracle.stdout','native.stdout'],'native-diff');if(report.nativeExit!==0 || report.nativeDiffExit!==0)throw Error('native escaper disagrees with Node');
 report.nativeMutantBuildExit=run([compiler,'build',path.join(out,'mutant.a'),'-o',path.join(out,'native-mutant')],'native-mutant-build',cwd);
 if(report.nativeMutantBuildExit!==0)throw Error('native mutant rejected before comparison');report.nativeMutantExit=run([path.join(out,'native-mutant')],'native-mutant');report.nativeMutantDiffExit=run(['diff','-u','oracle.stdout','native-mutant.stdout'],'native-mutant-diff');if(report.nativeMutantExit!==0 || report.nativeMutantDiffExit!==1)throw Error('native escape mutant escaped comparison');}
 }
fs.writeFileSync(path.join(out,'report.json'),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report));
