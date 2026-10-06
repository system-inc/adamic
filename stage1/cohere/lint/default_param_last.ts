import { panic } from 'adamic';
import type { RuleContext } from './rule_context.ts';
const messageDefaultParamLastShouldBeLast =
    'A parameter with a default sits before a parameter without one, so the default can never be taken without passing `undefined` explicitly at every call site. Move it after the required parameters, where omitting it is what selects the default.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('@typescript-eslint/default-param-last') || !ctx.functionLike(index) || ctx.body(index) < 0) return;
    const parameters = ctx.node(index).children.filter((child) => ctx.node(child).kind === 'Parameter');
    let last = -1;
    for(let cursor = 0; cursor < parameters.length; cursor++) {
        const parameter = parameters[cursor] ?? panic('parameter');
        if(
            (ctx.parts(parameter)[2] ?? -1) < 0 &&
            !ctx.hasKind(parameter, 'QuestionToken') &&
            !ctx.hasKind(parameter, 'DotDotDotToken')
        )
            last = cursor;
    }
    for(let cursor = 0; cursor < last; cursor++) {
        const parameter = parameters[cursor] ?? panic('parameter');
        if((ctx.parts(parameter)[2] ?? -1) >= 0 || ctx.hasKind(parameter, 'QuestionToken'))
            ctx.report(
                parameter,
                '@typescript-eslint/default-param-last',
                'shouldBeLast',
                messageDefaultParamLastShouldBeLast,
            );
    }
}
