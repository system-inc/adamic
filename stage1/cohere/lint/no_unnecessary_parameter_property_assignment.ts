import type { RuleContext } from './rule_context.ts';
const messageNoUnnecessaryParameterPropertyAssignment =
    'A constructor parameter carrying `public`, `private`, `protected`, `readonly` or `override` is already assigned to the matching property before the constructor body runs. Writing it again does nothing, and it invites the next reader to believe the two names could diverge. Delete the assignment.';
const suggestionNoUnnecessaryParameterPropertyAssignment = 'Remove the unnecessary assignment.';

function property(ctx: RuleContext, index: number): string {
    const target = ctx.unwrap(index);
    if(target < 0) return '';
    const node = ctx.node(target);
    const receiver = node.children[0] ?? -1;
    if(receiver < 0 || ctx.node(receiver).kind !== 'ThisKeyword') return '';
    if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
        const key = node.children[node.children.length - 1] ?? -1;
        if(
            key >= 0 &&
            ctx.node(key).kind === (node.kind === 'PropertyAccessExpression' ? 'Identifier' : 'StringLiteral')
        )
            return ctx.node(key).text;
    }
    return '';
}
function shadowed(ctx: RuleContext, index: number, name: string): boolean {
    for(let current = ctx.parent(index); current >= 0; current = ctx.parent(current)) {
        const kind = ctx.node(current).kind;
        if(kind === 'Constructor') return false;
        if(['Block', 'SourceFile', 'ModuleBlock'].includes(kind)) {
            for(const statement of ctx.node(current).children) {
                if(ctx.node(statement).kind !== 'VariableStatement') continue;
                const list = ctx.child(statement, 'VariableDeclarationList');
                if(list < 0) continue;
                for(const declaration of ctx.node(list).children) {
                    const binding = ctx.name(declaration);
                    if(binding >= 0 && ctx.node(binding).kind === 'Identifier' && ctx.node(binding).text === name)
                        return true;
                }
            }
        }
        if(ctx.functionLike(current))
            for(const parameter of ctx.node(current).children) {
                if(ctx.node(parameter).kind !== 'Parameter') continue;
                const binding = ctx.name(parameter);
                if(binding >= 0 && ctx.node(binding).kind === 'Identifier' && ctx.node(binding).text === name)
                    return true;
            }
    }
    return false;
}
class AssignmentState {
    readonly parameters: Set<string>;
    readonly before: Set<string>;
    readonly assigned = new Set<string>();
    constructor(parameters: Set<string>, before: Set<string>) {
        this.parameters = parameters;
        this.before = before;
    }
    mark(name: string): void {
        this.assigned.add(name);
    }
    excluded(name: string): boolean {
        return !this.parameters.has(name) || this.assigned.has(name) || this.before.has(name);
    }
}
function walk(ctx: RuleContext, index: number, state: AssignmentState): void {
    const node = ctx.node(index);
    if(ctx.functionLike(index) && node.kind !== 'ArrowFunction') return;
    if(ctx.assignment(index)) {
        const left = node.children[0] ?? -1;
        const operator = node.children[1] ?? -1;
        const right = ctx.outer(node.children[2] ?? -1);
        const name = property(ctx, left);
        if(name === '') return;
        if(
            operator < 0 ||
            ![
                'EqualsToken',
                'BarBarEqualsToken',
                'AmpersandAmpersandEqualsToken',
                'QuestionQuestionEqualsToken',
            ].includes(ctx.node(operator).kind)
        ) {
            state.mark(name);
            return;
        }
        if(
            right < 0 ||
            ctx.node(right).kind !== 'Identifier' ||
            ctx.node(right).text !== name ||
            state.excluded(name) ||
            shadowed(ctx, right, name)
        )
            return;
        const finding = ctx.report(
            index,
            '@typescript-eslint/no-unnecessary-parameter-property-assignment',
            'unnecessaryAssign',
            messageNoUnnecessaryParameterPropertyAssignment,
            'suggestion',
            '',
            suggestionNoUnnecessaryParameterPropertyAssignment,
        );
        finding.editEnd++;
        return;
    }
    for(const child of node.children) walk(ctx, child, state);
}
function initialAssignment(ctx: RuleContext, index: number): string {
    return ctx.assignment(index) ? property(ctx, ctx.node(index).children[0] ?? -1) : '';
}
export function visit(ctx: RuleContext, index: number): void {
    if(
        !ctx.enabled('@typescript-eslint/no-unnecessary-parameter-property-assignment') ||
        ctx.node(index).kind !== 'Constructor'
    )
        return;
    const body = ctx.body(index);
    const owner = ctx.parent(index);
    if(body < 0 || owner < 0) return;
    const parameters = new Set<string>();
    const before = new Set<string>();
    for(const parameter of ctx.node(index).children) {
        if(
            ctx.node(parameter).kind !== 'Parameter' ||
            !ctx
                .node(parameter)
                .children.some((child) =>
                    [
                        'PublicKeyword',
                        'PrivateKeyword',
                        'ProtectedKeyword',
                        'ReadonlyKeyword',
                        'OverrideKeyword',
                    ].includes(ctx.node(child).kind),
                )
        )
            continue;
        const name = ctx.name(parameter);
        if(name >= 0 && ctx.node(name).kind === 'Identifier') parameters.add(ctx.node(name).text);
    }
    if(parameters.size === 0) return;
    for(const member of ctx.node(owner).children) {
        if(ctx.node(member).kind !== 'PropertyDeclaration') continue;
        const initializer = ctx.unwrap(ctx.parts(member)[2] ?? -1);
        if(initializer < 0) continue;
        const node = ctx.node(initializer);
        if(node.kind === 'BinaryExpression') before.add(initialAssignment(ctx, initializer));
        if(node.kind !== 'CallExpression') continue;
        const callee = ctx.unwrap(node.children[0] ?? -1);
        if(callee < 0 || !['ArrowFunction', 'FunctionExpression'].includes(ctx.node(callee).kind)) continue;
        const initializerBody = ctx.body(callee);
        if(initializerBody < 0) continue;
        if(ctx.node(initializerBody).kind !== 'Block') {
            before.add(initialAssignment(ctx, initializerBody));
            continue;
        }
        for(const statement of ctx.node(initializerBody).children)
            if(ctx.node(statement).kind === 'ExpressionStatement')
                before.add(initialAssignment(ctx, ctx.node(statement).children[0] ?? statement));
    }
    walk(ctx, body, new AssignmentState(parameters, before));
}
