import { written } from '../encode.ts';
const inputs: readonly string[] = ['', 'hello', '\n', '\r\t', '\u0000', '\\', '\u007f', 'é', '😀', '\ud800', '\udfff', 'a\\b\nc', ' ~'];
for(const input of inputs) console.log(written(input));
