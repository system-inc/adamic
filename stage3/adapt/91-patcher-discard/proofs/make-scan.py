#!/usr/bin/env python3
"""Generate a separate, explicitly labeled compiler-owned void-only source scan."""
import json
from pathlib import Path
import subprocess
import sys

repository, existing, destination = map(lambda p: Path(p).resolve(), sys.argv[1:])
destination.mkdir(parents=True, exist_ok=False)
overlay = json.loads(existing.read_text())
main_key = str(repository / 'stage3/census/latent/tool/main.go')
main = Path(overlay['Replace'][main_key]).read_text()
call = 'lower.Latent(context.Background(), program, '
assert main.count(call) == 1
main = main.replace(call, 'lower.VoidWideningSourceScan(context.Background(), program, ')
assert main.count('"latent_mode": censusMode()') == 1
main = main.replace('"latent_mode": censusMode()', '"latent_mode": "void-widening-only-source-scan"')
main_path = destination / 'main.go'
main_path.write_text(main)
overlay['Replace'][main_key] = str(main_path)
scan = r'''package lower
import (
 "context"
 "fmt"
 "github.com/microsoft/TypeScript/tsc/shim/ast"
 "github.com/system-inc/adamic/internal/ir"
 "github.com/system-inc/adamic/internal/load"
)
// This scan invokes the production refusal unchanged. It performs no lowering
// and makes no claim about other refusals or concrete generic instantiations.
func VoidWideningSourceScan(ctx context.Context, program *load.Program, emit func(LatentFile)) {
 for _, file := range program.Files() {
  latentFindings = nil
  checker, release := program.Checker(ctx, file)
  l := &lowering{program: program, checker: checker, result: &ir.Program{}, this: -1, functionIndex: -1}
  var visit ast.Visitor
  visit = func(node *ast.Node) bool {
   latentFindingOwner, latentFindingPhase = program.Where(node), "void_widening_only_source_scan"
   func() {
    defer func() {
     if failure := recover(); failure != nil {
      latentFindings = append(latentFindings, LatentFinding{Measurement: LatentMeasurementLabel, Unit: latentFindingOwner, Phase: latentFindingPhase, Kind: "panic", Where: program.Where(node), Reason: fmt.Sprint(failure), Text: fmt.Sprint(failure)})
     }
    }()
    latentRecord(l.refuseVoidWidening(node))
   }()
   node.ForEachChild(visit)
   return false
  }
  visit(file.AsNode())
  release()
  emit(LatentFile{Measurement: LatentMeasurementLabel, File: program.FileName(file), Findings: latentFindings})
 }
}
'''
scan_path = destination / 'scan.go'
scan_path.write_text(scan)
overlay['Replace'][str(repository / 'internal/lower/void_widening_source_scan.go')] = str(scan_path)
subprocess.run(['gofmt', '-w', str(main_path), str(scan_path)], check=True, timeout=30)
(destination / 'overlay.json').write_text(json.dumps(overlay, indent=2)+'\n')
print(destination / 'overlay.json')
