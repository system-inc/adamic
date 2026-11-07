'use strict';
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const body='if(values!==undefined){for(let i=0;i<values.length;i++){const value=values[i]!;const key=`${keySelector(value)}`;const array=result[key]??=[];array.push(value);}}return result;';
const cases={
 'groupby-result-as-return-type':'function g<T,K extends string|number|boolean|null|undefined>(values:readonly T[]|undefined,keySelector:(v:T)=>K):{[P in K as `${P}`]?:T[]} {const result:{[P in K as `${P}`]?:T[]}={};'+body+'}',
 'groupby-record-result-mapped-return':'function g<T,K extends string|number|boolean|null|undefined>(values:readonly T[]|undefined,keySelector:(v:T)=>K):{[P in K as `${P}`]?:T[]} {const result:Record<string,T[]>={};'+body+'}',
 'groupby-record-intersection':'function g<T,K extends string|number|boolean|null|undefined>(values:readonly T[]|undefined,keySelector:(v:T)=>K):{[P in K as `${P}`]?:T[]} {const result:Record<string,T[]> & {[P in K as `${P}`]?:T[]}={};'+body+'}',
 'groupby-record-implementation':'function g<T,K extends string|number|boolean|null|undefined>(values:readonly T[]|undefined,keySelector:(v:T)=>K):Record<string,T[]> {const result:Record<string,T[]>={};'+body+'}',
 'clone-direct-read':'function clone<T>(object:T):T {const result:any={};for(const id in object)if(Object.prototype.hasOwnProperty.call(object,id))result[id]=object[id];return result;}',
 'clone-record-constraint':'function clone<T extends Record<string,unknown>>(object:T):T {const result:any={};for(const id in object)if(Object.prototype.hasOwnProperty.call(object,id))result[id]=object[id];return result;}'
};
const results=[];
for(const [name,source]of Object.entries(cases)){
 const file='/tmp/type-trial.ts',options={strict:true,noEmit:true,noUncheckedIndexedAccess:true,target:ts.ScriptTarget.ES2024},h=ts.createCompilerHost(options),old=h.getSourceFile;
 h.getSourceFile=(f,...a)=>f===file?ts.createSourceFile(f,source,ts.ScriptTarget.Latest,true):old(f,...a);
 const p=ts.createProgram([file],options,h),ds=ts.getPreEmitDiagnostics(p);results.push({name,source,diagnostics:ds.map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,' ')}))});
}
console.log(JSON.stringify({scope:'isolated strict TypeScript trials; selected complete owners are separately checked by Adamic over all 79 roots',results},null,2));
