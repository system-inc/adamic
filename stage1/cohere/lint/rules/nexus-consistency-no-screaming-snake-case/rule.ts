import type { RuleContext } from '../../context.ts';
import { message } from './messages.ts';
export function screaming(name: string): boolean {
    if(!name.includes('_') || name.endsWith('_')) { return false; }
    for(let index = 0; index < name.length; index++) {
        const character = name.slice(index, index + 1);
        const upper = character >= 'A' && character <= 'Z';
        if(index === 0 && !upper) { return false; }
        if(!upper && !(character >= '0' && character <= '9') && character !== '_') { return false; }
    }
    return true;
}
export function renamed(name: string, exported: boolean): string {
    const parts = name.split('_').filter(part => part !== '');
    let result = '';
    for(let index = 0; index < parts.length; index++) {
        const part = (parts[index] ?? '').toLowerCase();
        result += index === 0 && !exported ? part : part.slice(0, 1).toUpperCase() + part.slice(1);
    }
    return result;
}
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    environment(index: number): boolean {
        let cursor = index;
        const node = this.context.node(cursor);
        if(node.kind === 'BinaryExpression') {
            const operator = node.children[1] ?? -1;
            if(operator >= 0 && ['QuestionQuestionToken', 'BarBarToken'].includes(this.context.node(operator).kind)) { cursor = node.children[0] ?? -1; }
        }
        if(cursor < 0 || this.context.node(cursor).kind !== 'PropertyAccessExpression') { return false; }
        let sawEnvironment = false;
        while(cursor >= 0 && this.context.node(cursor).kind === 'PropertyAccessExpression') {
            const children = this.context.node(cursor).children;
            const property = children[children.length - 1] ?? -1;
            if(property >= 0 && this.context.node(property).text === 'env') { sawEnvironment = true; }
            cursor = children[0] ?? -1;
        }
        return cursor >= 0 && this.context.node(cursor).kind === 'Identifier' && this.context.node(cursor).text === 'process' && sawEnvironment;
    }
    visit(index: number, parent: number): void {
        const children = this.context.node(index).children;
        const identifier = children[0] ?? -1;
        if(identifier < 0 || this.context.node(identifier).kind !== 'Identifier') { return; }
        const name = this.context.node(identifier).text;
        if(!screaming(name) || this.context.settings.list('allow', []).includes(name)) { return; }
        if(parent < 0 || this.context.node(parent).kind !== 'VariableDeclarationList') { return; }
        this.context.scanner.pos = this.context.start(parent);
        if(this.context.scanner.scan() !== 'ConstKeyword') { return; }
        const initializer = children[children.length - 1] ?? -1;
        if(initializer !== identifier && initializer >= 0) {
            const between = this.context.source.slice(this.context.node(identifier).end, this.context.start(initializer));
            if(between.includes('=') && this.environment(initializer)) { return; }
        }
        const statement = this.context.parents[parent] ?? -1;
        const exported = statement >= 0 && this.context.node(statement).children.some(child => this.context.node(child).kind === 'ExportKeyword');
        const id = exported ? 'noScreamingSnakeCaseExported' : 'noScreamingSnakeCaseLocal';
        this.context.report(identifier, 'nexus/consistency-no-screaming-snake-case', id, message(name, renamed(name, exported), exported), '', '', '');
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
