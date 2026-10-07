"""Run compiling inventory mutants in disposable overlays; never edit the baseline."""
import json
from pathlib import Path
import subprocess
import sys
import tempfile

root = Path(__file__).resolve().parents[5]
source = Path(__file__).with_name('engine.go').read_text()
virtual = root / 'cohere/adamic_inventory.go'
virtual_test = root / 'cohere/adamic_inventory_test.go'
logdir = Path(sys.argv[1]).resolve()
logdir.mkdir(parents=True, exist_ok=True)
cases = [
 ('failed-family-as-pass', r'strings.HasPrefix(line, "ok  \t"+rulesPrefix+family+"\t")', 'strings.Contains(line,rulesPrefix+family)', 'TestFailedFamilyCannotBeMarkedPassed'),
 ('adamic-bridge-boundary', 'if err := check(".a"); err == nil', 'if err := check(".ts"); err == nil', 'TestAdamicCorpusRequiresExplicitBridge'),
 ('message-only-port', 'return regexp.MustCompile(`(?:enabled\\(\\s*|selected\\s*===\\s*)[\'"]` + regexp.QuoteMeta(name) + `[\'"]`).MatchString(text)', '_ = regexp.QuoteMeta(name); return strings.Contains(text,name)', 'TestBranchEvidenceRequiresExecutableSelector'),
 ('transitive-edge', 'walk(next, depth+1)', '_ = next', 'TestTransitiveSiblingAndMethodDependencies'),
 ('checker-field', 'questions["checker supplied through Context.TypeChecker"] = true', '_ = questions', 'TestTransitiveSiblingAndMethodDependencies'),
 ('ranking-denominator', 'Count: len(names),', 'Count: len(names)+1,', 'TestHelperRankingCountsRulesOnce'),
 ('unknown-as-zero', 'return "unknown"', 'return "0"', 'TestUnknownFrequencyIsNotZero'),
 ('finding-count', '*report.Count++', '*report.Count += 2', 'TestRegisteredCorpusControl'),
]
with tempfile.TemporaryDirectory(prefix='lint-inventory-mutants-') as scratch:
    scratch = Path(scratch)
    for name, before, after, test in cases:
        change_source = Path(__file__).with_name("engine_test.go").read_text() if name=="adamic-bridge-boundary" else source
        assert change_source.count(before) == 1, (name, change_source.count(before))
        changed = scratch / (name+'.go')
        changed.write_text(change_source.replace(before,after,1))
        overlay = scratch / (name+'.json')
        mutant_map={str(virtual):str(changed),str(virtual_test):str(Path(__file__).with_name('engine_test.go'))}
        if name=='adamic-bridge-boundary': mutant_map={str(virtual):str(Path(__file__).with_name('engine.go')),str(virtual_test):str(changed)}
        overlay.write_text(json.dumps({'Replace':mutant_map}))
        baseline_overlay=scratch/(name+'-baseline.json')
        baseline_overlay.write_text(json.dumps({'Replace': {str(virtual):str(Path(__file__).with_name('engine.go')),str(virtual_test):str(Path(__file__).with_name('engine_test.go'))}}))
        baseline_command=['go','test','-overlay='+str(baseline_overlay),'-count=1','-run','^'+test+'$','-v',str(virtual),str(virtual_test)]
        with (logdir/(name+'-baseline.log')).open('wb') as log:
            baseline=subprocess.run(baseline_command,cwd=root/'cohere',stdout=log,stderr=log)
        baseline_evidence=(logdir/(name+'-baseline.log')).read_text()
        assert baseline.returncode==0 and '--- PASS: '+test in baseline_evidence,(name,baseline_evidence)
        command = ['go','test' ,'-overlay='+str(overlay),'-count=1','-run','^'+test+'$','-v',str(virtual),str(virtual_test)]
        with (logdir/(name+'.log')).open('wb') as log:
            result = subprocess.run(command,cwd=root/'cohere',stdout=log,stderr=log)
        evidence=(logdir/(name+'.log')).read_text()
        assert result.returncode != 0 and '--- FAIL: '+test in evidence and '[build failed]' not in evidence, (name,evidence)
        print(name+': caught by '+test)
