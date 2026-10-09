import importlib.util,io,json,types,unittest,sys
from pathlib import Path
root=str(Path(__file__).resolve().parent.parent) + '/'
sys.path.insert(0,root)
spec=importlib.util.spec_from_file_location('suite',root+'run_test.py');suite=importlib.util.module_from_spec(spec);spec.loader.exec_module(suite)
source=open(root+'run.py').read()
mutants=[
 ('missing-full-fast-json','resultFiles = [self.kind + ".json"] + (["fast.json"] if self.kind == "full" else [])','resultFiles = [self.kind + ".json"]','FailClosed.test_all_pass_is_green'),
 ('wrong-gate-kind','self.result["gate_kind"] = self.kind','self.result["gate_kind"] = "fast"','FailClosed.test_all_pass_is_green'),
 ('round-away-overrun','elapsed > 30','round(elapsed, 3) > 30','PhaseUnitTests.test_units_coverage_failure_scratch_and_budget'),
 ('catalog-all-skipped-green','if not any(row.get("caught") and row.get("ok") for row in rows):','if False:','PhaseUnitTests.test_catalog_verdicts_are_per_entry'),
 ('omit-request-unit','fixtures += [json.loads(name) for name in runs if name != "fixture"]','fixtures += []','PhaseUnitTests.test_inventory_and_patterns'),
 ('unanchored-fixtures','"^(%s)$" % "|".join(re.escape(part) for part in level)','"(%s)" % "|".join(re.escape(part) for part in level)','PhaseUnitTests.test_inventory_and_patterns'),
 ('accept-dynamic-registration','raise ValueError("unrecognized TestWASI fixture registration")','return fixtures','PhaseUnitTests.test_inventory_refuses_unknown_registration'),
 ('ignore-cpu-quota','cpus = min(cpus, max(1, int(quota) // int(period)))','cpus = cpus','PhaseUnitTests.test_slot_cpu_quota'),
 ('unbounded-workers','max_workers=slots','max_workers=100','PhaseUnitTests.test_units_coverage_failure_scratch_and_budget'),
 ('shared-scratch','scratch = tempfile.mkdtemp(prefix="%02d-" % index, dir=root)','scratch = root','PhaseUnitTests.test_units_coverage_failure_scratch_and_budget'),
 ('drop-last-unit','zip(names, commands)','zip(names[:-1], commands[:-1])','PhaseUnitTests.test_units_coverage_failure_scratch_and_budget'),
 ('wrong-unit-count','len(rows)\n        self.result[phase + "_unit_budget_seconds"]','0\n        self.result[phase + "_unit_budget_seconds"]','PhaseUnitTests.test_units_coverage_failure_scratch_and_budget'),
 ('ignore-budget','elapsed > 30','False','PhaseUnitTests.test_units_coverage_failure_scratch_and_budget'),
 ('ignore-unit-failure','row["ok"] and not row["over_budget"]','True','PhaseUnitTests.test_units_coverage_failure_scratch_and_budget'),
 ('unnamed-unit-failure','self.fail(phase + "/" + row["name"], json.dumps(row))','self.fail(phase, json.dumps(row))','PhaseUnitTests.test_units_coverage_failure_scratch_and_budget'),
 ('aggregate-always-green','0 if rows and all(row["status"] == "passed" for row in rows) else 1','0','PhaseUnitTests.test_units_coverage_failure_scratch_and_budget'),
 ('catalog-wrong-entry','str(entry["number"])','"1"','PhaseUnitTests.test_catalog_verdicts_are_per_entry'),
 ('catalog-nested-parallel','"--jobs", "1", self.arguments.sha','"--jobs", "4", self.arguments.sha','PhaseUnitTests.test_catalog_verdicts_are_per_entry'),
 ('catalog-unbounded-go','"GOFLAGS": "-p=1"','"GOFLAGS": "-p=4"','PhaseUnitTests.test_catalog_verdicts_are_per_entry'),
 ('accept-catalog-error','(caught and code == 0) or (stale and code == 1) or (skipped and code == 0)','True','PhaseUnitTests.test_catalog_verdicts_are_per_entry'),
 ('accept-caught-with-error','(caught and code == 0)','caught','PhaseUnitTests.test_catalog_verdicts_are_per_entry'),
 ('reject-stale-upkeep','(stale and code == 1)','False','PhaseUnitTests.test_catalog_verdicts_are_per_entry'),
 ('reject-skipped-entry','(skipped and code == 0)','False','PhaseUnitTests.test_catalog_verdicts_are_per_entry'),
 ('duplicate-catalog-number',' or len({entry["number"] for entry in entries}) != len(entries)','', 'PhaseUnitTests.test_catalog_verdicts_are_per_entry'),
 ('missing-wasi-pass-green','code == 0 and passed and not unexpected','code == 0 and not unexpected','PhaseUnitTests.test_wasi_requires_named_pass_and_exact_selection'),
 ('broadened-wasi-green','code == 0 and passed and not unexpected','code == 0 and passed','PhaseUnitTests.test_wasi_requires_named_pass_and_exact_selection'),
 ('wasi-unbounded-build','"-p", "1", "-parallel", "1"','"-p", "4", "-parallel", "1"','PhaseUnitTests.test_wasi_requires_named_pass_and_exact_selection'),
 ('wasi-nested-unbounded-go','GOMAXPROCS="1", GOFLAGS="-p=1"','GOMAXPROCS="1", GOFLAGS="-p=4"','PhaseUnitTests.test_wasi_requires_named_pass_and_exact_selection'),
 ('wasi-native-clang','dict(environment or {}, TMPDIR=scratch, GOMAXPROCS="1", GOFLAGS="-p=1")','dict(TMPDIR=scratch, GOMAXPROCS="1", PATH="native")','PhaseUnitTests.test_wasi_requires_named_pass_and_exact_selection'),
 ('full-phase-unsplit','self.guarded("wasi", self.wasiSplit, log, wasi)','self.guarded("wasi", self.test, "wasi", ["go", "test", "-count=1", "-json", "-run", "^TestWASI$", "./internal/native"], log, wasi)','FailClosed.test_all_pass_is_green'),
]
results=[]
for name,old,new,test in mutants:
 assert old in source,name
 mod=types.ModuleType('mutant');exec(compile(source.replace(old,new,1),root+'run.py','exec'),mod.__dict__)
 suite.run=mod
 cases=unittest.defaultTestLoader.loadTestsFromName(test,suite)
 output=io.StringIO();result=unittest.TextTestRunner(stream=output).run(cases)
 row={'name':name,'test':test,'caught':not result.wasSuccessful(),'output':output.getvalue()}
 results.append(row);print(name, 'CAUGHT' if row['caught'] else 'SURVIVED',flush=True)
open(str(Path(__file__).with_name('mutants.json')),'w').write(json.dumps(results,indent=2)+'\n')
sys.exit(0 if all(r['caught'] for r in results) else 1)
