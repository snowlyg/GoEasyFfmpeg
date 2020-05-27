#!/bin/bash
CWD=$(cd "$(dirname $0)";pwd)
"$CWD"/easydarwin install
"$CWD"/easydarwin start
"$CWD"/rtsp-simple-server-master install
"$CWD"/rtsp-simple-server-master start