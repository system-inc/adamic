#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys
owned = Path(__file__).resolve().parent
scratch = Path(sys.argv[1]).resolve()
subprocess.run([sys.executable, str(owned.parent / 'typescript-no-restricted-types/prepare-validation.py'), str(scratch)], check=True)
target = scratch / 'stage1/cohere/lint/lint_test.go'
text = target.read_text()
anchor = 'path := filepath.Join(directory, fmt.Sprintf("case-%03d%s", i, filepath.Ext(row.File)))'
replacement = '''path := filepath.Join(directory, fmt.Sprintf("case-%03d", i), strings.TrimPrefix(strings.ReplaceAll(row.File, "\\\\", "/"), "/"))
        if err := os.MkdirAll(filepath.Dir(path),0755); err != nil { t.Fatal(err) }'''
assert anchor in text
text = text.replace(anchor, replacement)
text = text.replace('key := fmt.Sprintf("%s\\t%+v\\t%s", row.Rule, row.Options, row.Source)', 'key := fmt.Sprintf("%s\\t%+v\\t%s\\t%s", row.Rule, row.Options, row.Source, row.File)')
text += (owned / 'validation.go.txt').read_text()
target.write_text(text)
