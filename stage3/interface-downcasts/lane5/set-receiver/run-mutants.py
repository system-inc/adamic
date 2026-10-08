#!/usr/bin/env python3
"""Run each Set receiver regression mutant independently and restore its source."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[4]
logs = root / "stage3/interface-downcasts/lane5/set-receiver/logs"
mutants = [
    ("uncast-map", "internal/native/emit_functions.go", 'value = "(adamic_object *)" + value', 'value = value', '^TestCheckedViewSetIntrinsicsC$', 'incompatible pointer types'),
    ("native-drop-self", "internal/native/runtime/view_set_intrinsics.c", 'return (adamic_value){.reference = adamic_retain(set)};', 'return (adamic_value){.reference = NULL};', '^TestCheckedViewSetIntrinsics$/^(add|identity)$', 'former union frontier'),
    ("javascript-drop-self", "internal/javascript/view_callables.go", 'Reflect.apply(value, object, values)', '(Reflect.apply(value, object, values), undefined)', '^TestCheckedViewSetIntrinsics$/^(add|identity)$', 'former union frontier'),
    ("drop-self-signature", "internal/native/runtime/view_set_intrinsics.c", '{1, parameters + 2, 4, "Set<string>.add", NULL, 0}', '{1, parameters + 2, 254, "Set<string>.add", NULL, 0}', '^TestCheckedViewSetIntrinsics$/^add$', 'former union frontier'),
    ("native-skip-signature", "internal/native/view_callables_methods.go", 'e.line("(void)adamic_view_callable_shape(&%s.heap, %s, %s, %s, false);", witness, recorded, expected, cString(property.View))', 'e.line("(void)%s; (void)%s; (void)%s;", witness, recorded, expected)', '^TestCheckedViewSetIntrinsicSignatures$', 'signature refusal'),
    ("javascript-skip-signature", "internal/javascript/view_callables.go", 'value := e.emitViewCallableCertificate(property, raw("object"), "object")', 'value := raw("object")', '^TestCheckedViewSetIntrinsicSignatures$', 'signature refusal'),
]
for name, file, old, new, test, witness in mutants:
    path = root / file
    original = path.read_text()
    expected_anchors = 2 if name == "javascript-drop-self" else 1
    if original.count(old) != expected_anchors:
        raise RuntimeError(f"{name}: mutation anchor count {original.count(old)}")
    try:
        path.write_text(original.replace(old, new, 1))
        with (logs / f"mutant-{name}.log").open("w") as log:
            result = subprocess.run(["go", "test", "./internal/oracle", "-run", test, "-count=1", "-v", "-timeout", "5m"], cwd=root, stdout=log, stderr=subprocess.STDOUT)
        output = (logs / f"mutant-{name}.log").read_text()
        if result.returncode == 0 or witness not in output:
            raise RuntimeError(f"{name}: missing intended failure, exit {result.returncode}")
        if name != "uncast-map" and ("Sanitizer" in output or "error: " in output):
            raise RuntimeError(f"{name}: failure was not semantic")
        print(f"{name}: caught, exit {result.returncode}, witness {witness}", flush=True)
    finally:
        path.write_text(original)
