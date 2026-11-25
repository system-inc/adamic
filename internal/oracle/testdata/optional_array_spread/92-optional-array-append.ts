interface Resolution { failedLookupLocations?: string[] }
function incoming(): string[] { console.log("evaluated"); return ["next"]; }
function merge(state: Resolution): void {
    state.failedLookupLocations?.push(...incoming());
}
merge({});
const state: Resolution = { failedLookupLocations: ["first"] };
merge(state);
console.log(state.failedLookupLocations?.join(",") ?? "");
