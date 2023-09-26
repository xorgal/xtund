#!/bin/bash

BINARY="xtund"
REMOTE="test"
UPLOAD_DIR="tmp"

"$(dirname "$0")/build.sh"

rsync -v "$(dirname "$0")/../bin/$BINARY" $REMOTE:~/$UPLOAD_DIR
