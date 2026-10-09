import { CSTParser } from './cstParser.ts';
import { Composer } from './composer.ts';
import { UnistContext } from './unistContext.ts';
const parser = new CSTParser();
const context = new UnistContext('', parser, new Composer(parser));
for(const offset of [-1, 0, 1]) { const index = context.offset(offset); const point = context.point(index); console.log(`${offset}: ${point.line},${point.column},${point.offset}`); }
