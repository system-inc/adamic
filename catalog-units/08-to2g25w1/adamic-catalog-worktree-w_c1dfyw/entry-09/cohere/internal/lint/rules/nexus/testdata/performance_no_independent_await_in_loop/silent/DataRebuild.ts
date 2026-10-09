// Shape from ahra modules/data/DataRebuild.ts:589 (applying trimmed-row archives).
// Silent for three independent reasons, each also pinned alone by a synthetic case: a progress
// callback before and after the awaits, `totalRowsInserted +=` accumulating across iterations, and
// `results.length` naming each archive's scratch directory, which reads the collection the loop
// fills.
interface DataRebuildTrimmedRowsResultInterface {
    trimmedArchive: string;
    table: string;
    rowsParsed: number;
    rowsInserted: number;
}
type DataRebuildProgressCallbackType = (progress: { phase: string; message: string; trimmedArchivesApplied: number }) => void;
declare function discoverTrimmedRows(directory: string, database: string): Array<{ archivePath: string; checksumPath: string }>;
declare function verifyChecksumOrThrow(archivePath: string, checksumPath: string): Promise<void>;
declare function decompressArchive(archivePath: string, destination: string): Promise<void>;
declare function soleDumpFolder(directory: string): string;
declare function applyTrimmedRowsDumpFolder(
    databasePath: string,
    folder: string,
    database: string,
): Array<{ table: string; rowsParsed: number; rowsInserted: number }>;
declare const NodeFileSystem: { rmSync(path: string, options: { recursive: boolean; force: boolean }): void };
declare const NodePath: { basename(path: string): string };

export async function applyTrimmedRows(
    database: string,
    databasePath: string,
    scratchPrefix: string,
    planetScaleDirectory: string,
    onProgress?: DataRebuildProgressCallbackType,
): Promise<{ results: DataRebuildTrimmedRowsResultInterface[]; totalRowsInserted: number }> {
    const trimmedArchives = discoverTrimmedRows(planetScaleDirectory, database);
    const results: DataRebuildTrimmedRowsResultInterface[] = [];
    let totalRowsInserted = 0;

    for(const trimmed of trimmedArchives) {
        onProgress?.({
            phase: 'ApplyingTrimmedRows',
            message: `Verifying + applying ${NodePath.basename(trimmed.archivePath)}`,
            trimmedArchivesApplied: results.length,
        });
        await verifyChecksumOrThrow(trimmed.archivePath, trimmed.checksumPath);

        const trimmedScratch = `${scratchPrefix}-${results.length}`;
        NodeFileSystem.rmSync(trimmedScratch, { recursive: true, force: true });
        await decompressArchive(trimmed.archivePath, trimmedScratch);
        const trimmedDumpFolder = soleDumpFolder(trimmedScratch);
        const applied = applyTrimmedRowsDumpFolder(databasePath, trimmedDumpFolder, database);
        NodeFileSystem.rmSync(trimmedScratch, { recursive: true, force: true });

        for(const perTable of applied) {
            totalRowsInserted += perTable.rowsInserted;
            results.push({
                trimmedArchive: NodePath.basename(trimmed.archivePath),
                table: perTable.table,
                rowsParsed: perTable.rowsParsed,
                rowsInserted: perTable.rowsInserted,
            });
            onProgress?.({
                phase: 'ApplyingTrimmedRows',
                message: `${NodePath.basename(trimmed.archivePath)} -> ${perTable.table}`,
                trimmedArchivesApplied: results.length,
            });
        }
    }
    return { results, totalRowsInserted };
}
