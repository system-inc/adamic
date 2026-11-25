interface Resolution { failedLookupLocations?: string[] }
const emptyArray: never[] = [];
function merge(state: Resolution, result: Resolution): void {
    state.failedLookupLocations?.push(...result.failedLookupLocations ?? emptyArray);
    console.log(state.failedLookupLocations?.join(",") ?? "absent");
}
merge({}, {});
merge({}, { failedLookupLocations: ["next"] });
merge({ failedLookupLocations: ["first"] }, {});
merge({ failedLookupLocations: ["first"] }, { failedLookupLocations: ["next"] });
