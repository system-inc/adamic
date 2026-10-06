// Independent full-formatting oracle, pinned npm Prettier or cohere's embedded fork.
import {readFileSync} from 'node:fs';
import {createRequire} from 'node:module';
const require = createRequire(import.meta.url);
const [library,cases,mode = 'default',variant = 'npm'] = process.argv.slice(2);
let prettier,plugins;
if(variant === 'fork') {
  prettier = require(`${library}/standalone.js`);
  plugins = ['estree','typescript','babel','postcss','markdown','graphql','yaml'].map(name => require(`${library}/plugins/${name}.js`));
} else {
  prettier = require(`${library}/node_modules/prettier`);
  if(prettier.version !== '3.9.6') throw new Error(`expected Prettier 3.9.6, got ${prettier.version}`);
}
const options = mode === 'narrow' ? {printWidth:24,tabWidth:4,useTabs:true,singleQuote:true,trailingComma:'none'} : {printWidth:80,tabWidth:2,useTabs:false,singleQuote:false,trailingComma:'all'};
const unescape = s => s.replace(/\\([\\nrt])/g,(_,c)=>({n:'\n',r:'\r',t:'\t','\\':'\\'}[c]));
let index = 0; let formattedCount = 0; let units = 0;
const countOnly = process.argv[6] === 'count';
const lines = readFileSync(cases,'utf8').split('\n').filter(Boolean).map(line=>({text:unescape(line.slice(2)),scss:line[1]==='S'}));
for(let round = 0; round < (countOnly && process.argv[7] !== 'once' ? 10 : 1); round++) {
for(const line of lines) {
  const text = line.text;
  let answer;
  try { const formatted = await prettier.format(text,{...options,parser:line.scss ? 'scss' : 'css',...(plugins ? {plugins} : {})}); formattedCount++; units += formatted.length; if(!countOnly) answer = JSON.stringify(formatted); }
  catch(e) { answer = 'error '+JSON.stringify({name:e.name,message:e.message,loc:e.loc ?? null}); }
  if(!countOnly) process.stdout.write(`case ${index}\n${answer}\n`);
  index++;
}
}
if(countOnly) process.stdout.write(`${formattedCount} of ${index} stylesheets formatted, ${units} units\n`);
