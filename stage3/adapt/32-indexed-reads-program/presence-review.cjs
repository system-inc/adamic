#!/usr/bin/env node
"use strict";
// Stock symbol resolution, plus a Node witness at the public API escape.
const fs=require("node:fs"), path=require("node:path"), assert=require("node:assert/strict");
const ts=require("typescript");assert.equal(ts.version,"6.0.3");
const tree=path.resolve(process.argv[2]);
const roots=ts.sys.readDirectory(path.join(tree,"src"),[".ts"],undefined,undefined);
const program=ts.createProgram(roots,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true});
const checker=program.getTypeChecker();
const names=new Set(["createProgramHost","createCompilerHostFromProgramHost","createWatchCompilerHost","createWatchCompilerHostOfConfigFile","createWatchCompilerHostOfFilesAndCompilerOptions","createSolutionBuilderHostBase","createSolutionBuilderHost","createSolutionBuilderWithWatchHost"]);
const symbols=new Map(), references=[], locals=new Set(), localReferences=[];
function owner(n){while(n&&!ts.isFunctionLike(n))n=n.parent;return n?.name?.getText()||"<module>";}
function symbol(n){let s=checker.getSymbolAtLocation(n);if(s&&(s.flags&ts.SymbolFlags.Alias))s=checker.getAliasedSymbol(s);return s;}
function visit(n,fn){fn(n);ts.forEachChild(n,c=>visit(c,fn));}
const sources=program.getSourceFiles().filter(sf=>sf.fileName.startsWith(tree+path.sep+"src"+path.sep));
for(const sf of sources)visit(sf,n=>{if(ts.isFunctionDeclaration(n)&&n.name&&names.has(n.name.text)&&sf.fileName.includes("/compiler/"))symbols.set(symbol(n.name),n.name.text);});
for(const sf of sources)visit(sf,n=>{
 if(ts.isVariableDeclaration(n)&&ts.isIdentifier(n.name)&&["compilerHost","host","result"].includes(n.name.text)) {
  const scope=owner(n);
  if(["createCompilerHostFromProgramHost","createWatchProgram","createSolutionBuilderState","createWatchCompilerHost","createSolutionBuilderHostBase","createSolutionBuilderHost","createSolutionBuilderWithWatchHost","createWatchCompilerHostOfConfigFile","createWatchCompilerHostOfFilesAndCompilerOptions"].includes(scope))locals.add(symbol(n.name));
 }
});
function entry(n){const sf=n.getSourceFile();const lc=sf.getLineAndCharacterOfPosition(n.getStart(sf));let context=n.parent;while(context.parent&&!ts.isStatement(context)&&!ts.isVariableDeclaration(context))context=context.parent;return {file:path.relative(tree,sf.fileName),line:lc.line+1,scope:owner(n),kind:ts.SyntaxKind[n.parent.kind],context:context.getText(sf)};}
for(const sf of sources)visit(sf,n=>{if(ts.isIdentifier(n)){const s=symbol(n);if(symbols.has(s))references.push({symbol:symbols.get(s),...entry(n)});if(locals.has(s))localReferences.push(entry(n));}});
// Scan every reference's enclosing statement for all requested presence observers.
const observers=localReferences.filter(r=>/\bin\b|hasOwn|Object\.(keys|assign)|\.\.\.|\bfor\s*\(/.test(r.context));
const stock=require(path.join(tree,"built/local/typescript.js"));
const system={...stock.sys};delete system.createHash;delete system.realpath;delete system.getEnvironmentVariable;
let captured;
const host=stock.createWatchCompilerHost([], {noLib:true,noEmit:true},system,(...args)=>{captured=args[2];return stock.createEmitAndSemanticDiagnosticsBuilderProgram(...args);},()=>{},()=>{});
const watch=stock.createWatchProgram(host);watch.close();assert(captured);
function observe(o,key){return {in:key in o,hasOwn:Object.prototype.hasOwnProperty.call(o,key),keys:Object.keys(o).includes(key),spread:Object.prototype.hasOwnProperty.call({...o},key),forIn:(()=>{for(const k in o)if(k===key)return true;return false;})(),assign:Object.prototype.hasOwnProperty.call(Object.assign({},o),key)};}
const witnesses=[];
for(const [site,o] of [["createProgramHost",host],["createCompilerHostFromProgramHost",captured]]){
 assert.equal(o.createHash,undefined);const before=observe(o,"createHash");assert(Object.values(before).every(Boolean));
 const omitted={...o};delete omitted.createHash;const after=observe(omitted,"createHash");assert(Object.values(after).every(v=>v===false));
 witnesses.push({site,key:"createHash",before,after,verdict:"omission changes six observable results; decline"});
}
console.log(JSON.stringify({typescript:ts.version,source_files:sources.length,symbol_references:references,host_local_references:localReferences,potential_observers:observers,public_escape_witnesses:witnesses,decision:"decline both watch.ts TS2375 sites; public objects or callback arguments escape to arbitrary consumers"},null,2));
