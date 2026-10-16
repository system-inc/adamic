#!/usr/bin/env python3
"""Remove one array-union check per private overlay; the named pin must fail."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import sys

root = Path(__file__).resolve().parents[3]
scratch = Path(tempfile.mkdtemp(prefix="views-v3-completeness-mutants-"))
os.sched_setaffinity(0, sorted(os.sched_getaffinity(0))[:4])
os.environ["GOMAXPROCS"] = "4"
c = "internal/native/runtime/view_unions_untagged.c"
j = "internal/javascript/view_unions_untagged.go"
member = "^TestCheckedViewArrayUnionMembership$/^"
mutants = [
    ("native-array-kind", c, "value->kind != adamic_view_union_array || ", "", "^TestCheckedViewUntaggedArraySource$/^wrong$"),
    ("native-element", c, "else if (!plain_matches_depth(contracts, count, contract->element, &actual, depth + 1, path))", "else if (false && !plain_matches_depth(contracts, count, contract->element, &actual, depth + 1, path))", member + "boolean$"),
    ("native-selector", c, "if (!adamic_view_untagged_plain_slot(NULL, &actual, field->name, &selected) || !plain_matches_depth(contracts, count, field->contract, &selected, depth + 1, path))", "if (false && (!adamic_view_untagged_plain_slot(NULL, &actual, field->name, &selected) || !plain_matches_depth(contracts, count, field->contract, &selected, depth + 1, path)))", "^TestCheckedViewUntaggedArraySource$/^mixed$"),
    ("native-sparse-slots", c, "array->sparse == NULL ? array->length : array->sparse->used", "array->sparse == NULL ? array->length : 0", member + "sparse-wrong$"),
    ("native-present-undefined", c, "actual.kind = number.present ? adamic_view_union_number : adamic_view_union_undefined;", "actual.kind = adamic_view_union_number;", member + "maybe-number$"),
    ("native-tuple-kind", c, "object->tuple || (object->class != NULL && object->class->is_static)", "object->class != NULL && object->class->is_static", member + "tuple-object$"),
    ("javascript-element", j, "return matches(actual.value,contract.Element,depth+1);", "return true;", member + "boolean$"),
    ("javascript-selector", j, "return selectors.every(field=>{const selected=slot(actual.value,field.Name);return selected!==undefined && matches(selected.value,field.Contract,depth+1);});", "return true;", "^TestCheckedViewUntaggedArraySource$/^mixed$"),
    ("javascript-sparse-slots", j, "Number(key)<value.length", "Number(key)<2147483648", member + "sparse-wrong$"),
    ("javascript-tuple-kind", j, "Array.isArray(actual.value) || ", "", member + "tuple-object$"),
    ("array-read-activation", "internal/lower/view_array_demand.go", "p.ArrayViewEnabled = p.ArrayViewEnabled || demand", "p.ArrayViewEnabled = false", "^TestCheckedViewObjectPrimitiveSource$/^comment-flags-wrong$"),
    ("unavailable-element", "internal/lower/view_unions_untagged.go", "return contract.Element != 0 && supported(contract.Element)", "return true", "^TestArrayMembershipUsesUnionMatcher$"),
    ("narrowed-array-kind", "internal/lower/locals.go", "matches = ir.ArrayIsArray{Value: held}", "matches = ir.BooleanConstant{Value: true}", "^TestCheckedViewArrayUnionNarrowing$"),
    ("joined-storage", "internal/lower/object.go", "element == nil || !l.sameArrayMemberStorage(arrayType)", "element == nil", "^TestMixedArrayElementStorageRemainsRefused$"),
]
if len(sys.argv)>1:
    mutants=[row for row in mutants if row[0] in sys.argv[1:]]
    if not mutants:
        raise AssertionError("no named mutants selected")
results = []
for name, file, old, new, test in mutants:
    original = (root / file).read_text()
    if original.count(old) != 1:
        raise AssertionError((name, "mutation is not unique", original.count(old)))
    replacement = scratch / (name + Path(file).suffix)
    replacement.write_text(original.replace(old, new, 1))
    overlay = scratch / (name + ".json")
    overlay.write_text(json.dumps({"Replace": {str(root / file): str(replacement)}}))
    package = "lower" if name in {"unavailable-element", "joined-storage"} else "oracle"
    log = scratch / (name + ".jsonl")
    with log.open("w") as output:
        ran = subprocess.run(["go", "test", "-json", "-count=1", "-timeout", "5m", "-overlay", str(overlay), "./internal/" + package, "-run", test], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    events = []
    for line in log.read_text().splitlines():
        try:
            events.append(json.loads(line))
        except json.JSONDecodeError:
            pass
    failed = [event for event in events if event.get("Action") == "fail" and event.get("Test")]
    builds = [event for event in events if event.get("Action") == "build-fail" or event.get("FailedBuild")]
    caught = ran.returncode != 0 and bool(failed) and not builds
    row = {"name": name, "test": test, "caught": caught, "exit": ran.returncode, "failures": [{"test": event["Test"], "seconds": event.get("Elapsed")} for event in failed]}
    results.append(row)
    print(json.dumps(row), flush=True)
    (scratch / "results.json").write_text(json.dumps(results, indent=2) + "\n")
    if not caught:
        raise AssertionError((name, "pin did not catch omission", str(log)))
print("All", len(results), "omission mutants caught; logs:", scratch)
