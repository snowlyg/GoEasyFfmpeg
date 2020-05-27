#!/bin/bash
CWD=$(cd "$(dirname $0)";pwd)
"$CWD"/easydarwin install
"$CWD"/easydarwin start
"$CWD"/rtsp-server install
"$CWD"/rtsp-server start