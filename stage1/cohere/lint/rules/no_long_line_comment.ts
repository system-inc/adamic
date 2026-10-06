import { panic, utf8Length } from 'adamic';
import type { RuleContext } from './context.ts';
import type { CommentRange } from './comment_ranges.ts';
import { Finding } from '../finding.ts';
import { policyMessage } from '../volume_messages.ts';
import { trim } from './shared.ts';

function report(context: RuleContext, run: readonly CommentRange[], maximum: number): void {
    if(run.length <= maximum) {
        return;
    }
    const first = run[0] ?? panic('empty comment run');
    const last = run[run.length - 1] ?? panic('empty comment run');
    const indent = ' '.repeat(first.column);
    let replacement = '/*\n';
    let safe = true;
    for(const comment of run) {
        if(comment.text.includes('*/')) {
            safe = false;
        }
        let text = comment.text.slice(2);
        if(text.startsWith(' ')) {
            text = text.slice(1);
        }
        replacement += `${indent} *${trim(text) === '' ? '' : ` ${text}`}\n`;
    }
    replacement += `${indent} */`;
    const rule = 'nexus/consistency-no-long-line-comment';
    context.record(
        new Finding(
            rule,
            'longLineComment',
            policyMessage(rule, 'longLineComment', ['lineCount', run.length.toString()]),
            first.start,
            last.end,
            safe ? 'fix' : '',
            safe ? replacement : '',
            '',
        ),
    );
}
export function noLongLineComment(context: RuleContext, ranges: readonly CommentRange[]): void {
    const configured = parseInt(context.settings.read('maximumlinecount', '4'), 10);
    const maximum = configured > 0 ? configured : 4;
    const sourceLines = context.source.split('\n');
    let run: CommentRange[] = [];
    for(const comment of ranges) {
        const body = trim(comment.text.slice(2));
        let directive = false;
        for(const prefix of [
            'eslint-',
            '@ts-',
            'prettier-ignore',
            'istanbul ',
            'c8 ',
            'v8 ',
            '#__PURE__',
            'global ',
            'globals ',
            'exported ',
        ]) {
            if(body.startsWith(prefix)) {
                directive = true;
            }
        }
        const line = sourceLines[comment.line];
        let prefix = '';
        let bytes = 0;
        if(line !== undefined) {
            for(let cursor = 0; cursor < line.length && bytes < comment.column; cursor++) {
                const point = line.codePointAt(cursor) ?? 0;
                const character = String.fromCodePoint(point);
                prefix += character;
                bytes += utf8Length(character);
                if(point > 65535) {
                    cursor++;
                }
            }
        }
        if(
            !comment.text.startsWith('//') ||
            directive ||
            line === undefined ||
            bytes !== comment.column ||
            trim(prefix) !== ''
        ) {
            report(context, run, maximum);
            run = [];
            continue;
        }
        const previous = run[run.length - 1];
        if(previous !== undefined && (previous.endLine + 1 !== comment.line || previous.column !== comment.column)) {
            report(context, run, maximum);
            run = [];
        }
        run.push(comment);
    }
    report(context, run, maximum);
}

export function visit(context: RuleContext, index: number): void {
    if(context.node(index).kind === 'SourceFile' && context.enabled('nexus/consistency-no-long-line-comment')) {
        noLongLineComment(context, context.comments);
    }
}
