"use strict";

const fixed = "EmitAndSemanticDiagnosticsBuilderProgram";
const watchKeys = [
    "onWatchStatusChange", "watchFile", "watchDirectory", "setTimeout",
    "clearTimeout", "preferNonRecursiveWatch",
];

const programPrelude = [
    "const getDefaultLibLocation = memoize(() => getDirectoryPath(normalizePath(system.getExecutingFilePath())));",
    "const getCurrentDirectory = memoize(() => system.getCurrentDirectory());",
    "const realpath = system.realpath?.bind(system);",
    "const getEnvironmentVariable = system.getEnvironmentVariable?.bind(system);",
    "const createHash = system.createHash?.bind(system);",
    "const storeSignatureInfo = system.storeSignatureInfo;",
];
const programFields = [
    "useCaseSensitiveFileNames: () => system.useCaseSensitiveFileNames",
    "getNewLine: () => system.newLine",
    "getCurrentDirectory",
    "getDefaultLibLocation",
    "getDefaultLibFileName: options => combinePaths(getDefaultLibLocation(), getDefaultLibFileName(options))",
    "fileExists: path => system.fileExists(path)",
    "readFile: (path, encoding) => system.readFile(path, encoding)",
    "directoryExists: path => system.directoryExists(path)",
    "getDirectories: path => system.getDirectories(path)",
    "readDirectory: (path, extensions, exclude, include, depth) => system.readDirectory(path, extensions, exclude, include, depth)",
    "realpath",
    "getEnvironmentVariable",
    "trace: s => system.write(s + system.newLine)",
    "createDirectory: path => system.createDirectory(path)",
    "writeFile: (path, data, writeByteOrderMark) => system.writeFile(path, data, writeByteOrderMark)",
    "createHash",
    "createProgram: (rootNames, options, host, oldProgram, configFileParsingDiagnostics, projectReferences) => createEmitAndSemanticDiagnosticsBuilderProgram(rootNames, options, host, oldProgram, configFileParsingDiagnostics, projectReferences)",
    "storeSignatureInfo",
    "now",
];

function watchReads() {
    return [
        "const watcher = createWatchHost(system, reportWatchStatus || createWatchStatusReporter(system));",
        ...watchKeys.map(key => `const ${key} = watcher.${key};`),
    ];
}

const watchCallback = `afterProgramCreate: builderProgram => {
    const compilerOptions = builderProgram.getCompilerOptions();
    const newLine = getNewLineCharacter(compilerOptions);

    emitFilesAndReportErrors(
        builderProgram,
        diagnosticReporter,
        write,
        errorCount =>
            result.onWatchStatusChange!(
                createCompilerDiagnostic(getWatchErrorSummaryDiagnosticMessage(errorCount), errorCount),
                newLine,
                compilerOptions,
                errorCount,
            ),
    );
}`;

function buildFunction(name, parameters, type, prelude, fields, selfReferential) {
    const properties = fields.flatMap(field => field.split("\n").map((line, i, lines) =>
        `        ${line}${i === lines.length - 1 ? "," : ""}`));
    return [
        "/** @internal */",
        `export function ${name}(${parameters}): ${type} {`,
        ...prelude.map(line => `    ${line}`),
        ...(selfReferential ? [`    let result: ${type};`, `    const host: ${type} = {`] : ["    return {"]),
        ...properties,
        "    };",
        ...(selfReferential ? ["    result = host;", "    return host;"] : []),
        "}",
    ].join("\n");
}

const commonWatchInput = [
    "    system: System;",
    "    reportDiagnostic?: DiagnosticReporter | undefined;",
    "    reportWatchStatus?: WatchStatusReporter | undefined;",
];
function watchTwin(config) {
    const kind = config ? "OfConfigFile" : "OfFilesAndCompilerOptions";
    const shape = config ? [
        ["configFileName", "string"],
        ["optionsToExtend?", "CompilerOptions | undefined"],
        ["watchOptionsToExtend?", "WatchOptions | undefined"],
        ["extraFileExtensions?", "readonly FileExtensionInfo[]"],
    ] : [
        ["rootFiles", "string[]"],
        ["options", "CompilerOptions"],
        ["watchOptions", "WatchOptions | undefined"],
        ["projectReferences?", "readonly ProjectReference[]"],
    ];
    const fields = shape.map(([key]) => key.replace("?", ""));
    const members = shape.map(([key, type]) => `    ${key}: ${type};`);
    const params = [
        "{", ...fields.map(field => `    ${field},`),
        "    system: inputSystem,", "    reportDiagnostic,", "    reportWatchStatus,",
        "}: {", ...members, ...commonWatchInput, "}",
    ].join("\n");
    return buildFunction(`createDefaultWatchCompilerHost${kind}`, params, `WatchCompilerHost${kind}<${fixed}>`, [
        "const system = inputSystem;",
        "const diagnosticReporter = reportDiagnostic || createDiagnosticReporter(system);",
        "const write = (s: string) => system.write(s + system.newLine);",
        ...programPrelude,
        "const now = system.now?.bind(system);",
        ...watchReads(),
    ], [
        ...programFields, ...watchKeys, watchCallback,
        ...(config ? ["onUnRecoverableConfigFileDiagnostic: diagnostic => reportUnrecoverableDiagnostic(system, diagnosticReporter, diagnostic)"] : []),
        ...fields,
    ], true);
}

function solutionTwin(watch) {
    const name = watch ? "SolutionBuilderWithWatchHost" : "SolutionBuilderHost";
    const params = [
        "", "    system: System = sys,",
        "    reportDiagnostic?: DiagnosticReporter,",
        "    reportSolutionBuilderStatus?: DiagnosticReporter,",
        watch ? "    reportWatchStatus?: WatchStatusReporter," : "    reportErrorSummary?: ReportEmitErrorSummary,",
        "",
    ].join("\n");
    return buildFunction(`createDefault${name}`, params, `${name}<${fixed}>`, [
        ...programPrelude,
        "system.now?.bind(system);",
        `const getModifiedTime: ${name}<${fixed}>["getModifiedTime"] = system.getModifiedTime ? path => system.getModifiedTime!(path) : returnUndefined;`,
        `const setModifiedTime: ${name}<${fixed}>["setModifiedTime"] = system.setModifiedTime ? (path, date) => system.setModifiedTime!(path, date) : noop;`,
        `const deleteFile: ${name}<${fixed}>["deleteFile"] = system.deleteFile ? path => system.deleteFile!(path) : noop;`,
        "const diagnosticReporter = reportDiagnostic || createDiagnosticReporter(system);",
        "const statusReporter = reportSolutionBuilderStatus || createBuilderStatusReporter(system);",
        "const now = system.now?.bind(system);",
        ...(watch ? watchReads() : []),
    ], [
        ...programFields, "getModifiedTime", "setModifiedTime", "deleteFile",
        "reportDiagnostic: diagnosticReporter", "reportSolutionBuilderStatus: statusReporter",
        ...(watch ? watchKeys : ["reportErrorSummary"]),
    ], false);
}

module.exports = {
    watchKeys,
    twins: { "watch.ts": [watchTwin(true), watchTwin(false)], "tsbuildPublic.ts": [solutionTwin(false), solutionTwin(true)] },
};
