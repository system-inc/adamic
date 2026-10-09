// Shape from ahra modules/kling/KlingApi.ts:253, the probe loop finding the next free index.
// `while (true)`, each probe deciding whether there is another, `index` carried across iterations.
declare const NodeFileSystemPromises: {
    access(path: string): Promise<void>;
    mkdir(path: string, options: { recursive: boolean }): Promise<void>;
};
declare const NodePath: { join(...parts: string[]): string };

export async function nextOutputPath(outputDirectory: string): Promise<string> {
    await NodeFileSystemPromises.mkdir(outputDirectory, { recursive: true });

    // Find next available index
    let index = 0;
    while(true) {
        try {
            await NodeFileSystemPromises.access(NodePath.join(outputDirectory, `${index}.mp4`));
            index++;
        }
        catch {
            break;
        }
    }

    return NodePath.join(outputDirectory, `${index}.mp4`);
}
