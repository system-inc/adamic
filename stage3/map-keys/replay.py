"""Replay only the step 19 key-representation findings with the guarded census API."""
import argparse
import gzip
import json
import pathlib
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument("tree", type=pathlib.Path)
parser.add_argument("output", type=pathlib.Path)
args = parser.parse_args()
repository = pathlib.Path(__file__).resolve().parents[2]
reasons = {
    "a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions",
    "a Map of ResolvedConfigFilePath",
}
probes = {}
with gzip.open(repository / "stage3/meter/runs/20261008T035244Z.latent-full/tsc/full.jsonl.gz", "rt") as source:
    for line in source:
        for finding in json.loads(line).get("findings", []):
            reason = finding.get("reason", "")
            if finding.get("kind") != "NotYet" or not (reason in reasons or reason.startswith("a Set of __String (") or reason.startswith("a Set of Path (") or reason.startswith("a Set of ResolvedConfigFilePath (")):
                continue
            suffix = finding["where"].split("/src/", 1)[1]
            where = str(args.tree.resolve() / "src" / suffix)
            probes[(where, reason)] = {"where": where, "reason": reason, "kind": "NotYet"}
with tempfile.TemporaryDirectory(prefix="scout19-replay-") as directory:
    scratch = pathlib.Path(directory)
    subprocess.run(["python3", str(repository / "stage3/census/latent/make_overlay.py"), str(repository), str(scratch)], cwd=repository, check=True)
    overlay = json.loads((scratch / "overlay.json").read_text())
    worker = pathlib.Path(overlay["Replace"][str(repository / "stage3/census/latent/replay/worker/main.go")])
    worker.write_text(r'''package main
import (
    "context"
    "encoding/json"
    "os"
    "github.com/system-inc/adamic/internal/load"
    "github.com/system-inc/adamic/internal/lower"
)
func main() {
    if err := os.Setenv("LATENT_FULL", "1"); err != nil { panic(err) }
    program, err := load.LatentLoad([]string{os.Args[1]})
    if err != nil { panic(err) }
    program.LatentReplayEntryReach()
    output, failure := lower.Lower(context.Background(), program)
    if output != nil || failure == nil || failure.Error() != "latent census: measurement only; no IR output" { panic("measurement exposed IR") }
    exposed, failure := load.Load([]string{os.Args[1]})
    if exposed != nil || failure == nil { panic("measurement loader exposed a program") }
    var probes []struct { Where, Kind, Reason string }
    file, err := os.Open(os.Args[2]); if err != nil { panic(err) }
    if err := json.NewDecoder(file).Decode(&probes); err != nil { panic(err) }; file.Close()
    encoder := json.NewEncoder(os.Stdout)
    for _, probe := range probes {
        record, failure := lower.LatentReplay(context.Background(), program, probe.Where, probe.Kind, probe.Reason)
        result := map[string]any{"probe":probe, "reproduced":failure == nil, "units":record.Units, "findings":record.Findings, "checker_rejected":len(program.LatentDiagnostics()) > 0}
        if failure != nil { result["failure"] = failure.Error() }
        if err := encoder.Encode(result); err != nil { panic(err) }
    }
}
''')
    inputs = scratch / "probes.json"
    inputs.write_text(json.dumps(list(probes.values())))
    with args.output.open("w") as destination:
        subprocess.run(["go", "run", "-buildvcs=false", "-overlay=" + str(scratch / "overlay.json"), "./stage3/census/latent/replay/worker", str(args.tree.resolve() / "src/tsc/tsc.ts"), str(inputs)], cwd=repository, stdout=destination, check=True)
