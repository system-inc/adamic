import type { Batch5Context } from './batch5_context.ts';
import { visit as no_dynamic_delete } from './no_dynamic_delete.ts';
import { visit as no_extra_non_null_assertion } from './no_extra_non_null_assertion.ts';
import { visit as no_import_type_side_effects } from './no_import_type_side_effects.ts';
import { visit as no_useless_empty_export } from './no_useless_empty_export.ts';
import { visit as no_explicit_any } from './no_explicit_any.ts';
import { visit as no_duplicate_enum_values } from './no_duplicate_enum_values.ts';
import { visit as no_this_alias } from './no_this_alias.ts';
import { visit as no_misused_new } from './no_misused_new.ts';
import { visit as prefer_as_const } from './prefer_as_const.ts';
import { visit as no_confusing_non_null_assertion } from './no_confusing_non_null_assertion.ts';

export function visitBatch5(context: Batch5Context, index: number): void {
    no_dynamic_delete(context, index);
    no_import_type_side_effects(context, index);
    no_misused_new(context, index);
    no_this_alias(context, index);
    prefer_as_const(context, index);
    no_confusing_non_null_assertion(context, index);
    no_extra_non_null_assertion(context, index);
    no_duplicate_enum_values(context, index);
    no_explicit_any(context, index);
    no_useless_empty_export(context, index);
}
