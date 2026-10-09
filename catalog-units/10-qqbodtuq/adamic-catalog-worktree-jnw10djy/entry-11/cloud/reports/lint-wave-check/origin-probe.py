"""Point-in-time live-origin controls recorded by the lint-wave-check unit."""
import runpy

module = runpy.run_path("cloud/lint-wave-check.py")
worker = module["Worker"](".")
refs = worker.origin()
print("fetched origin heads:", len(refs))
for name, expected in [("no-continue", "referenced by legacy port"), ("consistent-this", "claimed in origin/")]:
    claim = {"version": 1, "branch": "codex/lint-wave-trial", "rules": [name]}
    try:
        worker.overlaps(refs, claim["branch"], module["CLAIMS"] + claim["branch"] + ".json", claim)
    except module["Rejected"] as error:
        assert name in str(error) and expected in str(error), str(error)
        print("expected rejection:", error)
    else:
        raise SystemExit("competing rule was accepted: " + name)
