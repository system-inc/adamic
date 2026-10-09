Delivered only 146: fieldTypesNeeded records CheckedCast.CheckedFields, so p18 prints 8 in both backends as Node does. The p18 lowering fixture is the only added fixture; no oracle registration or counts row changed.

Before the fix p18 compiled natively with empty stdout instead of Node's 8. The baseline log records that disagreement. Full internal/lower passed on the broader development snapshot (197.358s); TestCallTargetReaders passed (86.437s). The minimal delivery has its own targeted agreement run. Callable-contract retention and tuple recognition are saved in the named stash for separate units. They passed their behavioral probes, but are excluded from this commit. The broader census caught and then verified correction of an overly broad spread refusal.

Tool setup: Go 0.186s, Node 0.205s, submodules 0.842s, markdown dependencies 0.847s, clang 1.371s, build 79.433s, cache warm 79.629s, total 79.680s; nproc 5, CPU quota 4.
