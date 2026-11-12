// Extract the exact owner and full performanceCore source, following CLI's witness.
// oracle/node.mjs supplies its real ES-module stripping runner, without require.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require('typescript');
const {execFileSync}=require('node:child_process');
const [main,tip,output]=process.argv.slice(2);assert(main&&tip&&output,'usage: require-drop-witnesses.cjs MAIN_TREE TIP_TREE OUTPUT.json');
const root=path.resolve(__dirname,'../../..'),out=output+'.probes';fs.mkdirSync(out);
const observations={node:process.version,runner:'oracle/node.mjs',mode:'module; stripTypeScriptTypes transform; source owners and full performanceCore'};
for(const [side,tree] of [['main',main],['tip',tip]]) {
 const dir=path.join(out,side);fs.mkdirSync(dir);
 const core=ts.createSourceFile('core.ts',fs.readFileSync(tree+'/src/compiler/core.ts','utf8'),ts.ScriptTarget.Latest,true);
 const fn=core.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='isNodeLikeSystem');assert.equal(fn.length,1);
 const declarations=core.statements.filter(n=>ts.isVariableStatement(n)&&n.modifiers?.some(m=>m.kind===ts.SyntaxKind.DeclareKeyword));
 fs.writeFileSync(dir+'/core.a',[...declarations,fn[0]].map(n=>n.getText(core)).join('\n'));
 const perf=fs.readFileSync(tree+'/src/compiler/performanceCore.ts','utf8');
 const anchor='import { isNodeLikeSystem } from "./_namespaces/ts.js";';assert.equal(perf.split(anchor).length,2);
 fs.writeFileSync(dir+'/performance.a',perf.replace(anchor,'import { isNodeLikeSystem } from "./core.a";'));
 fs.writeFileSync(dir+'/main.a',`import {isNodeLikeSystem} from './core.a';\nimport {tryGetNativePerformanceHooks} from './performance.a';\nimport * as perf from 'node:perf_hooks';\nconst hooks=tryGetNativePerformanceHooks();\nconsole.log(JSON.stringify({isNodeLikeSystem:isNodeLikeSystem(),requireType:typeof require,shouldWriteNativeEvents:hooks?.shouldWriteNativeEvents,samePerformance:hooks?.performance===perf.performance,path:hooks?.shouldWriteNativeEvents?'browser global':'node:perf_hooks'}));\n`);
 observations[side]=JSON.parse(execFileSync(process.execPath,['--disable-warning=ExperimentalWarning',root+'/oracle/node.mjs',dir+'/main.a'],{encoding:'utf8'}));
}
assert.equal(observations.main.isNodeLikeSystem,false);assert.equal(observations.tip.isNodeLikeSystem,true);assert.equal(observations.tip.path,'node:perf_hooks');assert.equal(observations.tip.samePerformance,true);
fs.writeFileSync(output,JSON.stringify(observations,null,2)+'\n');console.log(JSON.stringify(observations,null,2));
