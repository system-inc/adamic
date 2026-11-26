// moduleNameResolver.ts 245:77's shape: { ...resolved, path, originalPath } over Resolved, whose originalPath is
// optional. A source built without it (:140) takes it appended; one built with it undefined (:2774) keeps its place.
interface Resolved {
    path: string;
    extension: string;
    packageId: string | undefined;
    originalPath?: string | true | undefined;
    resolvedUsingTsExtension: boolean | undefined;
}
function built(path: string): Resolved | undefined {
    return { path, extension: ".ts", packageId: undefined, resolvedUsingTsExtension: false };
}
function cached(): Resolved {
    return { path: "/a.ts", extension: ".ts", packageId: undefined, originalPath: undefined, resolvedUsingTsExtension: false };
}
function handle(resolved: Resolved | undefined, resolvedFileName: string, found: string | undefined): Resolved | undefined {
    if (resolved && !resolved.originalPath) {
        const originalPath = found;
        if (originalPath) resolved = { ...resolved, path: resolvedFileName, originalPath };
    }
    return resolved;
}
function show(resolved: Resolved | undefined): void {
    if (resolved) {
        console.log(Object.keys(resolved).join(","));
        console.log(String(JSON.stringify(resolved)));
        console.log(String(resolved.originalPath));
    }
}
show(handle(built("/a.ts"), "/b.ts", "/c.ts"));
show(handle(cached(), "/b.ts", "/c.ts"));
show(handle(built("/a.ts"), "/b.ts", undefined));
