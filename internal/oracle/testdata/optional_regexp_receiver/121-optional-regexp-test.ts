function pattern(enabled: boolean): RegExp | undefined {
    return enabled ? /entry/ : undefined;
}
function input(): string { console.log("evaluated"); return "entry.js"; }
function excluded(enabled: boolean): boolean { return !!pattern(enabled)?.test(input()); }
console.log(String(excluded(false)));
console.log(String(excluded(true)));
