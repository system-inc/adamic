#!/usr/bin/env python3
"""Controlled child for exit-code propagation, never a compiler input."""
import sys,time
if sys.argv[1]=='wait':time.sleep(10)
else:sys.exit(int(sys.argv[1]))
