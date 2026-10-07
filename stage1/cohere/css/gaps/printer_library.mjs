// Full Prettier API versus the direct Go stylesheet printer's boundary.
import {createRequire} from 'node:module';
const require = createRequire(import.meta.url);
const root = process.argv[2];
const prettier = require(`${root}/node_modules/prettier`);
if(prettier.version !== '3.9.6') throw new Error('expected Prettier 3.9.6');
for(const text of ['\ufeffa{b:c}','// x\ra{}','\u00a0','---\na:     b\n---\na{}']) {
    try { console.log(JSON.stringify({input:text,output:await prettier.format(text,{parser:'css',printWidth:80,tabWidth:2,useTabs:false,singleQuote:false,trailingComma:'all'})})); }
    catch(e) { console.log(JSON.stringify({input:text,error:e.message})); }
}
