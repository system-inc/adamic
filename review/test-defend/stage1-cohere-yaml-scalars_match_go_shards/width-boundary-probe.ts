import {stringWidth as original} from './width-original/width.ts';
import {stringWidth as mutated} from '/workspace/adamic/stage1/cohere/yaml/width.ts';
const text = String.fromCodePoint(0x3fffd);
console.log('U+3FFFD: original=' + original(text) + ', D3=' + mutated(text));
