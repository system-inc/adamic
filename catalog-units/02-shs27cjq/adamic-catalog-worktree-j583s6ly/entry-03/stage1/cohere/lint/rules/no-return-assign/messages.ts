export const messageReturnAssignment =
    'This return statement assigns rather than returns a value it already has, so it does two things at once and the assignment is the one the reader is likeliest to miss. It is usually a `==` typed as `=`, and where it is deliberate it reads as a value being returned rather than as a write happening on the way out. Assign on its own line and return the variable, or wrap the assignment in its own parentheses to say the write was meant.';

export const messageArrowAssignment =
    "This arrow function's body is an assignment, so calling it writes to something outside itself and returns what it wrote. A concise arrow body reads as a value, which is what makes the write easy to miss at the call site. Give the arrow a block body and assign inside it, or wrap the assignment in its own parentheses to say the write was meant.";
