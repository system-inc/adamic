#!/usr/bin/env python3
"""Fit and project the measurements in rows.json; writes summary.json and prints markdown tables.

    python3 -I docs/tsc-generated-c/report.py OUT/rows.json

Fits are least squares in log-log space, y = a * x^b, over every row that compiled to C (the tsc
files that lower plus the stage1 proxies). The error is the residual standard deviation in natural
log, so exp(s) is the typical multiplicative miss; ranges are exp(+-2s) around the fit.
"""
import json
import math
import os
import re
import sys

LEVELS = ("-O0", "-O1", "-O2")


def fit(points):
    xs = [math.log(x) for x, _ in points]
    ys = [math.log(y) for _, y in points]
    n = len(points)
    mean_x, mean_y = sum(xs) / n, sum(ys) / n
    sxx = sum((x - mean_x) ** 2 for x in xs)
    b = sum((x - mean_x) * (y - mean_y) for x, y in zip(xs, ys)) / sxx
    a = mean_y - b * mean_x
    residuals = [y - (a + b * x) for x, y in zip(xs, ys)]
    s = math.sqrt(sum(r * r for r in residuals) / (n - 2)) if n > 2 else float("nan")
    # Standard error of the exponent, for saying whether superlinear is distinguishable from linear.
    b_error = s / math.sqrt(sxx) if n > 2 else float("nan")
    return {"a": math.exp(a), "b": b, "b_standard_error": b_error, "log_residual_sd": s, "n": n,
            "x_min": min(x for x, _ in points), "x_max": max(x for x, _ in points)}


def predict(model, x, sigmas=0.0):
    return model["a"] * x ** model["b"] * math.exp(sigmas * model["log_residual_sd"])


def clang(row, level):
    for sample in row.get("clang", []):
        if sample["level"] == level and sample["exit"] == 0:
            return sample
    return None


def main():
    path = sys.argv[1]
    with open(path) as handle:
        data = json.load(handle)
    rows = data["rows"]
    # The 77 files of tsc's own src/compiler at the pin, one per line (pristine.txt beside rows.json).
    # The adapted tree adds two: hostErrors.ts (adaptation 47) and diagnosticInformationMap.generated.ts.
    with open(os.path.join(os.path.dirname(path), "pristine.txt")) as handle:
        pristine = set(handle.read().split())
    for row in rows:
        row["in_tsc_77"] = row["kind"] == "tsc" and row["file"] in pristine
        if "first_reason" in row:
            row["first_reason"] = re.sub(r"\S*/src/compiler/", "", row["first_reason"])
    compiled = [row for row in rows if row.get("c_bytes")]
    tsc = [row for row in rows if row["in_tsc_77"]]
    summary = {"rows_compiled": len(compiled), "tsc_files": len(tsc),
               "tsc_files_lowering": [row["file"] for row in tsc if row["adamic_exit"] == 0],
               "adapted_files_lowering": [row["file"] for row in rows if row["kind"] == "tsc" and row["adamic_exit"] == 0],
               "fits": {}, "projection": {}}

    # C bytes against the TypeScript the compile read (the entry's relative-import closure).
    bytes_model = fit([(row["closure_ts_bytes"], row["c_bytes"]) for row in compiled])
    lines_model = fit([(row["closure_ts_lines"], row["c_lines"]) for row in compiled])
    summary["fits"]["c_bytes_by_ts_bytes"] = bytes_model
    summary["fits"]["c_lines_by_ts_lines"] = lines_model
    for level in LEVELS:
        points = [(row["c_bytes"], clang(row, level)["wall_seconds"]) for row in compiled
                  if clang(row, level) and clang(row, level)["wall_seconds"] > 0]
        summary["fits"][f"clang{level}_seconds_by_c_bytes"] = fit(points)
        points = [(row["closure_ts_bytes"], clang(row, level)["wall_seconds"]) for row in compiled
                  if clang(row, level) and clang(row, level)["wall_seconds"] > 0]
        summary["fits"][f"clang{level}_seconds_by_ts_bytes"] = fit(points)
        points = [(row["c_bytes"], clang(row, level)["max_rss_kb"]) for row in compiled if clang(row, level)]
        summary["fits"][f"clang{level}_rss_kb_by_c_bytes"] = fit(points)

    # Projection one: each of the 77 files as its own unit (its own TypeScript, not its closure).
    # Projection two: all 77 as one program, one unit.
    own = [row["own_ts_bytes"] for row in tsc]
    whole = sum(own)
    projection = {"tsc_ts_bytes": whole, "tsc_ts_lines": sum(row["own_ts_lines"] for row in tsc),
                  "largest_file": max(tsc, key=lambda row: row["own_ts_bytes"])["file"],
                  "largest_file_ts_bytes": max(own)}
    for name, sizes in (("per_file_units", own), ("one_unit", [whole])):
        entry = {}
        for sigmas, label in ((0, "fit"), (-2, "low"), (2, "high")):
            c_bytes = [predict(bytes_model, size, sigmas) for size in sizes]
            entry[f"c_bytes_{label}"] = round(sum(c_bytes))
            for level in LEVELS:
                model = summary["fits"][f"clang{level}_seconds_by_ts_bytes"]
                entry[f"clang{level}_seconds_{label}"] = round(sum(predict(model, size, sigmas) for size in sizes), 1)
                model = summary["fits"][f"clang{level}_rss_kb_by_c_bytes"]
                entry[f"clang{level}_peak_rss_kb_{label}"] = round(max(predict(model, predict(bytes_model, size), sigmas) for size in sizes))
        projection[name] = entry
    projection["extrapolation_factor_one_unit"] = whole / bytes_model["x_max"]
    projection["extrapolation_factor_largest_file"] = max(own) / bytes_model["x_max"]
    summary["projection"] = projection

    # Seconds per megabyte of C, by unit size: where it rises, clang has turned superlinear.
    per_megabyte = []
    for row in sorted(compiled, key=lambda row: row["c_bytes"]):
        entry = {"file": row["file"], "c_bytes": row["c_bytes"], "largest_function_lines": row["largest_function_lines"]}
        for level in LEVELS:
            sample = clang(row, level)
            if sample:
                entry[f"{level}_seconds_per_mb"] = round(sample["wall_seconds"] / (row["c_bytes"] / 1e6), 3)
                entry[f"{level}_rss_mb"] = round(sample["max_rss_kb"] / 1024, 1)
        per_megabyte.append(entry)
    summary["seconds_per_mb"] = per_megabyte
    output = os.path.join(os.path.dirname(path), "summary.json")
    with open(output, "w") as handle:
        json.dump(summary, handle, indent=1)
        handle.write("\n")

    print("| File | Kind | TS bytes (closure) | C bytes | C lines | Units | Split units | Largest function (lines) | -O0 s | -O1 s | -O2 s | -O0 MB | -O1 MB | -O2 MB |")
    print("|---|---|---:|---:|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|")
    for row in sorted(compiled, key=lambda row: row["c_bytes"]):
        cells = [row["file"], row["kind"], f"{row['closure_ts_bytes']:,}", f"{row['c_bytes']:,}", f"{row['c_lines']:,}",
                 str(row["units"]), str(row["split_units_if_opted_in"]),
                 f"`{row['largest_function']}` ({row['largest_function_lines']:,})"]
        for level in LEVELS:
            sample = clang(row, level)
            cells.append(f"{sample['wall_seconds']:.2f}" if sample else "failed")
        for level in LEVELS:
            sample = clang(row, level)
            cells.append(f"{sample['max_rss_kb'] / 1024:.0f}" if sample else "")
        print("| " + " | ".join(cells) + " |")
    print()
    for name, model in summary["fits"].items():
        print(f"{name}: y = {model['a']:.4g} * x^{model['b']:.3f} (+-{model['b_standard_error']:.3f}), "
              f"log residual sd {model['log_residual_sd']:.3f} (x{math.exp(model['log_residual_sd']):.2f}), n={model['n']}, "
              f"x in [{model['x_min']:,}, {model['x_max']:,}]")
    print()
    print(json.dumps(projection, indent=1))
    print()
    for entry in per_megabyte:
        print(entry)
    if data.get("time_trace"):
        print(json.dumps(data["time_trace"], indent=1))
    print()
    print("| File | In tsc's 77 | TS bytes | Lowers | First reason (adamic c) | Meter's first reason |")
    print("|---|---|---:|---|---|---|")
    for row in rows:
        if row["kind"] != "tsc":
            continue
        reason = row.get("first_reason", "").replace("|", "\\|")
        meter = row.get("meter_reason", "").replace("|", "\\|")
        print(f"| {row['file']} | {'yes' if row['in_tsc_77'] else 'no'} | {row['own_ts_bytes']:,} | {'yes' if row['adamic_exit'] == 0 else 'no'} | {reason} | {meter} |")
    return 0


if __name__ == "__main__":
    sys.exit(main())
