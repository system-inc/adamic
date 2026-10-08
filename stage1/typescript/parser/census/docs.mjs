import {readFileSync} from 'node:fs';
import {registerHooks,stripTypeScriptTypes} from 'node:module';
registerHooks({resolve(s,c,next){if(s==='adamic')return {url:new URL('../../../../oracle/adamic.mjs',import.meta.url).href,shortCircuit:true};return next(s,c)},load(u,c,next){if(u.endsWith('.ts'))return {format:'module',source:stripTypeScriptTypes(readFileSync(new URL(u),'utf8'),{mode:'transform'}),shortCircuit:true};return next(u,c)}});
const {Parser}=await import('../parser.ts');
const {written}=await import('../nodes.ts');
const source=readFileSync(process.argv[2],'utf8');const p=new Parser(source,'probe.ts');const roots=p.docTypes();console.log(JSON.stringify({source,roots:roots.map(id=>p.nodes[id]),diagnostics:p.diagnostics}));
