"""reached.py: the packages whose test binaries depend on any of the given package directories, from go list in a
checkout of the base: a package reaches a mutant when the mutated package is among its own dependencies or among the
dependencies of its test imports. Import paths, one per line, sorted."""
import json, subprocess, sys
checkout, changed = sys.argv[1], sys.argv[2:]
module = "github.com/system-inc/adamic/"
listing = subprocess.run(["go", "list", "-e", "-json=ImportPath,Deps,TestImports,XTestImports", "./..."], cwd=checkout, check=True, capture_output=True, text=True).stdout
packages = {}
for chunk in listing.replace("}\n{", "}\x00{").split("\x00"):
    package = json.loads(chunk)
    packages[package["ImportPath"]] = package
targets = {module + directory.strip("/") for directory in changed}
reached = []
for path, package in packages.items():
    deps = set(package.get("Deps") or []) | {path}
    for imported in (package.get("TestImports") or []) + (package.get("XTestImports") or []):
        deps.add(imported)
        deps |= set((packages.get(imported) or {}).get("Deps") or [])
    if deps & targets:
        reached.append(path)
print("\n".join(sorted(reached)))
