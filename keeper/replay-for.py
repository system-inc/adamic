"""replay-for.py: the replay brief for one mutant diff, its packages from narrow.py (covered first, then unmapped).

	replay-for.py <brief> <base sha> <name> <mutant.diff> <package> ...
"""
import hashlib, sys

brief, base, name, diffPath = sys.argv[1:5]
packages = sys.argv[5:]
diff = open(diffPath).read().rstrip("\n")
digest = hashlib.sha256(open(diffPath, "rb").read()).hexdigest()[:12]
module = "github.com/system-inc/adamic/"
listed = "\n".join("  - `./%s`" % package.replace(module, "") for package in packages)
print(open(brief).read().replace("__BASE__", base).replace("__NAME__", name).replace("__HASH__", digest).replace("__DIFF__", diff).replace("__PACKAGES__", listed))
