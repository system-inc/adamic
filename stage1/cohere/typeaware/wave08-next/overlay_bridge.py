"""Create an owned build overlay; shared dispatcher source remains unchanged."""
import pathlib,json,argparse
parser=argparse.ArgumentParser();parser.add_argument('directory');args=parser.parse_args()
source=pathlib.Path(__file__).resolve().parent;repository=source.parents[3];root=pathlib.Path(args.directory).resolve();root.mkdir(parents=True,exist_ok=True)
original=repository/'bridge/tsgo/checker/facts.go';text=original.read_text();needle='import (';assert text.count(needle)==1
text=text.replace(needle,needle+'\n wave08next "github.com/system-inc/adamic/stage1/cohere/typeaware/wave08-next"',1)
needle='switch mode {';assert text.count(needle)==1
text=text.replace(needle,needle+'''
 case "wave08-resolved-callee", "wave08-program-loads", "wave08-symbol-ancestry":
  var values []string
  var failure error
  if mode != question { return "", fmt.Errorf("unexpected wave08 fact suffix") }
  if mode == "wave08-resolved-callee" { values, failure = wave08next.ResolvedCalleeFields(c, node) }
  if mode == "wave08-program-loads" { values, failure = wave08next.ProgramLoadsFields(p.Compiler, node) }
  if mode == "wave08-symbol-ancestry" { values = wave08next.SymbolAncestryFields(c, node) }
  if failure != nil { return "", failure }
  for _, value := range values[2:] { out.text(value) }
''',1)
replacement=root/'facts.go';replacement.write_text(text);(root/'overlay.json').write_text(json.dumps({'Replace':{str(original):str(replacement)}}))
