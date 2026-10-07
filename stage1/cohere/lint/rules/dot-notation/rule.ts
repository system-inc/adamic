import { syntaxKinds as listenerKinds } from './listeners.a';
export const syntaxKinds: readonly number[] = listenerKinds;
import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { comments, type Comment } from '../eslint-comments-require-description/comments.a';
import { Fix, requireDetailedReporting } from '../typescript-no-non-null-assertion/suggestions.a';
import { FixedFinding } from '../arrow-body-style/edits.a';
const name = 'dot-notation';
const keywords: readonly string[] = ['abstract', 'boolean', 'break', 'byte', 'case', 'catch', 'char', 'class', 'const', 'continue', 'debugger', 'default', 'delete', 'do', 'double', 'else', 'enum', 'export', 'extends', 'false', 'final', 'finally', 'float', 'for', 'function', 'goto', 'if', 'implements', 'import', 'in', 'instanceof', 'int', 'interface', 'long', 'native', 'new', 'null', 'package', 'private', 'protected', 'public', 'return', 'short', 'static', 'super', 'switch', 'synchronized', 'this', 'throw', 'throws', 'transient', 'true', 'try', 'typeof', 'var', 'void', 'volatile', 'while', 'with'];
export class Rule {
    readonly context: RuleContext;
    readonly trivia: readonly Comment[];
    constructor(context: RuleContext) {
        this.context = context;
        this.trivia = context.enabled(name) ? comments(context) : [];
        if(context.enabled(name) && context.settings.read('allowpattern', '') !== '') panic('dot-notation: runtime regex option requires shared matcher support');
    }
    between(start: number, end: number): boolean { return this.trivia.some((comment) => comment.pos >= start && comment.end <= end); }
    report(index: number, id: string, message: string, fixes: Fix[]): void {
        this.context.findings.push(new FixedFinding(name, id, message, this.context.start(index), this.context.node(index).end, fixes));
    }
    visit(index: number): void {
        const context = this.context;
        requireDetailedReporting(context, name);
        const node = context.node(index);
        const receiver = node.children[0] ?? -1;
        const argument = node.children[node.children.length - 1] ?? -1;
        if(receiver < 0 || argument < 0) return;
        const allowKeywords = context.settings.read('allowkeywords', 'true') !== 'false';
        if(node.kind === 'PropertyAccessExpression') {
            const key = context.node(argument);
            if(allowKeywords || key.kind !== 'Identifier' || !keywords.includes(key.text)) return;
            const message = `.${key.text} is a syntax error. This project has turned off \`allowKeywords\`, which targets engines old enough that a reserved word after a dot fails to parse at all. The access itself is fine; only the spelling has to change, so write it in brackets.`;
            const from = context.node(receiver).end;
            const optional = context.source.slice(from, key.pos).includes('?.');
            const fixes: Fix[] = [];
            if((optional || context.node(receiver).kind !== 'Identifier' || context.node(receiver).text !== 'let') && !this.between(from, key.end)) {
                const start = optional ? key.pos : from + context.source.slice(from, key.pos).lastIndexOf('.');
                if(start >= from) fixes.push(new Fix(start, key.end, `["${key.text}"]`));
            }
            this.report(argument, 'useBrackets', message, fixes);
            return;
        }
        const value = context.unwrap(argument);
        const key = context.node(value);
        if(!['StringLiteral', 'NoSubstitutionTemplateLiteral', 'TrueKeyword', 'FalseKeyword', 'NullKeyword'].includes(key.kind)) return;
        const text = key.kind === 'TrueKeyword' ? 'true' : key.kind === 'FalseKeyword' ? 'false' : key.kind === 'NullKeyword' ? 'null' : key.text;
        if(!/^[a-zA-Z_$][\w$]*$/u.test(text) || (!allowKeywords && keywords.includes(text))) return;
        const formatted = key.kind === 'StringLiteral' ? `"${text}"` : key.kind === 'NoSubstitutionTemplateLiteral' ? `\`${text}\`` : text;
        const message = `[${formatted}] is better written in dot notation. Bracket notation is for keys computed at runtime; with a literal inside, it is the same access written the long way, and it reads as though the key were dynamic when it is not. Editors also stop offering completion and rename through a bracketed key, so the property quietly drops out of every refactor that touches it.`;
        const from = context.node(receiver).end;
        const opening = context.source.indexOf('[', from);
        const closing = node.end - 1;
        const fixes: Fix[] = [];
        if(opening >= from && closing > opening && context.source.slice(closing, closing + 1) === ']' && !this.between(opening, closing)) {
            const optional = context.source.slice(from, opening).includes('?.');
            const decimal = /^(?:0|0[0-7]*[89][0-9]*|[1-9](?:_?[0-9])*)$/u.test(context.source.slice(context.node(receiver).pos, from).trim());
            let replacement = `${optional ? '' : decimal ? ' .' : '.'}${text}`;
            if(/^[a-zA-Z0-9_$]/u.test(context.source.slice(closing + 1, closing + 2))) replacement += ' ';
            fixes.push(new Fix(opening, closing + 1, replacement));
        }
        this.report(value, 'useDot', message, fixes);
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
