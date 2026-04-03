#!/bin/sh
set -e

BASE="$HOME/pjf-test-repos"
rm -rf "$BASE"

groups="frontend backend infrastructure data-science mobile devops"

for group in $groups; do
    for i in $(seq 1 8); do
        dir="$BASE/$group/very-long-project-name-$group-service-$i"
        mkdir -p "$dir"
        git -C "$dir" init -q
        git -C "$dir" checkout -q -b "feature/really-long-branch-name-for-testing-$i" 2>/dev/null || true
        touch "$dir/go.mod"
    done
done

echo "Created 48 test repos in $BASE"
echo "Add $BASE to pjf scan dirs to test"
