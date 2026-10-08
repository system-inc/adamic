// Complete files over supported statement shapes; unsupported syntax fails loudly.
import { Parser } from '../../typescript/parser/parser.ts';
import { Documents, type SettingsOptions } from './doc.ts';
import { Expressions, type ResultType } from './expressions.ts';
import { hasBlankLine } from './syntax.ts';
export function formatFile(source: string, settings: SettingsOptions): ResultType {
    if(source.startsWith('#!') || source.startsWith('\ufeff#!') || source.includes('/*') || source.includes('//'))
        return { kind: 'NotYet', reason: 'comment-attachment' };
    if(hasBlankLine(source) || source.includes('\r')) return { kind: 'NotYet', reason: 'source-trivia' };
    const parser = new Parser(source, 'statements.ts');
    const root = parser.file();
    const docs = new Documents(settings);
    const printer = new Expressions(parser, source, docs);
    const reason = printer.statementUnsupported(root);
    if(reason !== '') return { kind: 'NotYet', reason };
    const printed = docs.print(printer.statementDoc(printer.normalize(root), -1));
    return { kind: 'Ok', text: printed === '' ? '' : `${printed}\n` };
}
