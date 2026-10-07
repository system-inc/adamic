import fs from 'node:fs';
import {createRequire} from 'node:module';
const require = createRequire(import.meta.url);
const fork = process.argv[4] === 'fork';
const prettier = require(fork ? process.argv[2]+'/standalone.js' : process.argv[2]);
const plugins = fork ? ['estree','typescript','babel','postcss','markdown','graphql','yaml'].map(name => require(process.argv[2]+'/plugins/'+name+'.js')) : undefined;
if(prettier.version !== '3.9.6') throw new Error(`expected 3.9.6, got ${prettier.version}`);
for(const line of fs.readFileSync(process.argv[3], 'utf8').split('\n')) {
    if(!line) continue;
    const {name,text} = JSON.parse(line);
    const result = {name};
    for(const mode of process.argv[5] === 'off-only' ? ['off'] : ['auto','off']) {
        try {
            result[mode] = await prettier.format(text,{parser:'markdown',tabWidth:4,useTabs:false,semi:true,singleQuote:true,printWidth:120,embeddedLanguageFormatting:mode,plugins});
        }
        catch(error) { result[mode+'Error'] = String(error); }
    }
    process.stdout.write(JSON.stringify(result)+'\n');
}
