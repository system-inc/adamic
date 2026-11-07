'use strict';
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),crypto=require('node:crypto');
const s='/workspace/scratch/gov-b',repo='/workspace/adamic';
cp.execFileSync('git',['diff','--quiet','5a3e6607','origin/main','--','stage3/adapt','stage3/api','stage3/apply.py','stage3/source.json'],{cwd:repo});
fs.mkdirSync(s+'/raw',{recursive:true});
const changes=[['    clear,\r\n',''],['    type MappedType,','    type MapLike,\r\n    type MappedType,'],['const visitedTypes: Type[] = []; // Sparse array from id to type','const visitedTypes: MapLike<Type> = {}; // Record from id to type'],['const visitedSymbols: Symbol[] = []; // Sparse array from id to symbol','const visitedSymbols: MapLike<Symbol> = {}; // Record from id to symbol'],['clear(visitedTypes);','for (const key in visitedTypes) delete visitedTypes[key];'],['clear(visitedSymbols);','for (const key in visitedSymbols) delete visitedSymbols[key];']];
for(const side of ['before','after']){
 const tree=s+'/'+side;cp.execFileSync('cp',['-a','/workspace/scratch/gov-a/counted',tree]);
 const core=tree+'/src/compiler/core.ts';let text=fs.readFileSync(core,'utf8');const old=' (globalThis as any).__getOwnValuesCensus(collection);';if(text.split(old).length!==2)throw Error('old counter');text=text.replace(old,'');
 if(side==='after')text=text.replace('getOwnValues<T>(collection: MapLike<T> | T[])','getOwnValues<T>(collection: MapLike<T>)').replace('(collection as MapLike<T>)[key]!','collection[key]!');
 if(text.split('    return values;').length<2)throw Error('return values');
 const start=text.indexOf('export function getOwnValues<T>'),end=text.indexOf('\n}',start)+2;const owner=text.slice(start,end);if(!owner.includes('return values;'))throw Error('owner');
 text=text.slice(0,start)+owner.replace('return values;','return (globalThis as any).__govValues(collection, values);')+text.slice(end);fs.writeFileSync(core,text);
 let walk=fs.readFileSync(tree+'/src/compiler/symbolWalker.ts','utf8');
 if(side==='after')for(const [from,to] of changes){const n=walk.split(from).length-1;if(n!==(from.startsWith('clear(')?2:1))throw Error('walker shape '+from);walk=walk.replaceAll(from,to);}
 for(const [from,to] of [['visitedTypes[type.id] = type;','visitedTypes[(globalThis as any).__govId("type", type.id, type)] = type;'],['visitedSymbols[symbolId] = symbol;','visitedSymbols[(globalThis as any).__govId("symbol", symbolId, symbol)] = symbol;']]){if(walk.split(from).length!==2)throw Error('store count');walk=walk.replace(from,to);}
 fs.writeFileSync(tree+'/src/compiler/symbolWalker.ts',walk);fs.symlinkSync('/workspace/adamic/stage3/typescript/node_modules',tree+'/node_modules','dir');
}
fs.writeFileSync(s+'/raw/plan.json',JSON.stringify({main:cp.execFileSync('git',['rev-parse','origin/main'],{cwd:repo,encoding:'utf8'}).trim(),base:cp.execFileSync('git',['rev-parse','HEAD'],{cwd:repo,encoding:'utf8'}).trim(),walkerOwner:'32-indexed-reads-program',valuesOwner:'30-indexed-reads',changes,measurementOnly:true},null,2)+'\n');
