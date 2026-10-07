const fs=require('node:fs'),path=require('node:path');
const tree=path.resolve(process.argv[2]);
if(!fs.existsSync(path.join(tree,'slice.json')))throw new Error('adaptation 81 requires a declaration slice');
const file=path.join(tree,'src/compiler/scanner.ts'),before=fs.readFileSync(file,'utf8');
const after=before.replace('function reScanJsxToken(allowMultilineJsxText = true)', 'function reScanJsxToken(allowMultilineJsxText: boolean | undefined = true)');
if(after!==before)fs.writeFileSync(file,after);
console.log(JSON.stringify({jsxParameter:after===before?0:1}));
