// Complete files over supported statement shapes; unsupported syntax fails loudly.
import { Parser } from '../../typescript/parser/parser.ts';
import type { SettingsOptions } from './doc.ts';
import { formatProgram, type ResultType } from './expressions.ts';
import { hasBlankLine } from './syntax.ts';
export function formatFile(source: string, settings: SettingsOptions): ResultType {
    if(source.startsWith('#!') || source.startsWith('\ufeff#!') || source.includes('/*') || source.includes('//'))
        return { kind: 'NotYet', reason: 'comment-attachment' };
    if(hasBlankLine(source) || source.includes('\r')) return { kind: 'NotYet', reason: 'source-trivia' };
    const parser = new Parser(source, 'statements.ts');
    return formatProgram(parser, source, settings);
}
