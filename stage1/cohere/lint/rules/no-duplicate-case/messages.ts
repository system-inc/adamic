export const duplicateMessage =
    'Duplicate case label: an earlier arm in this switch tests the same expression, so this one can never run. Almost always a clause was copied and its test never updated, which means the body here is dead and the case it was meant to handle falls through to `default`. Change the test to the value this arm was written for, or delete the arm.';
