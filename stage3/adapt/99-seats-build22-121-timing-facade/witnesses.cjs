"use strict";

const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");

const relative = "src/compiler/performanceCore.ts";

function runAdapter(tree) {
    const before = fs.readFileSync(path.join(tree, relative), "utf8");
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "timing-facade-"));
    try {
        const file = path.join(scratch, relative);
        fs.mkdirSync(path.dirname(file), { recursive: true });
        fs.writeFileSync(file, before);
        const report = JSON.parse(execFileSync(process.execPath, [path.join(__dirname, "adapt.cjs"), scratch], {
            encoding: "utf8",
        }));
        assert.deepEqual(report, { file: relative, line: 74, edits: 3, javascriptChanging: true });
        const after = fs.readFileSync(file, "utf8");
        const oldLines = before.split(/(?<=\n)/);
        const newLines = after.split(/(?<=\n)/);
        assert.equal(newLines.length, oldLines.length + 1, "one inserted line");
        assert.deepEqual(newLines.slice(0, 72), oldLines.slice(0, 72), "unchanged through line 72");
        assert.deepEqual(newLines.slice(75), oldLines.slice(74), "unchanged from old line 75 onward");
        return { before, after, report };
    } finally {
        fs.rmSync(scratch, { recursive: true, force: true });
    }
}

function compile(text) {
    const output = ts.transpileModule(text, {
        fileName: relative,
        reportDiagnostics: true,
        compilerOptions: { module: ts.ModuleKind.CommonJS },
    });
    assert.deepEqual(output.diagnostics, []);
    return new vm.Script(output.outputText, { filename: relative });
}

function init(script, host, browser = false) {
    const exports = {};
    script.runInNewContext({
        exports,
        performance: browser ? host : undefined,
        Date: { now: () => -1 },
        require(id) {
            if (id === "node:perf_hooks") return { performance: host };
            if (id === "./_namespaces/ts.js") return { isNodeLikeSystem: () => !browser };
            throw new Error(`unexpected module: ${id}`);
        },
    });
    return exports;
}

function hostClock(origin) {
    let clock = 10;
    const host = {
        timeOrigin: origin,
        now() {
            assert.equal(this, host, "host now receiver");
            return clock;
        },
        mark() {},
        measure() {},
        clearMarks() {},
        clearMeasures() {},
    };
    return { host, advance() { clock = 20; } };
}

function checkTiming(timing, host, adapted, expectedOrigin) {
    if (adapted) assert.notEqual(timing, host, "timing facade identity");
    else assert.equal(timing, host, "original timing slot identity");
    assert.equal(timing.timeOrigin, expectedOrigin, "copied origin");
}

function clockRow(script, adapted, origin, browser) {
    const { host, advance } = hostClock(origin);
    const core = init(script, host, browser);
    const hooks = core.tryGetNativePerformanceHooks();
    assert.equal(hooks.performance, host, "event slot identity");
    assert.equal(hooks.shouldWriteNativeEvents, browser, "event flag");
    checkTiming(hooks.performanceTime, host, adapted, host.timeOrigin);
    assert.equal(core.timestamp(), 10, "timestamp delegates to the host clock");
    advance();
    assert.equal(core.timestamp(), 20, "timestamp tracks the host clock");
    return "receiver checked at 10 and 20; event slot is the host";
}

function timingOnly(script, adapted, origin) {
    const host = { timeOrigin: origin, now: () => 20 };
    const core = init(script, host);
    const hooks = core.tryGetNativePerformanceHooks();
    assert.equal(hooks.performance, undefined);
    checkTiming(hooks.performanceTime, host, adapted, host.timeOrigin);
    assert.equal(core.timestamp(), 20);
}

function missingNow(script, _adapted, origin) {
    const core = init(script, { timeOrigin: origin });
    const hooks = core.tryGetNativePerformanceHooks();
    assert.equal(hooks.performance, undefined);
    assert.equal(hooks.performanceTime, undefined);
    assert.equal(core.timestamp(), -1);
}

function missingHost(script) {
    const core = init(script, undefined);
    assert.equal(core.tryGetNativePerformanceHooks(), undefined);
    assert.equal(core.timestamp(), -1);
}

function polyfill(script, adapted, origin) {
    let reads = 0;
    const { host } = hostClock(origin);
    Object.defineProperty(host, "timeOrigin", {
        get() {
            if (++reads > 1) throw new Error("second timeOrigin read");
            return origin;
        },
    });
    const core = init(script, host);
    const hooks = core.tryGetNativePerformanceHooks();
    assert.equal(hooks.performance, host, "polyfill event identity");
    if (adapted) checkTiming(hooks.performanceTime, host, true, origin);
    else assert.equal(hooks.performanceTime, host, "original polyfill timing slot identity");
    assert.equal(reads, 1, "timeOrigin getter reads");
    return adapted ? `initialized; reads=1; facade origin=${origin}` : "initialized; reads=1 at guard; timing slot is the host";
}

function main(tree, output) {
    assert.equal(ts.version, "6.0.3");
    const destination = path.join(fs.realpathSync(path.dirname(path.resolve(output))), path.basename(output));
    const relativeOutput = path.relative(fs.realpathSync(__dirname), destination);
    assert(relativeOutput.startsWith(`..${path.sep}`), "output must be outside the adapter directory");
    const { before, after, report } = runAdapter(tree);
    const cases = [
        ["clock", 1000, (script, adapted, origin) => clockRow(script, adapted, origin, false)],
        ["timing-only", 1000, timingOnly],
        ["missing-now", 2718, missingNow],
        ["missing-host", null, missingHost],
        ["browser", 2718, (script, adapted, origin) => clockRow(script, adapted, origin, true)],
        ["polyfill", 2718, polyfill],
    ];
    const verdicts = [];
    for (const [side, text] of [["before", before], ["adapted", after]]) {
        const script = compile(text);
        for (const [row, origin, check] of cases) {
            let verdict;
            try {
                verdict = { side, row, origin, passed: true, detail: check(script, side === "adapted", origin) };
            } catch (error) {
                const detail = error.code === "ERR_ASSERTION" && error.actual !== undefined
                    ? `${error.message.split("\n")[0]} (actual ${error.actual}, expected ${error.expected})`
                    : error.message;
                verdict = { side, row, origin, passed: false, detail };
            }
            verdicts.push(verdict);
            console.log(`${verdict.passed ? "PASS" : "FAIL"} ${side} ${row} (origin=${origin ?? "absent"})${verdict.detail ? `: ${verdict.detail}` : ""}`);
        }
    }
    fs.writeFileSync(destination, JSON.stringify({ report, verdicts }, null, 2) + "\n", { flag: "wx" });
    if (verdicts.some(verdict => !verdict.passed)) process.exitCode = 1;
}

assert.equal(process.argv.length, 4, "usage: node witnesses.cjs BEFORE_TREE OUTPUT.json");
main(process.argv[2], process.argv[3]);
