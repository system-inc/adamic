const fs = require('node:fs'), path = require('node:path');
const tree = path.resolve(process.argv[2]);
if (!fs.existsSync(path.join(tree,'slice.json'))) throw new Error('adaptation 58 requires a declaration slice');
const file = path.join(tree,'src/compiler/scanner.ts');
let text = fs.readFileSync(file,'utf8'), changes = 0;
for (const [before,after] of [
    ['var text: string = undefined!;', 'var text: string;'],
    ['var tokenValue!: string;', 'var tokenValue: string;'],
]) {
    if (text.includes(before)) {
        if (text.indexOf(before) !== text.lastIndexOf(before)) throw new Error('duplicate declaration');
        text = text.replace(before,after); changes++;
    }
}
if (changes) fs.writeFileSync(file,text);
console.log(JSON.stringify({changes}));
