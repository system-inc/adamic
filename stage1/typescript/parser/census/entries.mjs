import {readFileSync} from 'node:fs';
import {registerHooks,stripTypeScriptTypes} from 'node:module';
registerHooks({resolve(s,c,next){if(s==='adamic')return {url:new URL('../../../../oracle/adamic.mjs',import.meta.url).href,shortCircuit:true};return next(s,c)},load(u,c,next){if(u.endsWith('.ts'))return {format:'module',source:stripTypeScriptTypes(readFileSync(new URL(u),'utf8'),{mode:'transform'}),shortCircuit:true};return next(u,c)}});
const {Parser}=await import('../parser.ts');
const {written}=await import('../nodes.ts');
const rows=[];
for(const method of ['expression','assignment','rootExpression','rootAssignment','allowInExpression','allowInAssignment'])for(const context of [false,true]){
 const p=new Parser('({m(){}})','probe.ts');if(context)p.beginList('source');const root=p[method]();if(context)p.endList('source');
 rows.push({method,sourceContext:context,root:p.nodes[root].kind,next:p.kind(),diagnostics:p.diagnostics,listContexts:p.listContexts});
}
process.stdout.write(JSON.stringify(rows,null,2)+'\n');
