// Each rule owns its dispatch and implementation; registration is one line.
import type { RuleContext } from './context.ts';
import { ancestry } from './shared.ts';
import { visit as oneVar } from './one_var.ts';
import { visit as noAmbiguousIdentifier } from './no_ambiguous_identifier.ts';
import { visit as noAbbreviatedIdentifier } from './no_abbreviated_identifier.ts';
import { visit as noNonNullAssertion } from './no_non_null_assertion.ts';
import { visit as preferEnumInitializers } from './prefer_enum_initializers.ts';
import { visit as noSingleLineJsdoc } from './no_single_line_jsdoc.ts';
import { visit as noLongLineComment } from './no_long_line_comment.ts';
import { visit as noShouting } from './no_shouting.ts';
import { visit as noMultilineArrowFunction } from './no_multiline_arrow_function.ts';
import { visit as preferDestructuring } from './prefer_destructuring.ts';

export function visitRegistered(context: RuleContext, index: number): void {
    if(context.node(index).kind === 'SourceFile') {
        ancestry(context, index, -1);
    }
    oneVar(context, index);
    noAmbiguousIdentifier(context, index);
    noAbbreviatedIdentifier(context, index);
    noNonNullAssertion(context, index);
    preferEnumInitializers(context, index);
    noSingleLineJsdoc(context, index);
    noLongLineComment(context, index);
    noShouting(context, index);
    noMultilineArrowFunction(context, index);
    preferDestructuring(context, index);
}
