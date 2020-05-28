#!/bin/bash
CWD=$(cd "$(dirname $0)";pwd)
"$CWD"/goeasyffmpeg stop
"$CWD"/goeasyffmpeg uninstall