#!/usr/bin/env bash

current_branch=$(git rev-parse --abbrev-ref HEAD)
echo "current branch is $current_branch"

echo "pull down code..."
if [[ "$current_branch" == "master" ]]; then
    git pull
else
    echo "Reset to HEAD and pull $current_branch from remote..."
    git reset --hard HEAD
    git checkout develop
    git branch -D "$current_branch"
    git pull
    git checkout "$current_branch"
fi
