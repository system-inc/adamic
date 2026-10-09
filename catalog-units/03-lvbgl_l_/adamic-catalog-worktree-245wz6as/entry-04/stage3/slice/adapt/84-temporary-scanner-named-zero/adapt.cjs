const fs=require('node:fs'),path=require('node:path');const tree=path.resolve(process.argv[2]);
if(!fs.existsSync(path.join(tree,'slice.json')))throw new Error('adaptation 84 requires a declaration slice');
const file=path.join(tree,'src/compiler/scanner.ts'),before=fs.readFileSync(file,'utf8'),after=before.replace('tokenFlags = 0;','tokenFlags = TokenFlags.None;');
if(after!==before)fs.writeFileSync(file,after);console.log(JSON.stringify({namedZero:after===before?0:1}));
