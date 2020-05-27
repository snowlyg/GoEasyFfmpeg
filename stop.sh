#!/bin/bash
CWD=$(cd "$(dirname $0)";pwd)
"$CWD"/easydarwin stop
"$CWD"/easydarwin uninstall
"$CWD"/rtspserver stop
"$CWD"/rtspserver uninstall