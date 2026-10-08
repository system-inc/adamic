#!/usr/bin/env python3
"""Build census-only Go overlay. Usage: build.py [output] [go executable]."""
import pathlib,json,tempfile,subprocess,sys
base=pathlib.Path(__file__).resolve().parent
root=base.parents[3];virtual=root/'cohere/TypeScript/tsc/adamic_parser_census.go'
with tempfile.NamedTemporaryFile(mode='w',suffix='.json') as overlay:
 json.dump({'Replace':{str(virtual):str(base/'oracle.go.txt')}},overlay);overlay.flush()
 subprocess.run([sys.argv[2] if len(sys.argv)>2 else 'go','build','-overlay='+overlay.name,'-o',sys.argv[1] if len(sys.argv)>1 else '/tmp/parser-census-oracle',str(virtual)],cwd=root,check=True)
