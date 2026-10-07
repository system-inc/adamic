import { identifier } from '../identifier.ts';
const output: string[] = [];
for(let code = 0; code <= 1114111; code++) {
    if(code >= 55296 && code <= 57343) continue;
    output.push(`${identifier(String.fromCodePoint(code)).codePointAt(0) ?? -1}`);
}
console.log(output.join('\n'));
