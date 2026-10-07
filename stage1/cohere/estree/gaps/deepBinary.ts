import { utf8Length } from 'adamic';
import { answer } from '../pipeline.ts';
const parts: string[] = [];
for(let index = 0; index < 4096; index++) {
    parts.push(index.toString());
}
console.log(utf8Length(answer('source.ts', `${parts.join('+')};`)).toString());
