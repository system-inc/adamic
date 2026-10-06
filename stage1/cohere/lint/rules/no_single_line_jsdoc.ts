import type { RuleContext } from './context.ts';
import type { CommentRange } from './comment_ranges.ts';
import { Finding } from '../finding.ts';
import { policyMessage } from '../volume_messages.ts';
import { trim } from './shared.ts';

export function noSingleLineJsdoc(context: RuleContext, ranges: readonly CommentRange[]): void {
    const rule = 'nexus/consistency-no-single-line-jsdoc';
    for(const comment of ranges) {
        if(!comment.text.startsWith('/**')) {
            continue;
        }
        let body = comment.text.slice(3);
        if(body.endsWith('*/')) {
            body = body.slice(0, -2);
        }
        const lines: string[] = [];
        for(const line of body.split('\n')) {
            let text = trim(line);
            if(text.startsWith('*')) {
                text = trim(text.slice(1));
            }
            if(text !== '') {
                lines.push(text);
            }
        }
        const description = lines[0] ?? '';
        if(lines.length !== 1 || description.startsWith('@')) {
            continue;
        }
        let codeFollows = false;
        for(let cursor = comment.end; cursor < context.source.length; cursor++) {
            const character = context.source[cursor] ?? '';
            if(character === '\n' || character === '\r') {
                break;
            }
            if(character !== ' ' && character !== '\t') {
                codeFollows = true;
                break;
            }
        }
        const fixes = !description.includes('//') && comment.line === comment.endLine && !codeFollows;
        context.record(
            new Finding(
                rule,
                'useSimpleComment',
                policyMessage(rule, 'useSimpleComment', []),
                comment.start,
                comment.end,
                fixes ? 'fix' : '',
                fixes ? `// ${description}` : '',
                '',
            ),
        );
    }
}

export function visit(context: RuleContext, index: number): void {
    if(context.node(index).kind === 'SourceFile' && context.enabled('nexus/consistency-no-single-line-jsdoc')) {
        noSingleLineJsdoc(context, context.comments);
    }
}
