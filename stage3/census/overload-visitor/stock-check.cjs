'use strict';
const fs=require('node:fs'),path=require('node:path');
const ts=require('../../api/node_modules/typescript');
const {compiler}=require('../../adapt/47-host-errors/adapt.cjs');
const tree=path.resolve(process.argv[2]),out=process.argv[3],initial=compiler(tree,ts);
const program=ts.createProgram(initial.getRootFileNames(),{...initial.getCompilerOptions(),typeRoots:[path.resolve(__dirname,'../../api/node_modules/@types')]});
const roots=program.getRootFileNames(),diagnostics=ts.getPreEmitDiagnostics(program);
const result={typescript:ts.version,roots:roots.length,diagnosticCount:diagnostics.length,diagnostics:diagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n'))};
fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({tree,roots:roots.length,diagnostics:diagnostics.length}));process.exitCode=diagnostics.length?1:0;
