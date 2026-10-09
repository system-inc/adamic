import { Parser } from '../../../stage1/typescript/parser/parser.ts';
import { FileCommentCache } from '../../../stage1/cohere/lint/helpers/comments/for_file.ts';
const parser = new Parser('// first\nconst x=1; /* second */\n', 'probe.ts');
const root = parser.file();
const cache = new FileCommentCache(parser, root);
const first = cache.get();
const second = cache.get();
console.log(`counts=${first.length},${second.length} same=${first === second} ready=${cache.ready}`);
