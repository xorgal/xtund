#!/bin/bash

BINARY="xtund"
REMOTE="gate"
UPLOAD_DIR="."

"$(dirname "$0")/build.sh"

rsync -v "$(dirname "$0")/../bin/$BINARY" $REMOTE:~/$UPLOAD_DIR
