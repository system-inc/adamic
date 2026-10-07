export { syntaxKinds } from './listeners.a';
import { panic } from 'adamic';
import { Parser } from '../../../../typescript/parser/parser.ts';
import type { RuleContext } from '../../context.ts';
import { Fix, Suggestion, SuggestedFinding, requireDetailedReporting } from '../typescript-no-non-null-assertion/suggestions.a';

const name = '@typescript-eslint/no-restricted-types';
// This is Go unicode.IsSpace, including NEL and excluding the JavaScript-only BOM.
export function stripped(text: string): string {
    let result = '';
    for(let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        const space = (code >= 9 && code <= 13) || code === 32 || code === 133 || code === 160 || code === 5760 ||
            (code >= 8192 && code <= 8202) || code === 8232 || code === 8233 || code === 8239 || code === 8287 || code === 12288;
        if(!space) result += text.slice(index, index + 1);
    }
    return result;
}
class Ban {
    allowed = false;
    message = '';
    fix = '';
    readonly suggest: string[] = [];
}
function bans(text: string): Map<string, Ban> {
    const result = new Map<string, Ban>();
    if(text === '' || text === 'null') return result;
    const parser = new Parser(`(${text});`, 'restricted type settings');
    const root = parser.file();
    const statement = parser.node(root).children[0] ?? panic('missing settings statement');
    const expression = parser.node(statement).children[0] ?? panic('missing settings expression');
    const object = parser.node(expression).children[0] ?? panic('missing settings object');
    if(parser.node(object).kind !== 'ObjectLiteralExpression') panic('restricted type settings must be an object');
    for(const property of parser.node(object).children) {
        const key = parser.node(property).children[0] ?? panic('missing settings key');
        if(parser.node(key).text.toLowerCase() !== 'types') continue;
        const types = parser.node(property).children[1] ?? panic('missing types');
        if(parser.node(types).kind === 'NullKeyword') continue;
        if(parser.node(types).kind !== 'ObjectLiteralExpression') panic('restricted types must be an object');
        for(const entry of parser.node(types).children) {
            const spelling = parser.node(entry).children[0] ?? panic('missing type key');
            const value = parser.node(entry).children[1] ?? panic('missing type ban');
            const ban = new Ban();
            const node = parser.node(value);
            if(node.kind === 'FalseKeyword' || node.kind === 'NullKeyword') ban.allowed = true;
            else if(node.kind === 'StringLiteral') ban.message = node.text;
            else if(node.kind === 'ObjectLiteralExpression') {
                for(const field of node.children) {
                    const fieldName = parser.node(field).children[0] ?? panic('missing ban key');
                    const fieldValue = parser.node(field).children[1] ?? panic('missing ban value');
                    const label = parser.node(fieldName).text;
                    const item = parser.node(fieldValue);
                    if(label === 'Allowed') ban.allowed = item.kind === 'TrueKeyword';
                    else if(label === 'Message' || label === 'message') ban.message = item.kind === 'NullKeyword' ? '' : item.text;
                    else if(label === 'FixWith' || label === 'fixWith') ban.fix = item.kind === 'NullKeyword' ? '' : item.text;
                    else if(label === 'Suggest' || label === 'suggest') {
                        if(item.kind === 'NullKeyword') continue;
                        if(item.kind !== 'ArrayLiteralExpression') panic('restricted type suggestions must be strings');
                        for(const replacement of item.children) {
                            if(parser.node(replacement).kind !== 'StringLiteral') panic('restricted type suggestion must be a string');
                            ban.suggest.push(parser.node(replacement).text);
                        }
                    }
                    else panic('unknown restricted type ban field');
                }
            }
            else if(node.kind !== 'TrueKeyword') continue;
            result.set(stripped(parser.node(spelling).text), ban);
        }
    }
    return result;
}
export class Rule {
    readonly context: RuleContext;
    readonly bans: Map<string, Ban>;
    constructor(context: RuleContext) { this.context = context; this.bans = context.enabled(name) ? bans(context.settings.text) : new Map<string, Ban>(); }
    check(index: number, spelling: string): void {
        const context = this.context;
        const ban = this.bans.get(spelling);
        if(ban === undefined || ban.allowed) return;
        const start = context.start(index);
        const end = context.node(index).end;
        const message = "Don't use `" + spelling + '` as a type.' + (ban.message === '' ? '' : ' ' + ban.message);
        const suggestions: Suggestion[] = [];
        for(const replacement of ban.suggest) suggestions.push(new Suggestion('bannedTypeReplacement',
            'Replace `' + spelling + '` with `' + replacement + '`.', [new Fix(start, end, replacement)]));
        context.findings.push(new SuggestedFinding(name, 'bannedTypeMessage', message, start, end, suggestions,
            ban.fix === '' ? '' : 'fix', ban.fix));
    }
    checkNode(index: number): void {
        this.check(index, stripped(this.context.source.slice(this.context.start(index), this.context.node(index).end)));
    }
    visit(index: number): void {
        const node = this.context.node(index);
        if(node.kind === 'TypeReference') {
            const typeName = node.children[0] ?? -1;
            if(typeName >= 0) this.checkNode(typeName);
            if(node.children.length > 1 || this.context.source.slice(this.context.start(index), node.end).includes('<')) this.checkNode(index);
        }
        else if(node.kind === 'TupleType' || node.kind === 'TypeLiteral') {
            if(node.children.length === 0) this.checkNode(index);
        }
        else this.check(index, node.kind.slice(0, -7).toLowerCase());
    }
}
export function create(context: RuleContext): Rule {
    requireDetailedReporting(context, name);
    return new Rule(context);
}
