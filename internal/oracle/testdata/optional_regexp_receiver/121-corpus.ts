function toRegex(name: string): RegExp | undefined {
    if (name === "matching") return /entry/;
    if (name === "other") return /else/;
    return undefined;
}
function excluded(patterns: string[], specifier: string): boolean {
    return patterns.some(pattern => !!toRegex(pattern)?.test(specifier));
}
console.log(String(excluded(["invalid", "other", "matching"], "entry.js")));
console.log(String(excluded(["invalid", "other"], "entry.js")));
