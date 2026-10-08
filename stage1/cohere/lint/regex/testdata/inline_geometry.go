package core

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"unicode/utf16"
)

func waveInlineGeometry(ctx rule.Context, list []comments.Comment, pattern string) {
	path := os.Getenv("ADAMIC_INLINE_GEOMETRY")
	if path == "" {
		return
	}
	source := ctx.SourceFile.Text()
	rows := []map[string]any{}
	for _, c := range list {
		rows = append(rows, map[string]any{"start": len(utf16.Encode([]rune(source[:c.Range.Pos()]))), "end": len(utf16.Encode([]rune(source[:c.Range.End()]))), "text": c.Text, "block": c.IsBlock, "emptyJsx": noInlineCommentsIsInsideEmptyJsxExpression(ctx, &c)})
	}
	record := map[string]any{"kind": int(ctx.SourceFile.AsNode().Kind), "source": source, "pattern": pattern, "comments": rows}
	waveRegexLock.Lock()
	defer waveRegexLock.Unlock()
	f, e := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if e := json.NewEncoder(f).Encode(record); e != nil {
		panic(e)
	}
}
