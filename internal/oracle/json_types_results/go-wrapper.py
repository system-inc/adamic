#!/usr/bin/env python3
import os,sys,json
with open(os.environ["JSON_GO_LOG"],"a") as f:f.write(json.dumps(sys.argv[1:])+"\n")
os.execv('/workspace/adamic-tools/go/bin/go',sys.argv)
