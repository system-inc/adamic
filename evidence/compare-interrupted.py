"""Build the report from both raw census streams."""
import argparse
import gzip
from html import escape
import json
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("report", type=Path)
args = parser.parse_args()
evidence = Path(__file__).resolve().parent
phrases = ("polymorphic recursion", "instantiates itself deeper")


def rows(label, name):
    path = evidence / label / (name + ".jsonl")
    if path.is_file():
        with path.open() as stream:
            return [json.loads(line) for line in stream]
    with gzip.open(path.with_suffix(".jsonl.gz"), "rt") as stream:
        return [json.loads(line) for line in stream]


expected_files = set()
for line in (evidence / "adapted-src.sha256").read_text().splitlines():
    _, path = line.split("  ", 1)
    if path.startswith("src/compiler/"):
        expected_files.add(path)
roots = sorted(path for path in expected_files if Path(path).suffix in {".ts", ".a"})
assert len(roots) == 79
tree, marker, _ = rows("base", "census")[0]["roots"][0].rpartition("/src/compiler/")
assert tree and marker


def relative(value):
    return value.replace(tree + "/", "")


def stop_message(stop):
    where = stop.get("where", "")
    reason = stop.get("reason", "")
    if stop["kind"] == "accepted":
        return "Accepted"
    if stop["kind"] == "checker":
        return "Checker: " + "\n".join(stop["diagnostics"])
    if stop["kind"] == "NotYet":
        return f"NotYet — {where}: stage 0 can't lower {reason} yet"
    if stop["kind"] == "Refused":
        return f"Refused — {where}: Adamic 0.1 refuses {reason}; {stop.get('fix', '')}"
    if stop["kind"] == "error":
        return f"Error — {reason}"
    raise ValueError(stop)


def cell(value):
    return escape(value, quote=False).replace("|", "&#124;").replace("\n", "<br>")


def census(label):
    observed = {}
    recovered = set()
    directory_rows = rows(label, "census")
    for row in directory_rows:
        assert row["module_format"] == "CommonJSBundle"
        (name,) = row["roots"]
        name = relative(name)
        if name in roots:
            assert name not in observed
            observed[name] = row
    directory = evidence / label / "recovered"
    paths = sorted(directory.glob("*.jsonl")) or sorted(directory.glob("*.jsonl.gz"))
    for path in paths:
        with (gzip.open(path, "rt") if path.suffix == ".gz" else path.open()) as stream:
            records = [json.loads(line) for line in stream]
        if not records:
            continue
        row = records[0]
        assert row["module_format"] == "CommonJSBundle"
        (name,) = row["roots"]
        name = relative(name)
        assert name in roots and name not in observed
        observed[name] = row
        recovered.add(name)
    return {
        name: {k: list(map(relative, v)) if k == "diagnostics" else relative(v)
               for k, v in row.items() if k in {"kind", "where", "reason", "fix", "diagnostics"}}
        for name, row in observed.items()
    }, recovered


(base, base_recovered), (head, head_recovered) = map(census, ("base", "revised-head"))
paired = base.keys() & head.keys()
moves = [name for name in sorted(paired) if base[name] != head[name]]
hits = [(name, stop_message(head[name])) for name in sorted(head)
        if any(phrase in stop_message(head[name]) for phrase in phrases)]
lines = [
    "# TypeScript expansive generics census",
    "",
    f"- **{len(moves)}** first stops moved across **{len(paired)} of 79** paired source roots.",
    f"- **{len(hits)}** of the **{len(head)}** recorded head root stops contained either search phrase.",
    "- **Entry latent comparison unavailable.** Both entry processes were interrupted before either emitted findings.",
    "",
    "## Method and interruption",
    "",
    "Base was `44361d277f655f96613a60a7d3c2e8516ce61d2c`. Head was its direct child `bfe663fba078d96362cabd80801f6a8ec225a116` (Lower #ykd89v9). Both used tools built from their own revision, including the latent and entry overlays, with recursively initialized submodules. The superseded head `169bcb04` produced no measurements.",
    "",
    "Both revisions used one tree adapted with the base `stage3/apply.sh`. The upstream input was the meter's shallow TypeScript `v6.0.3` tag clone, verified at `050880ce59e30b356b686bd3144efe24f875ebc8`.",
    "",
    "The worker interrupted both directory censuses and both entry passes. Each directory census had emitted 71 per-file records, including 69 of the 79 TypeScript roots. The entry tools wrote their checker-clean headers and unit timings. Their child processes were defunct after the environment interruption, with no recorded exit codes or relation reports.",
    "",
    f"The root table uses the completed directory records and **{len(base_recovered)} base** and **{len(head_recovered)} head** additional first records from invocations of the same pinned census binaries on the missing files. These recovery invocations use the same singleton-root inspection as the meter's directory loop. Both variants used `CENSUS_MODULE_FORMAT=CommonJSBundle`.",
    "",
    "## First stops for all 79 roots",
    "",
    "| Root | First stop |",
    "| --- | --- |",
]
for root in roots:
    left = stop_message(base[root]) if root in base else "Unobserved"
    right = stop_message(head[root]) if root in head else "Unobserved"
    value = left if left == right else f"**Base** {left}\n**Head** {right}"
    lines.append(f"| `{root}` | {cell(value)} |")
lines += ["", "## Head diagnostic phrase search", ""]
if hits:
    lines.extend(f"- `{name}` — {escape(message, quote=False)}" for name, message in hits)
else:
    lines.append("The recorded head first stops contained neither `polymorphic recursion` nor `instantiates itself deeper`.")
lines += ["", "The entry latent phrase check is unresolved because neither entry run emitted its finding records.",
          "", "## Entry latent comparison", ""]
timed = {}
for label in ("base", "revised-head"):
    (header,) = rows(label, "latent")
    assert header["status"] == "measurement" and not header["checker_rejected"]
    assert header["module_format"] == "CommonJSBundle" and header["latent_mode"] == "first-error"
    assert [relative(path) for path in header["roots"]] == ["src/tsc/tsc.ts"]
    timed[label] = len(rows(label, "latent-times"))
lines += [
    "Both entry passes ran at `src/tsc/tsc.ts` with `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`, and `LATENT_TAIL=1` in CommonJSBundle mode.",
    "",
    f"Base recorded {timed['base']:,} timed units and head recorded {timed['revised-head']:,}. The entry tool holds its findings until after program-wide processing and relation checks. The timing records cannot establish which findings were new, gone, changed, or newly reached.",
    "",
    "See [evidence](evidence/) for the surviving raw streams, input pins, tool hashes, and reproduction commands.",
    "",
]
(evidence / "comparison.json").write_text(json.dumps({
    "paired_roots": len(paired),
    "unobserved_base_roots": sorted(set(roots) - base.keys()),
    "unobserved_head_roots": sorted(set(roots) - head.keys()),
    "first_stop_moves": {name: {"base": base[name], "head": head[name]} for name in moves},
    "head_first_stop_phrase_hits": hits,
    "entry_status": "interrupted before findings were emitted",
    "entry_timed_units": timed,
}, indent=2) + "\n")
args.report.write_text("\n".join(lines))
