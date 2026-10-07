export const blockMessage =
    'This block is empty, which leaves a reader unable to tell a deliberate no-op from a body that was deleted or never written. The distinction matters most in a `catch`, where an empty block is the difference between swallowing an error on purpose and swallowing it by accident. Write a comment saying why nothing happens here, or remove the block.';

export const switchMessage =
    'This switch has no clauses at all, so it evaluates its subject and does nothing with the result. Either the arms were removed and the statement outlived them, or they were never written. Add the clauses, or remove the switch and keep whatever the subject expression was doing.';
