import { combinePaths, Debug, getLineAndCharacterOfPosition, getSourceFileOfNode, ObjectFlags, timestamp, TypeFlags, unescapeLeadingUnderscores, } from "./_namespaces/ts.js";
import * as performance from "./_namespaces/ts.performance.js";
import { errorMessage } from "./hostErrors.js";
/* Tracing events for the compiler. */
// should be used as tracing?.___
/** @internal */
export let tracing;
// enable the above using startTracing()
/**
 * Do not use this directly; instead @see {tracing}.
 * @internal
 */
export var tracingEnabled;
(function (tracingEnabled) {
    let fs;
    let traceCount = 0;
    let traceFd = 0;
    let mode;
    const typeCatalog = []; // NB: id is index + 1
    let legendPath;
    const legend = [];
    /** Starts tracing for the given project. */
    function startTracing(tracingMode, traceDir, configFilePath) {
        Debug.assert(!tracing, "Tracing already started");
        if (fs === undefined) {
            try {
                fs = require("fs");
            }
            catch (e) {
                throw new Error(`tracing requires having fs\n(original error: ${errorMessage(e) || e})`);
            }
        }
        mode = tracingMode;
        typeCatalog.length = 0;
        if (legendPath === undefined) {
            legendPath = combinePaths(traceDir, "legend.json");
        }
        // Note that writing will fail later on if it exists and is not a directory
        if (!fs.existsSync(traceDir)) {
            fs.mkdirSync(traceDir, { recursive: true });
        }
        const countPart = mode === "build" ? `.${process.pid}-${++traceCount}`
            : mode === "server" ? `.${process.pid}`
                : ``;
        const tracePath = combinePaths(traceDir, `trace${countPart}.json`);
        const typesPath = combinePaths(traceDir, `types${countPart}.json`);
        legend.push({
            configFilePath,
            tracePath,
            typesPath,
        });
        traceFd = fs.openSync(tracePath, "w");
        tracing = tracingEnabled; // only when traceFd is properly set
        // Start with a prefix that contains some metadata that the devtools profiler expects (also avoids a warning on import)
        const meta = { cat: "__metadata", ph: "M", ts: 1000 * timestamp(), pid: 1, tid: 1 };
        fs.writeSync(traceFd, "[\n"
            + [{ name: "process_name", args: { name: "tsc" }, ...meta }, { name: "thread_name", args: { name: "Main" }, ...meta }, { name: "TracingStartedInBrowser", ...meta, cat: "disabled-by-default-devtools.timeline" }]
                .map(v => JSON.stringify(v)).join(",\n"));
    }
    tracingEnabled.startTracing = startTracing;
    /** Stops tracing for the in-progress project and dumps the type catalog. */
    function stopTracing() {
        Debug.assert(!!tracing, "Tracing is not in progress");
        Debug.assert(!!typeCatalog.length === (mode !== "server")); // Have a type catalog iff not in server mode
        fs.writeSync(traceFd, `\n]\n`);
        fs.closeSync(traceFd);
        tracing = undefined;
        if (typeCatalog.length) {
            dumpTypes(typeCatalog);
        }
        else {
            // We pre-computed this path for convenience, but clear it
            // now that the file won't be created.
            legend[legend.length - 1].typesPath = undefined;
        }
    }
    tracingEnabled.stopTracing = stopTracing;
    function recordType(type) {
        if (mode !== "server") {
            typeCatalog.push(type);
        }
    }
    tracingEnabled.recordType = recordType;
    let Phase;
    (function (Phase) {
        Phase["Parse"] = "parse";
        Phase["Program"] = "program";
        Phase["Bind"] = "bind";
        Phase["Check"] = "check";
        Phase["CheckTypes"] = "checkTypes";
        Phase["Emit"] = "emit";
        Phase["Session"] = "session";
    })(Phase = tracingEnabled.Phase || (tracingEnabled.Phase = {}));
    function instant(phase, name, args) {
        writeEvent("I", phase, name, args, `"s":"g"`);
    }
    tracingEnabled.instant = instant;
    const eventStack = [];
    /**
     * @param separateBeginAndEnd - used for special cases where we need the trace point even if the event
     * never terminates (typically for reducing a scenario too big to trace to one that can be completed).
     * In the future we might implement an exit handler to dump unfinished events which would deprecate
     * these operations.
     */
    function push(phase, name, args, separateBeginAndEnd = false) {
        if (separateBeginAndEnd) {
            writeEvent("B", phase, name, args);
        }
        eventStack.push({ phase, name, args, time: 1000 * timestamp(), separateBeginAndEnd });
    }
    tracingEnabled.push = push;
    function pop(results) {
        Debug.assert(eventStack.length > 0);
        writeStackEvent(eventStack.length - 1, 1000 * timestamp(), results);
        eventStack.length--;
    }
    tracingEnabled.pop = pop;
    function popAll() {
        const endTime = 1000 * timestamp();
        for (let i = eventStack.length - 1; i >= 0; i--) {
            writeStackEvent(i, endTime);
        }
        eventStack.length = 0;
    }
    tracingEnabled.popAll = popAll;
    // sample every 10ms
    const sampleInterval = 1000 * 10;
    function writeStackEvent(index, endTime, results) {
        const { phase, name, args, time, separateBeginAndEnd } = eventStack[index];
        if (separateBeginAndEnd) {
            Debug.assert(!results, "`results` are not supported for events with `separateBeginAndEnd`");
            writeEvent("E", phase, name, args, /*extras*/ undefined, endTime);
        }
        // test if [time,endTime) straddles a sampling point
        else if (sampleInterval - (time % sampleInterval) <= endTime - time) {
            writeEvent("X", phase, name, { ...args, results }, `"dur":${endTime - time}`, time);
        }
    }
    function writeEvent(eventType, phase, name, args, extras, time = 1000 * timestamp()) {
        // In server mode, there's no easy way to dump type information, so we drop events that would require it.
        if (mode === "server" && phase === Phase.CheckTypes)
            return;
        performance.mark("beginTracing");
        fs.writeSync(traceFd, `,\n{"pid":1,"tid":1,"ph":"${eventType}","cat":"${phase}","ts":${time},"name":"${name}"`);
        if (extras)
            fs.writeSync(traceFd, `,${extras}`);
        if (args)
            fs.writeSync(traceFd, `,"args":${JSON.stringify(args)}`);
        fs.writeSync(traceFd, `}`);
        performance.mark("endTracing");
        performance.measure("Tracing", "beginTracing", "endTracing");
    }
    function getLocation(node) {
        const file = getSourceFileOfNode(node);
        return !file
            ? undefined
            : {
                path: file.path,
                start: indexFromOne(getLineAndCharacterOfPosition(file, node.pos)),
                end: indexFromOne(getLineAndCharacterOfPosition(file, node.end)),
            };
        function indexFromOne(lc) {
            return {
                line: lc.line + 1,
                character: lc.character + 1,
            };
        }
    }
    function dumpTypes(types) {
        performance.mark("beginDumpTypes");
        const typesPath = legend[legend.length - 1].typesPath;
        const typesFd = fs.openSync(typesPath, "w");
        const recursionIdentityMap = new Map();
        // Cleverness: no line break here so that the type ID will match the line number
        fs.writeSync(typesFd, "[");
        const numTypes = types.length;
        for (let i = 0; i < numTypes; i++) {
            const type = types[i];
            const objectFlags = type.objectFlags;
            const symbol = type.aliasSymbol ?? type.symbol;
            // It's slow to compute the display text, so skip it unless it's really valuable (or cheap)
            let display;
            if ((objectFlags & ObjectFlags.Anonymous) | (type.flags & TypeFlags.Literal)) {
                try {
                    display = type.checker?.typeToString(type);
                }
                catch {
                    display = undefined;
                }
            }
            let indexedAccessProperties = {};
            if (type.flags & TypeFlags.IndexedAccess) {
                const indexedAccessType = type;
                indexedAccessProperties = {
                    indexedAccessObjectType: indexedAccessType.objectType?.id,
                    indexedAccessIndexType: indexedAccessType.indexType?.id,
                };
            }
            let referenceProperties = {};
            if (objectFlags & ObjectFlags.Reference) {
                const referenceType = type;
                referenceProperties = {
                    instantiatedType: referenceType.target?.id,
                    typeArguments: referenceType.resolvedTypeArguments?.map(t => t.id),
                    referenceLocation: getLocation(referenceType.node),
                };
            }
            let conditionalProperties = {};
            if (type.flags & TypeFlags.Conditional) {
                const conditionalType = type;
                conditionalProperties = {
                    conditionalCheckType: conditionalType.checkType?.id,
                    conditionalExtendsType: conditionalType.extendsType?.id,
                    conditionalTrueType: conditionalType.resolvedTrueType?.id ?? -1,
                    conditionalFalseType: conditionalType.resolvedFalseType?.id ?? -1,
                };
            }
            let substitutionProperties = {};
            if (type.flags & TypeFlags.Substitution) {
                const substitutionType = type;
                substitutionProperties = {
                    substitutionBaseType: substitutionType.baseType?.id,
                    constraintType: substitutionType.constraint?.id,
                };
            }
            let reverseMappedProperties = {};
            if (objectFlags & ObjectFlags.ReverseMapped) {
                const reverseMappedType = type;
                reverseMappedProperties = {
                    reverseMappedSourceType: reverseMappedType.source?.id,
                    reverseMappedMappedType: reverseMappedType.mappedType?.id,
                    reverseMappedConstraintType: reverseMappedType.constraintType?.id,
                };
            }
            let evolvingArrayProperties = {};
            if (objectFlags & ObjectFlags.EvolvingArray) {
                const evolvingArrayType = type;
                evolvingArrayProperties = {
                    evolvingArrayElementType: evolvingArrayType.elementType.id,
                    evolvingArrayFinalType: evolvingArrayType.finalArrayType?.id,
                };
            }
            // We can't print out an arbitrary object, so just assign each one a unique number.
            // Don't call it an "id" so people don't treat it as a type id.
            let recursionToken;
            const recursionIdentity = type.checker.getRecursionIdentity(type);
            if (recursionIdentity) {
                recursionToken = recursionIdentityMap.get(recursionIdentity);
                if (!recursionToken) {
                    recursionToken = recursionIdentityMap.size;
                    recursionIdentityMap.set(recursionIdentity, recursionToken);
                }
            }
            const descriptor = {
                id: type.id,
                intrinsicName: type.intrinsicName,
                symbolName: symbol?.escapedName && unescapeLeadingUnderscores(symbol.escapedName),
                recursionId: recursionToken,
                isTuple: objectFlags & ObjectFlags.Tuple ? true : undefined,
                unionTypes: (type.flags & TypeFlags.Union) ? type.types?.map(t => t.id) : undefined,
                intersectionTypes: (type.flags & TypeFlags.Intersection) ? type.types.map(t => t.id) : undefined,
                aliasTypeArguments: type.aliasTypeArguments?.map(t => t.id),
                keyofType: (type.flags & TypeFlags.Index) ? type.type?.id : undefined,
                ...indexedAccessProperties,
                ...referenceProperties,
                ...conditionalProperties,
                ...substitutionProperties,
                ...reverseMappedProperties,
                ...evolvingArrayProperties,
                destructuringPattern: getLocation(type.pattern),
                firstDeclaration: getLocation(symbol?.declarations?.[0]),
                flags: Debug.formatTypeFlags(type.flags).split("|"),
                display,
            };
            fs.writeSync(typesFd, JSON.stringify(descriptor));
            if (i < numTypes - 1) {
                fs.writeSync(typesFd, ",\n");
            }
        }
        fs.writeSync(typesFd, "]\n");
        fs.closeSync(typesFd);
        performance.mark("endDumpTypes");
        performance.measure("Dump types", "beginDumpTypes", "endDumpTypes");
    }
    function dumpLegend() {
        if (!legendPath) {
            return;
        }
        fs.writeFileSync(legendPath, JSON.stringify(legend));
    }
    tracingEnabled.dumpLegend = dumpLegend;
})(tracingEnabled || (tracingEnabled = {}));
// define after tracingEnabled is initialized
/** @internal */
export const startTracing = tracingEnabled.startTracing;
/** @internal */
export const dumpTracingLegend = tracingEnabled.dumpLegend;
