// Each rule owns its implementation; each registration occupies one line.
import type { RuleContext } from './rule_context.ts';
import { visit as no_caller } from './no_caller.ts';
import { visit as no_eq_null } from './no_eq_null.ts';
import { visit as no_empty_static_block } from './no_empty_static_block.ts';
import { visit as no_proto } from './no_proto.ts';
import { visit as no_script_url } from './no_script_url.ts';
import { visit as no_self_compare } from './no_self_compare.ts';
import { visit as no_delete_var } from './no_delete_var.ts';
import { visit as no_iterator } from './no_iterator.ts';
import { visit as no_compare_neg_zero } from './no_compare_neg_zero.ts';
import { visit as no_async_promise_executor } from './no_async_promise_executor.ts';
import { visit as no_empty_pattern } from './no_empty_pattern.ts';
import { visit as no_constructor_return } from './no_constructor_return.ts';
import { visit as no_multi_assign } from './no_multi_assign.ts';
import { visit as no_useless_catch } from './no_useless_catch.ts';
import { visit as no_unsafe_finally } from './no_unsafe_finally.ts';
import { visit as no_useless_concat } from './no_useless_concat.ts';
import { visit as no_empty_character_class } from './no_empty_character_class.ts';
import { visit as no_ex_assign } from './no_ex_assign.ts';
import { visit as no_unnecessary_type_constraint } from './no_unnecessary_type_constraint.ts';
import { visit as prefer_namespace_keyword } from './prefer_namespace_keyword.ts';

export function visitBatch2(ctx: RuleContext, index: number): void {
    no_caller(ctx, index);
    no_eq_null(ctx, index);
    no_empty_static_block(ctx, index);
    no_proto(ctx, index);
    no_script_url(ctx, index);
    no_self_compare(ctx, index);
    no_delete_var(ctx, index);
    no_iterator(ctx, index);
    no_compare_neg_zero(ctx, index);
    no_async_promise_executor(ctx, index);
    no_empty_pattern(ctx, index);
    no_constructor_return(ctx, index);
    no_multi_assign(ctx, index);
    no_useless_catch(ctx, index);
    no_unsafe_finally(ctx, index);
    no_useless_concat(ctx, index);
    no_empty_character_class(ctx, index);
    no_ex_assign(ctx, index);
    no_unnecessary_type_constraint(ctx, index);
    prefer_namespace_keyword(ctx, index);
}
