import type { RuleContext } from './rule_context.ts';
import { visit as rule1 } from './no_assign_module_variable.ts';
import { visit as rule2 } from './default_param_last.ts';
import { visit as rule3 } from './no_confusing_non_null_assertion.ts';
import { visit as rule4 } from './no_duplicate_enum_values.ts';
import { visit as rule5 } from './no_dynamic_delete.ts';
import { visit as rule6 } from './no_extra_non_null_assertion.ts';
import { visit as rule7 } from './no_misused_new.ts';
import { visit as rule8 } from './no_unnecessary_parameter_property_assignment.ts';
import { visit as rule9 } from './prefer_as_const.ts';
import { visit as rule10 } from './default_case_last.ts';
export function visitRules(ctx: RuleContext, index: number): void {
    rule1(ctx, index);
    rule2(ctx, index);
    rule3(ctx, index);
    rule4(ctx, index);
    rule5(ctx, index);
    rule6(ctx, index);
    rule7(ctx, index);
    rule8(ctx, index);
    rule9(ctx, index);
    rule10(ctx, index);
}
