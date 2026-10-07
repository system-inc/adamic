// messageUnexpectedUnaryOp names the operator used.
export function messageUnexpectedUnaryOp(operator: string): string {
    return `Unary operator \`${operator}\` used. It mutates its operand and evaluates to a different value depending on which side it sits, so \`a[i++]\` and \`a[++i]\` read almost identically and index different elements. Automatic semicolon insertion compounds it: a line ending in a value followed by a line starting with \`++\` joins into one statement. Write the assignment out as \`x += 1\`, which does one thing and reads the same wherever it appears.`;
}
