"""Create an owned build overlay; shared dispatcher source remains unchanged."""
import pathlib,json,argparse
parser=argparse.ArgumentParser();parser.add_argument('directory');args=parser.parse_args()
source=pathlib.Path(__file__).resolve().parent;repository=source.parents[3];root=pathlib.Path(args.directory).resolve();root.mkdir(parents=True,exist_ok=True)
original=repository/'bridge/tsgo/checker/facts.go';text=original.read_text();needle='import (';assert text.count(needle)==1
text=text.replace(needle,needle+'\n wave08next "github.com/system-inc/adamic/stage1/cohere/typeaware/wave08-next"\n wave08core "github.com/system-inc/adamic/stage1/cohere/typeaware/wave08-core-next"',1)
needle='switch mode {';assert text.count(needle)==1
text=text.replace(needle,needle+'''
 case "wave08-await-contract", "wave08-await-type":
  g := &graph{program:p, checker:c, seen:make(map[*checker.Type]bool)}
  var values []string
  if mode == "wave08-await-contract" {
   if mode != question { return "", fmt.Errorf("unexpected await contract suffix") }
   values = wave08core.AwaitContractFields(c, node, g.id)
  } else {
   split := strings.Split(question, "\\n")
   if len(split) != 3 { return "", fmt.Errorf("await type needs identity and property") }
   id, err := strconv.ParseUint(split[1], 10, 64)
   if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id,10) != split[1] { return "", fmt.Errorf("unknown await type identity") }
   values = wave08core.AwaitTypeFields(c, node, p.typesByID[id-1], split[2], g.id)
  }
  for _,value := range values[2:] { out.text(value) }
 case "wave08-resolved-callee", "wave08-program-loads", "wave08-symbol-ancestry", "wave08-symbol-origins", "wave08-syntax-snapshot":
  var values []string
  var failure error
  if mode != question { return "", fmt.Errorf("unexpected wave08 fact suffix") }
  if mode == "wave08-resolved-callee" { values, failure = wave08next.ResolvedCalleeFields(c, node) }
  if mode == "wave08-program-loads" { values, failure = wave08next.ProgramLoadsFields(p.Compiler, node) }
  if mode == "wave08-symbol-ancestry" { values = wave08next.SymbolAncestryFields(c, node) }
   if mode == "wave08-symbol-origins" { values = wave08core.SymbolOriginsFields(c, node) }
  if mode == "wave08-syntax-snapshot" { values = wave08core.SyntaxSnapshotFields(c, node, p.symbolID) }
  if failure != nil { return "", failure }
  for _, value := range values[2:] { out.text(value) }
''',1)
replacement=root/'facts.go';replacement.write_text(text);(root/'overlay.json').write_text(json.dumps({'Replace':{str(original):str(replacement)}}))
