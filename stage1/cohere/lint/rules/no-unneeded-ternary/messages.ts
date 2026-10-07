export const messageNoUnneededTernaryConditionalExpression =
    'This conditional returns a boolean literal on both branches, so the whole ternary is a long way of writing the condition itself. The branches carry no information the test does not already have, and a reader has to check both to learn that. Write the condition, negated if the branches are the other way round.';
export const messageNoUnneededTernaryConditionalAssignment =
    'This conditional tests a value and then returns that same value, which is the pre-`||` way of spelling a default. `a ? a : b` evaluates `a` twice and reads as though the two branches might differ. Write `a || b`, which says the same thing once.';
