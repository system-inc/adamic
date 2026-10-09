import { DocumentArena, printDocument } from '/workspace/adamic/stage1/cohere/markdownblocks/document.ts';
import { printSentence as control } from './D2-control-structure.ts';
import { printSentence as mutant } from './D2-mutant-structure.ts';
for(const print of [control, mutant]) {
 const arena = new DocumentArena();
 const children = [{doc: arena.text('a'), whitespace: false}, {doc: arena.add('h', '', 0, []), whitespace: true}, {doc: arena.text('b'), whitespace: false}, {doc: arena.add('h', '', 0, []), whitespace: true}, {doc: arena.text('c'), whitespace: false}];
 console.log(JSON.stringify(printDocument(arena, print(arena, children), 3, 4)));
}
