const fs=require('node:fs'), path=require('node:path');
const tree=path.resolve(process.argv[2]);
if(!fs.existsSync(path.join(tree,'slice.json'))) throw new Error('adaptation 59 requires a declaration slice');
const file=path.join(tree,'src/compiler/scanner.ts'), before=fs.readFileSync(file,'utf8');
const after=before.replace('let operand!: string;','let operand: string | undefined;');
if(after!==before) fs.writeFileSync(file,after);
console.log(JSON.stringify({optionalOperand:after===before?0:1}));
