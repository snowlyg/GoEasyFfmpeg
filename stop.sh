#!/bin/bash
CWD=$(cd "$(dirname $0)";pwd)
"$CWD"/easydarwin stop
"$CWD"/easydarwin uninstall
#"$CWD"/rtsp-server/rtsp-server stop
#"$CWD"/rtsp-server/rtsp-server uninstall