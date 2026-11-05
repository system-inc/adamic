'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict');
const root=process.env.VISIBLE_ACTUAL_CONTROLS;
if(root){
 const ts=require(path.join(root,'built/local/typescript.js'));
 const parser=require('typescript');
 const {owners}=require('./owner-tools.cjs');
 const Debug=ts.Debug;Debug.enableDebugInfo();Debug.isDebugging=true;
 const good={__debugTypeToString(){return 'good'}},goodMapper={__debugToString(){return 'good'}};
 const wrap=o=>Debug.attachDebugPrototypeIfDebug(o);
 function control(id,fn){globalThis.__visibleControlTitle=id;try{fn()}catch(e){}finally{globalThis.__visibleControlTitle=null}}
 control('CW144',()=>wrap({kind:ts.TypeMapKind.Simple,source:{},target:good}).__debugToString());
 control('CW145',()=>wrap({kind:ts.TypeMapKind.Simple,source:good,target:{}}).__debugToString());
 control('CW146-array',()=>wrap({kind:ts.TypeMapKind.Array,sources:null,targets:[]}).__debugToString());
 control('CW147-array',()=>wrap({kind:ts.TypeMapKind.Array,sources:[],targets:null}).__debugToString());
 control('CW146-element',()=>wrap({kind:ts.TypeMapKind.Array,sources:[{}],targets:[good]}).__debugToString());
 control('CW147-element',()=>wrap({kind:ts.TypeMapKind.Array,sources:[good],targets:[{}]}).__debugToString());
 control('CW148',()=>wrap({kind:ts.TypeMapKind.Deferred,sources:[{}],targets:[()=>good]}).__debugToString());
 control('CW149',()=>wrap({kind:ts.TypeMapKind.Deferred,sources:[good],targets:[()=>({})]}).__debugToString());
 control('CW150',()=>wrap({kind:ts.TypeMapKind.Merged,mapper1:{},mapper2:goodMapper}).__debugToString());
 control('CW151',()=>wrap({kind:ts.TypeMapKind.Merged,mapper1:goodMapper,mapper2:{}}).__debugToString());
 control('CW143',()=>{const sf=ts.createSourceFile('plant.ts','let x = 1;',ts.ScriptTarget.Latest,true);let reads=0;Object.defineProperty(sf,'kind',{get(){return reads++===0?ts.SyntaxKind.SourceFile:ts.SyntaxKind.Identifier}});sf.__debugGetText()});
 control('CW154',()=>Debug.formatControlFlowGraph({id:0,flags:ts.FlowFlags.BranchLabel,antecedent:{*[Symbol.iterator](){}}}));
 const node=()=>ts.factory.createIdentifier('plant');
 control('DeclarationWithTypeParameters',()=>ts.isDeclarationWithTypeParameters(node()));
 control('DeclarationWithTypeParameterChildren',()=>ts.isDeclarationWithTypeParameterChildren(node()));
 control('PrimitiveLiteral',()=>ts.isPrimitiveLiteralValue(node()));
 control('HasInferredType',()=>ts.hasInferredType(node()));
 function extract(file,owner,bindings){const text=fs.readFileSync(path.join(root,'src/compiler',file),'utf8'),o=owners(text,file).get(owner);assert(o,owner);const js=parser.transpileModule(o.text,{compilerOptions:{target:parser.ScriptTarget.ES2024}}).outputText;const name=o.node.name.text;return new Function(...Object.keys(bindings),js+'\nreturn '+name)(...Object.values(bindings))}
 control('ArrowFunction | FunctionExpression',()=>{
 const fn=extract('expressionToTypeNode.ts','createSyntacticTypeNodeBuilder.typeFromExpression',{SyntaxKind:ts.SyntaxKind,Debug,isFunctionLikeDeclaration:ts.isFunctionLikeDeclaration,withNewScope:(c,n,cb)=>cb(),typeFromFunctionLikeExpression:()=>'(plant)'});
 let reads=0;fn({get kind(){return reads++===0?ts.SyntaxKind.ArrowFunction:ts.SyntaxKind.Identifier}},{});
 });
 control('WithIsolatedDeclarationDiagnostic',()=>{
 const fn=extract('transformers/declarations/diagnostics.ts','createGetIsolatedDeclarationErrors.getDiagnostic',{SyntaxKind:ts.SyntaxKind,Debug,findAncestor:()=>undefined,isHeritageClause:()=>false,isPartOfTypeNode:()=>false,isTypeQueryNode:()=>false,assertType:()=>{},createExpressionError:()=>'(plant)'});fn({kind:ts.SyntaxKind.Identifier,parent:{}});
 });
 control('ModuleResolutionHost',()=>{
 const fn=extract('moduleSpecifiers.ts','getAllModulePathsWorker',{Debug,pathContainsNodeModules:()=>false,getTemporaryModuleResolutionState:()=>{throw Error('stop after counter')}});
 fn({importingSourceFileName:'plant.ts'},'x.ts',{getModuleResolutionCache:()=>({getPackageJsonInfoCache:()=>({})}),getSymlinkCache:()=>({}),readFile:{}},{},{});
 });
}
