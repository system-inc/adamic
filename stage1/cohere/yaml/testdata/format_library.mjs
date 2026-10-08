import {createRequire} from 'node:module';
import {readFileSync} from 'node:fs';
import {join} from 'node:path';
const [library,,path]=process.argv.slice(2);
const require=createRequire(join(library,'package.json'));
const prettier=require('prettier');
if(prettier.version!=='3.9.6')throw new Error('expected prettier 3.9.6');
const unescape=text=>text.replace(/\\([\\nrt])/g,(_,char)=>char==='n'?'\n':char==='r'?'\r':char==='t'?'\t':char);
const escape=text=>text.replaceAll('\\','\\\\').replaceAll('\n','\\n').replaceAll('\r','\\r').replaceAll('\t','\\t');
for(const line of readFileSync(path,'utf8').split('\n')){
 if(line==='')continue;
 const text=unescape(line.slice(line.indexOf('\t')+1));
 try{console.log(`ok\t${escape(await prettier.format(text,{parser:'yaml',tabWidth:2,printWidth:80,singleQuote:false}))}`);}
 catch(error){const cause=error.cause;console.log(`error\t${escape(cause?.name==='YAMLSyntaxError'?cause.message:error.name+': '+error.message)}`);}
}
