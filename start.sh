#!/bin/bash
CWD=$(cd "$(dirname $0)";pwd)
"$CWD"/goeasyffmpeg install
"$CWD"/goeasyffmpeg start