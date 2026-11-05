import json,pathlib,sys
base,output=map(pathlib.Path,sys.argv[1:3]); output.mkdir(parents=True,exist_ok=True)
meta=json.loads(base.read_text()); key=next(k for k in meta['Replace'] if k.endswith('/internal/lower/latent_hook.go')); text=pathlib.Path(meta['Replace'][key]).read_text(); text=text.replace('import (','import (\n "strings"',1)
needle='for _, file := range program.Files() {\n\t\tlatentFindings = nil'; assert text.count(needle)==1
text=text.replace(needle,'for _, file := range program.Files() {\n        if os.Getenv("LATENT_SELECT_NAMES") != "" && !strings.HasSuffix(program.FileName(file), "/src/compiler/utilitiesPublic.ts") { continue }\n\t\tlatentFindings = nil',1)
needle='for _, node := range file.Statements.Nodes {\n\t\t\tlatentFindingOwner'; assert text.count(needle)==1
text=text.replace(needle,'for _, node := range file.Statements.Nodes {\n        if selected := os.Getenv("LATENT_SELECT_NAMES"); selected != "" {\n            if node.Kind != ast.KindFunctionDeclaration || node.Name() == nil || !strings.Contains("|"+selected+"|", "|"+node.Name().Text()+"|") { continue }\n            fmt.Fprintf(os.Stderr, "SELECTED_BEGIN %s %s\\n", program.Where(node), node.Name().Text())\n        }\n\t\t\tlatentFindingOwner',1)
path=output/'selected_hook.go'; path.write_text(text); meta['Replace'][key]=str(path.resolve()); (output/'overlay.json').write_text(json.dumps(meta,indent=2)+'\n')
