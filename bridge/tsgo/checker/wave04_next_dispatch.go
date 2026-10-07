package checker

import (
	"context"
	"fmt"
	"strings"
)

// InspectWave04Next is an isolated adapter for this worker's checker questions.
// Existing questions retain the production dispatch, indexing and lease behavior.
func (p *Program) InspectWave04Next(file string, start, end uint64, kind, question string) (string, error) {
	if !strings.HasPrefix(question, "wave04-next-") {
		return p.Inspect(file, start, end, kind, question)
	}
	source, node, err := p.exact(file, start, end, kind)
	if err != nil {
		return "", err
	}
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	out := &fields{}
	out.number(1)
	out.text(strings.Split(question, "\n")[0])
	switch question {
	case "wave04-next-source-context":
		return p.wave04NextSourceContext(out, node, question)
	case "wave04-next-program-modules":
		return p.wave04NextProgramModules(out, node, question)
	case "wave04-next-declaration-context":
		return p.wave04NextDeclarationContext(out, c, node, question)
	case "wave04-next-resolved-signature-declaration":
		return p.wave04NextResolvedSignatureDeclaration(out, c, node, question)
	default:
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
}
