import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
import { space } from '../../comments.ts';
import { comments, type Comment } from './comments.a';
import { description } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    report(comment: Comment, value: string): void {
        this.context.findings.push(new Finding('@typescript-eslint/triple-slash-reference', 'tripleSlashReference', description(value), comment.start, comment.end, '', '', ''));
    }
    visit(index: number, parent: number): void {
        if(!['.ts', '.tsx', '.mts', '.cts'].some(extension => this.context.parser.path.endsWith(extension))) { return; }
        const statements = this.context.node(index).children;
        const cutoff = statements.length === 0 ? this.context.node(index).end : this.context.start(statements[0] ?? panic('missing statement'));
        const pending = new Map<string, Comment>();
        for(const comment of comments(this.context)) {
            if(comment.start >= cutoff) { break; }
            const content = comment.text.slice(2);
            if(!content.startsWith('/')) { continue; }
            const opening = content.indexOf('<reference ');
            if(opening < 0) { continue; }
            const end = content.indexOf('/>', opening);
            if(end < 0) { continue; }
            const inner = content.slice(opening + '<reference '.length, end);
            const fields: string[] = [];
            let start = 0;
            for(let cursor = 0; cursor <= inner.length; cursor++) {
                if(cursor === inner.length || space(inner[cursor] ?? '')) {
                    if(cursor > start) { fields.push(inner.slice(start, cursor)); }
                    start = cursor + 1;
                }
            }
            for(const field of fields) {
                if(!['types=', 'path=', 'lib='].some(prefix => field.startsWith(prefix))) { continue; }
                const parts = field.split('=');
                if(parts.length !== 2) { break; }
                const key = parts[0] ?? panic('missing key');
                const value = (parts[1] ?? '').replace(/^"+|"+$/g, '').replace(/\/+$/g, '');
                const fallback = key === 'types' ? 'prefer-import' : key === 'path' ? 'never' : 'always';
                const setting = this.context.settings.read(key, fallback);
                if(setting === 'never') { this.report(comment, value); }
                if(key === 'types' && setting === 'prefer-import') { pending.set(value, comment); }
                break;
            }
        }
        for(const statement of statements) {
            const node = this.context.node(statement);
            let specifier = -1;
            if(node.kind === 'ImportDeclaration') {
                for(const child of node.children) { if(['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(this.context.node(child).kind)) { specifier = child; } }
            }
            if(node.kind === 'ImportEqualsDeclaration') {
                for(const child of node.children) {
                    if(this.context.node(child).kind === 'ExternalModuleReference') { specifier = this.context.node(child).children[0] ?? -1; }
                }
            }
            if(specifier < 0 || !['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(this.context.node(specifier).kind)) { continue; }
            const value = this.context.node(specifier).text;
            const comment = pending.get(value);
            if(comment !== undefined) { this.report(comment, value); }
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
