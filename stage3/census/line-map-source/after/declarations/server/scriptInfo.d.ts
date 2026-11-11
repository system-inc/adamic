import { FormatCodeSettings, IScriptSnapshot, Path, ScriptKind, TextSpan } from "./_namespaces/ts.js";
import { NormalizedPath, Project, ServerHost } from "./_namespaces/ts.server.js";
import * as protocol from "./protocol.js";
export declare function isDynamicFileName(fileName: NormalizedPath): boolean;
export declare class ScriptInfo {
    private readonly host;
    readonly fileName: NormalizedPath;
    readonly scriptKind: ScriptKind;
    readonly hasMixedContent: boolean;
    readonly path: Path;
    /**
     * All projects that include this file
     */
    readonly containingProjects: Project[];
    private formatSettings;
    private preferences;
    /**
     * Set to real path if path is different from info.path
     */
    private realpath;
    constructor(host: ServerHost, fileName: NormalizedPath, scriptKind: ScriptKind, hasMixedContent: boolean, path: Path, initialVersion?: number);
    isScriptOpen(): boolean;
    open(newText: string | undefined): void;
    close(fileExists?: boolean): void;
    getSnapshot(): IScriptSnapshot;
    private ensureRealPath;
    getFormatCodeSettings(): FormatCodeSettings | undefined;
    getPreferences(): protocol.UserPreferences | undefined;
    attachToProject(project: Project): boolean;
    isAttached(project: Project): boolean;
    detachFromProject(project: Project): void;
    detachAllProjects(): void;
    getDefaultProject(): Project;
    registerFileUpdate(): void;
    setOptions(formatSettings: FormatCodeSettings, preferences: protocol.UserPreferences | undefined): void;
    getLatestVersion(): string;
    saveTo(fileName: string): void;
    reloadFromFile(tempFileName?: NormalizedPath): boolean;
    editContent(start: number, end: number, newText: string): void;
    markContainingProjectsAsDirty(): void;
    isOrphan(): boolean;
    /**
     *  @param line 1 based index
     */
    lineToTextSpan(line: number): TextSpan;
    /**
     * @param line 1 based index
     * @param offset 1 based index
     */
    lineOffsetToPosition(line: number, offset: number): number;
    positionToLineOffset(position: number): protocol.Location;
    isJavaScript(): boolean;
}
