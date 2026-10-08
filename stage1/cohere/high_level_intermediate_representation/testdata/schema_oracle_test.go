//go:build lintoracle

package high_level_intermediate_representation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStage1SharedTerminalSchemas(t *testing.T) {
	t.Parallel()
	destination := os.Getenv("HIR_SHARED_SCHEMA_OUT")
	if destination == "" {
		t.Fatal("HIR_SHARED_SCHEMA_OUT is required")
	}
	write := func(name string, f *Function, ids []OracleExtraIdentity) {
		t.Helper()
		checkpoint := OracleWriteCheckpoint(OracleCheckpoint{Key: name, Pass: "shared-schema", Graph: oracleDump(f), Identities: append(OracleDormantIdentities(f), ids...), Sidecars: OracleInputFacts(f, nil)})
		if err := os.WriteFile(filepath.Join(destination, name+".checkpoint"), []byte(checkpoint), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, kind := range []string{"scope", "sequence", "catch", "escape"} {
		f := NewFunction(nil, "f", FunctionKindOther)
		entry := f.NewBlock(BlockKindBlock)
		body := f.NewBlock(BlockKindBlock)
		exit := f.NewBlock(BlockKindBlock)
		f.Entry = entry.Id
		f.Returns = Place{Identifier: f.NewIdentifier("", nil, 0).Id}
		body.Terminal = &Goto{Block: exit.Id}
		exit.Terminal = &Return{Value: f.Returns}
		var ids []OracleExtraIdentity
		switch kind {
		case "scope":
			entry.Terminal = &Scope{Scope: 17, Block: body.Id, Fallthrough: exit.Id}
			ids = []OracleExtraIdentity{{FunctionPath: "$", Kind: "scope", Id: 17}}
		case "sequence":
			entry.Terminal = &Sequence{Block: body.Id, Fallthrough: exit.Id}
		case "catch":
			entry.Terminal = &MaybeThrow{Continuation: body.Id, Handler: exit.Id}
		case "escape":
			entry.Terminal = &MaybeThrow{Continuation: body.Id, Handler: InvalidBlock}
		}
		write(kind+"-before", f, ids)
		Finalize(f)
		write(kind+"-after", f, ids)
	}
	f := lowerSource(t, "function f(x){let y=0;if(x){y=1;}return y;}")
	Construct(f)
	write("phi-before", f, nil)
	Finalize(f)
	write("phi-after", f, nil)
}
