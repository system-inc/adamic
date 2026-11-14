// Positive end-to-end controls: run the actual AST instrumenter, then execute JS.
const fs=require('fs'),path=require('path'),cp=require('child_process'),assert=require('assert'),ts=require('../../api/node_modules/typescript');
const root='/workspace/scratch/receiver-counter-controls',output='/workspace/scratch/split-fixture-controls.json';
fs.mkdirSync(root+'/src/compiler',{recursive:true});fs.mkdirSync(root+'/src/compiler/_namespaces',{recursive:true});
fs.writeFileSync(root+'/src/compiler/tsconfig.json',JSON.stringify({compilerOptions:{strict:true,target:'ES2022',module:'CommonJS'}}));
fs.writeFileSync(root+'/src/compiler/_namespaces/ts.ts',`export interface NodeFactory { createIdentifier(a: string): string; } export interface TransformationContext { hoistVariableDeclaration(node: string): void; flag: boolean; } export interface PrintHandlers { onEmitNode?: (x: string) => void; }`);
fs.writeFileSync(root+'/src/compiler/parser.ts',`import { NodeFactory, TransformationContext, PrintHandlers } from './_namespaces/ts.js';
const factory: NodeFactory = {createIdentifier(a) { return a; }};
const original = factory.createIdentifier;
factory.createIdentifier('ok');
factory.createIdentifier === original;
original === original;
factory.createIdentifier.length;
factory.createIdentifier = original;
Object.defineProperty(factory, 'createIdentifier', {value: original, configurable: true});
const context: TransformationContext = { flag: true, hoistVariableDeclaration(node) { if (this.flag) void node; } };
const hoistVariableDeclaration = (node: string): void => context.hoistVariableDeclaration(node);
hoistVariableDeclaration('ok');
function noEmitNotification(x: string): void { void x; }
function createPrinter(handlers: PrintHandlers = {}) {
 const handlersView = handlers;
 const onEmitNode = handlersView.onEmitNode ?? noEmitNotification;
 onEmitNode === noEmitNotification;
 return onEmitNode;
}
createPrinter({onEmitNode: noEmitNotification});
`);
const originalFixture=fs.readFileSync(root+'/src/compiler/parser.ts','utf8'),splitAt=originalFixture.indexOf('function noEmitNotification');
const contextAt=originalFixture.indexOf('const context:');fs.mkdirSync(root+'/src/compiler/transformers',{recursive:true});
fs.writeFileSync(root+'/src/compiler/parser.ts',originalFixture.slice(0,contextAt));
fs.writeFileSync(root+'/src/compiler/transformers/es2020.ts',`import { TransformationContext } from '../_namespaces/ts.js';\n`+originalFixture.slice(contextAt,splitAt));
fs.writeFileSync(root+'/src/compiler/emitter.ts',`import { PrintHandlers } from './_namespaces/ts.js';\n`+originalFixture.slice(splitAt));
cp.execFileSync(process.execPath,[path.join(__dirname,'split-instrument.cjs'),root,'/workspace/scratch/receiver-heat.json',root+'/instrument.json']);
const input=fs.readFileSync(root+'/src/compiler/parser.ts','utf8');const js=ts.transpileModule(input,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS}}).outputText;
fs.writeFileSync(root+'/control.cjs',js);require('./split-hooks.cjs');require(root+'/control.cjs');const emitterJS=ts.transpileModule(fs.readFileSync(root+'/src/compiler/emitter.ts','utf8'),{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS}}).outputText;fs.writeFileSync(root+'/emitter.cjs',emitterJS);require(root+'/emitter.cjs');const transformerJS=ts.transpileModule(fs.readFileSync(root+'/src/compiler/transformers/es2020.ts','utf8'),{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS}}).outputText;fs.writeFileSync(root+'/transformer.cjs',transformerJS);require(root+'/transformer.cjs');const snapshot=globalThis.__receiverPrice.snapshot();
for(const key of ['parserCalls.createIdentifier','transformerCalls.es2020.ts.hoistVariableDeclaration','printerEntries','printerAliases.onEmitNode','receiverIdentity','receiverLength','printerAliasSentinelIdentity','astWrites.NodeFactory.createIdentifier','writes.NodeFactory.createIdentifier','defineWrites.createIdentifier'])assert.strictEqual(snapshot.counts[key],1,key);
assert.strictEqual(snapshot.counts.functionLength,1);assert.strictEqual(snapshot.counts.memberIdentityComparisons,2);assert.strictEqual(snapshot.counts.memberIdentityOperands,4);assert(snapshot.failures.length===0);fs.writeFileSync(output,JSON.stringify({pass:true,...snapshot},null,2)+'\n');
console.log('PASS: actual AST instrumenter fires every requested counter');
