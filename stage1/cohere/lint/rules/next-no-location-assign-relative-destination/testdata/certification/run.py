"""Run the shared harness with additional per-rule native mutant certification."""
from pathlib import Path
import json, subprocess, tempfile, sys
root = Path(__file__).resolve().parents[7]
source = Path(__file__).resolve().with_name('native_test.go')
with tempfile.TemporaryDirectory(prefix='nextjs-certification-') as scratch:
    overlay = Path(scratch) / 'overlay.json'
    virtual = root / 'stage1/cohere/lint/nextjs_certification_test.go'
    mounted = Path(scratch) / 'native_test.go'
    mounted.write_text(source.read_text().replace('//go:build lintoracle\n', '', 1))
    overlay.write_text(json.dumps({'Replace': {str(virtual): str(mounted)}}))
    result = subprocess.run(['go', 'test', '-overlay='+str(overlay), './stage1/cohere/lint', '-run', 'TestNextjsCertification', '-count=1', '-v', '-timeout', '60m', *sys.argv[1:]], cwd=root)
    sys.exit(result.returncode)
