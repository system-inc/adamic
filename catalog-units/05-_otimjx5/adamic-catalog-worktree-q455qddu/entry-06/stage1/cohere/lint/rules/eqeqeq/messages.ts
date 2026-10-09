export function description(actual: string, expected: string): string {
    return `This compares with \`${actual}\`, which coerces its operands to a common type before comparing. That makes the comparison non-transitive, which is the one property every reader assumes a comparison has: \`0 == '0'\` and \`'0' == []\` disagree while \`0 == []\` is true. Write \`${expected}\`, which compares without coercing.`;
}

export function suggestionDescription(actual: string, expected: string): string {
    return `Use \`${expected}\` instead of \`${actual}\`. This changes what the comparison answers wherever the operands can differ in type, so it is offered rather than applied.`;
}
