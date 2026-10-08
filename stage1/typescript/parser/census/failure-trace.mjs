import {readFileSync} from 'node:fs';
import {registerHooks,stripTypeScriptTypes} from 'node:module';
registerHooks({resolve(s,c,next){if(s==='adamic')return {url:new URL('../../../../oracle/adamic.mjs',import.meta.url).href,shortCircuit:true};return next(s,c)},load(u,c,next){if(u.endsWith('.ts'))return {format:'module',source:stripTypeScriptTypes(readFileSync(new URL(u),'utf8'),{mode:'transform'}),shortCircuit:true};return next(u,c)}});
const {Parser}=await import('../parser.ts');
const {written}=await import('../nodes.ts');
const counts=new Map();
for(const name of Object.getOwnPropertyNames(Parser.prototype)){
 const original=Parser.prototype[name];if(typeof original!=='function'||name==='constructor')continue;
 Parser.prototype[name]=function(...args){const key=name+':'+this.scanner.fullStart+':'+this.scanner.kind;const count=(counts.get(key)??0)+1;counts.set(key,count);if(count>5000)throw Error('Repeated parser operation '+key+'\n'+new Error().stack);return original.apply(this,args);};
}
const path=process.argv[2];try{const p=new Parser(readFileSync(path,'utf8'),path);p.file();console.log('completed');}catch(e){console.log(e.stack);}
