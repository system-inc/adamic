import { panic } from 'adamic';
import { space } from './comments.ts';
import type { Batch4Context } from './batch4_context.ts';

function message(name: string): string {
    return `This file pulls in \`${name}\` with a triple-slash reference directive. A directive is a global instruction to the compiler with no binding and no local name, so nothing in the file records that the dependency is used here and no tool can follow it the way it follows an import. Write an \`import\` instead, which names what it brings in and participates in module resolution.`;
}
function fields(text: string): string[] {
    const result: string[] = [];
    let current = '';
    for(const character of text) {
        if(space(character)) {
            if(current !== '') result.push(current);
            current = '';
        }
        else current += character;
    }
    if(current !== '') result.push(current);
    return result;
}
export function visit(ctx: Batch4Context, index: number): void {
    if(!['.ts', '.tsx', '.mts', '.cts'].some((suffix) => ctx.parser.path.endsWith(suffix))) return;
    if(!ctx.enabled('@typescript-eslint/triple-slash-reference') || ctx.node(index).kind !== 'SourceFile') return;
    const first = ctx.node(index).children[0] ?? -1;
    const cutoff = first < 0 ? ctx.source.length : ctx.start(first);
    const awaiting = new Map<string, number>();
    for(let cursor = 0; cursor < ctx.commentStarts.length; cursor++) {
        const start = ctx.commentStarts[cursor] ?? panic('comment');
        const end = ctx.commentEnds[cursor] ?? panic('comment');
        if(start >= cutoff) break;
        const content = ctx.source.slice(start + 2, end);
        if(!content.startsWith('/')) continue;
        const open = content.indexOf('<reference ');
        const close = content.indexOf('/>', open);
        if(open < 0 || close < 0) continue;
        for(const part of fields(content.slice(open + 11, close))) {
            if(!['types=', 'path=', 'lib='].some((prefix) => part.startsWith(prefix))) continue;
            const halves = part.split('=');
            if(halves.length !== 2) break;
            const key = (halves[0] ?? '').trim().replaceAll('"', '');
            let value = halves[1] ?? '';
            while(value.startsWith('"')) value = value.slice(1);
            while(value.endsWith('"')) value = value.slice(0, -1);
            while(value.endsWith('/')) value = value.slice(0, -1);
            const option = ctx.settings.read(
                key,
                key === 'types' ? 'prefer-import' : key === 'path' ? 'never' : 'always',
            );
            if(option === 'never') {
                const finding = ctx.report(
                    index,
                    '@typescript-eslint/triple-slash-reference',
                    'tripleSlashReference',
                    message(value),
                );
                ctx.replaceRange(finding, start, end);
            }
            if(key === 'types' && option === 'prefer-import') awaiting.set(value, cursor);
            break;
        }
    }
    for(const statement of ctx.node(index).children) {
        const kind = ctx.node(statement).kind;
        if(kind !== 'ImportDeclaration' && kind !== 'ImportEqualsDeclaration') continue;
        let specifier = ctx.child(statement, 'StringLiteral');
        const external = ctx.child(statement, 'ExternalModuleReference');
        if(specifier < 0 && external >= 0) specifier = ctx.child(external, 'StringLiteral');
        if(specifier < 0) continue;
        const name = ctx.node(specifier).text;
        const comment = awaiting.get(name);
        if(comment !== undefined) {
            const finding = ctx.report(
                index,
                '@typescript-eslint/triple-slash-reference',
                'tripleSlashReference',
                message(name),
            );
            ctx.replaceRange(finding, ctx.commentStarts[comment] ?? 0, ctx.commentEnds[comment] ?? 0);
        }
    }
}
