import { Linter } from '../../../stage1/cohere/lint/lint.ts';
import { Settings } from '../../../stage1/cohere/lint/settings.ts';
import { Parser } from '../../../stage1/typescript/parser/parser.ts';
import { Scanner } from '../../../stage1/typescript/scanner/scanner.ts';
const source = 'debugger;';
const parser = new Parser(source, 'source.ts');
const linter = new Linter(source, parser, new Scanner(source), 'all', 'Always', 'Always', false, new Settings());
linter.run();
console.log(`nodes=${parser.nodes.length} findings=${linter.findings.length}`);
