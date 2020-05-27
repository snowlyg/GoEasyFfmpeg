#!/bin/bash
CWD=$(cd "$(dirname $0)";pwd)
"$CWD"/easydarwin stop
"$CWD"/easydarwin uninstall
"$CWD"/rtsp-simple-server-master stop
"$CWD"/rtsp-simple-server-master uninstall