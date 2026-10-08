"""Undo each tsc-corpus fix in scratch and require a file-named oracle failure."""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile


def main():
    arguments = argparse.ArgumentParser()
    arguments.add_argument("--logs", type=Path, default=Path("/tmp/tsprinter-tsc-mutants"))
    arguments.add_argument("--roots-only", action="store_true")
    arguments.add_argument("--upstream-only", action="store_true")
    options = arguments.parse_args()
    options.logs.mkdir(parents=True, exist_ok=True)
    port = Path(__file__).resolve().parents[1]
    repository = port.parents[2]
    parser = repository / "stage1/typescript/parser"
    scanner = repository / "stage1/typescript/scanner"
    if options.upstream_only:
        pins = json.loads((port / "testdata/tsc-upstream-differences.json").read_text())
        helper = (port / "upstream_protocol_test.go").read_text()
        for name, field in [("upstream-text", "Prettier"), ("upstream-error", "EmbeddedError")]:
            with tempfile.TemporaryDirectory(prefix="tsprinter-upstream-mutant-") as directory:
                scratch = Path(directory)
                data = json.loads(json.dumps(pins))
                row = next(record for record in data["Records"] if record.get(field))
                row[field] += " mutant"
                pin = scratch / "pins.json"
                pin.write_text(json.dumps(data))
                side = scratch / "upstream_test.go"
                side.write_text(helper.replace("testdata/tsc-upstream-differences.json", str(pin)))
                overlay = scratch / "overlay.json"
                overlay.write_text(json.dumps({"Replace": {str(port / "upstream_protocol_test.go"): str(side)}}))
                log = options.logs / (name + ".log")
                with log.open("w") as output:
                    result = subprocess.run(["go", "test", "-v", "-count=1", "-overlay=" + str(overlay),
                                             "-run", "^TestTSCCorpusUpstreamDifferences$", "./stage1/cohere/tsprinter"],
                                            cwd=repository, stdout=output, stderr=subprocess.STDOUT)
                text = log.read_text()
                if result.returncode != 1 or row["Label"] not in text or "exit 1 stderr" in text:
                    raise RuntimeError(f"{name}: external outcome mutation escaped; see {log}")
                print(f"{name}: caught exact external outcome change at {row['Label']}; {log}", flush=True)
        return
    mutations = [
        ("yield-literal", "parser/parser.ts", "this.yieldContext || this.yieldOperandAhead()",
         "this.yieldContext || this.peek() === 'Identifier'", "expressions", "080_builtinIterator/builtinIterator.ts"),
        ("leading-block", "printer/expressions.ts",
         "if(parser.kind() === 'OpenBraceToken') return formatProgram(parser, source, settings);",
         "if(false) return formatProgram(parser, source, settings);", "expressions", "146_noImplicitAnyIndexing/noImplicitAnyIndexing.ts"),
        ("bodyless-declaration", "printer/syntax.ts",
         "const signatureLength = node.children.length - (bodylessDeclaration ? 0 : 1);",
         "const signatureLength = node.children.length - 1;", "statements", "065_callOverloads2/callOverloads2.ts"),
        ("statement-conditional", "printer/expressions.ts",
         "if(parent < 0 || parent === firstNonConditional) result = this.docs.group(result);",
         "if(parent === firstNonConditional) result = this.docs.group(result);", "statements", "114_truthinessPromiseCoercion/truthinessPromiseCoercion.ts"),
    ]
    for name, file, before, after, family, witness in ([] if options.roots_only else mutations):
        with tempfile.TemporaryDirectory(prefix="tsprinter-mutant-") as scratch:
            scratch = Path(scratch)
            for directory, source in [("parser", parser), ("printer", port)]:
                (scratch / directory).mkdir()
                for path in source.glob("*.ts"):
                    text = path.read_text()
                    text = text.replace("../scanner/", str(scanner) + "/")
                    text = text.replace("../../typescript/parser/", str(scratch / "parser") + "/")
                    (scratch / directory / path.name).write_text(text)
            target = scratch / file
            text = target.read_text()
            if text.count(before) != 1:
                raise RuntimeError(f"{name}: mutation must change one site")
            target.write_text(text.replace(before, after, 1))
            audit = (port / "tsc_corpus_test.go").read_text()
            needle = "port, _ := filepath.Abs(entry)"
            if audit.count(needle) != 1:
                raise RuntimeError("audit port path changed")
            audit = audit.replace(needle, "port := filepath.Join(" + json.dumps(str(scratch / "printer")) + ", entry)")
            side = scratch / "audit_test.go"
            side.write_text(audit)
            overlay = scratch / "overlay.json"
            overlay.write_text(json.dumps({"Replace": {str(port / "tsc_corpus_test.go"): str(side)}}))
            command = ["go", "test", "-v", "-count=1", "-timeout", "30m", "-overlay=" + str(overlay),
                       "./stage1/cohere/tsprinter", "-run", "^TestTSCCorpusAgreement/" + family + "$"]
            log = options.logs / (name + ".log")
            with log.open("w") as output:
                result = subprocess.run(command, cwd=repository, stdout=output, stderr=subprocess.STDOUT)
            text = log.read_text()
            file = "stage3/drivers/tsc/corpus/" + witness
            if result.returncode != 1 or any(side + " " + file not in text for side in ["Node", "native", "backend"]):
                raise RuntimeError(f"{name}: expected all three file-named byte failures; see {log}")
            if "exit 70" in text or "Lower:" in text or "Load:" in text or "AddressSanitizer" in text:
                raise RuntimeError(f"{name}: mutation did not finish normally; see {log}")
            print(f"{name}: caught on Node, native and backend by {file}; {log}", flush=True)

    roots = (port / "corpus_test.go").read_text()
    controls = [
        ("missing-root", roots.replace("if err != nil || !info.IsDir() {", "if info != nil && false {", 1)
         .replace("if len(files) == 0 {", "if false {", 1), "missing root was not named"),
        ("untracked-input", roots.replace("data, err := command.Output()",
         'data, err := command.Output()\n\tdata = append(data, []byte(root+"/untracked.ts\\x00")...)', 1), "untracked.ts"),
    ]
    for name, mutant, witness in controls:
        with tempfile.TemporaryDirectory(prefix="tsprinter-root-mutant-") as scratch:
            scratch = Path(scratch)
            side = scratch / "corpus_test.go"
            side.write_text(mutant)
            overlay = scratch / "overlay.json"
            overlay.write_text(json.dumps({"Replace": {str(port / "corpus_test.go"): str(side)}}))
            log = options.logs / (name + ".log")
            with log.open("w") as output:
                result = subprocess.run(["go", "test", "-v", "-count=1", "-overlay=" + str(overlay),
                                         "./stage1/cohere/tsprinter", "-run", "^TestTrackedCorpusRoots$"],
                                        cwd=repository, stdout=output, stderr=subprocess.STDOUT)
            if result.returncode != 1 or witness not in log.read_text():
                raise RuntimeError(f"{name}: root control escaped; see {log}")
            print(f"{name}: caught by TestTrackedCorpusRoots; {log}", flush=True)


if __name__ == "__main__":
    main()
