#!/bin/bash
# Push script for chenjie222/engine fork
# Usage: ./push.sh

echo "Pushing to fork repository..."
export GIT_CONFIG_NOSYSTEM=1
git push origin master "$@"
echo "Done!"
