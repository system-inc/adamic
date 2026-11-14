"""Add the saved scan-only wrapper/driver to an existing scratch latent overlay.
Usage: python build_scan.py REPOSITORY FULL_OVERLAY_JSON NEW_SCRATCH_DIRECTORY
Then build the tool with the printed overlay path. No compiler source is edited.
"""
import pathlib, json, sys
repo=pathlib.Path(sys.argv[1]).resolve();original=pathlib.Path(sys.argv[2]).resolve();out=pathlib.Path(sys.argv[3]).resolve()
assert not out.is_relative_to(repo), 'generated Go must stay outside repository'
out.mkdir(parents=True,exist_ok=False);sources=pathlib.Path(__file__).resolve().parent
wrapper=(sources/'scan-only.go.txt').read_text();driver=(sources/'scan-driver.go.txt').read_text()
assert 'lower.LatentScanOnly(context.Background()' in driver
assert 'lower.LowerWithOptions' in driver and 'load.Load(paths)' in driver
(out/'scan-only.go').write_text(wrapper);(out/'main.go').write_text(driver)
d=json.loads(original.read_text());d['Replace'][str(repo/'internal/lower/latent_scan_only.go')]=str(out/'scan-only.go');d['Replace'][str(repo/'stage3/census/latent/tool/main.go')]=str(out/'main.go')
(out/'overlay.json').write_text(json.dumps(d,indent=2)+'\n');print(out/'overlay.json')
