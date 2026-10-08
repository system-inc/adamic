import hashlib, json, os, pathlib, statistics, subprocess, sys, time
artifacts = pathlib.Path(sys.argv[1]).resolve()
repository = pathlib.Path(sys.argv[2]).resolve()
compiler = pathlib.Path(sys.argv[3]).resolve()
output = pathlib.Path(sys.argv[4]).resolve()
output.mkdir(exist_ok=True, parents=True)
results = {}
for corpus, config in [("repository", str(repository / "tsconfig.json")), ("compiler", str(compiler / "src/compiler/tsconfig.json"))]:
    runs = {"native": [], "go": []}
    expected = None
    for iteration in range(3):
        for kind in (["go", "native"] if iteration % 2 == 0 else ["native", "go"]):
            binary = artifacts / ("wave-22-next" if kind == "native" else "wave-22-next-oracle")
            prefix = output / f"{corpus}-{iteration}-{kind}"
            env = dict(os.environ, ADAMIC_TSGO_TIMING="1")
            with prefix.with_suffix(".stdout").open("wb") as stdout, prefix.with_suffix(".stderr").open("wb") as stderr:
                start = time.perf_counter()
                result = subprocess.run([str(binary), config, str(artifacts / (corpus + ".manifest"))], cwd=repository, env=env, stdout=stdout, stderr=stderr, check=True)
                elapsed = time.perf_counter() - start
            stream = prefix.with_suffix(".stdout").read_bytes()
            if expected is None: expected = stream
            assert expected == stream, "timed diagnostic streams differ"
            counters = dict(item.split("=") for item in prefix.with_suffix(".stderr").read_text().split() if "=" in item)
            runs[kind].append({"elapsed_seconds": elapsed, "counters": {k:int(v) for k,v in counters.items()}})
    medians = {k: statistics.median(r["elapsed_seconds"] for r in v) for k,v in runs.items()}
    results[corpus] = {"runs": runs, "median_seconds": medians, "native_over_go": medians["native"] / medians["go"], "diagnostic_sha256": hashlib.sha256(expected).hexdigest()}
output.joinpath("timings.json").write_text(json.dumps(results, indent=2) + "\n")
for corpus, result in results.items(): print(corpus, result["median_seconds"], "native/Go", result["native_over_go"])
