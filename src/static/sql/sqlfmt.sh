#!/bin/env bash

SQLDIR=$(dirname -- "$(readlink -f -- "$0")")

for filename in $(find $SQLDIR -type f -name "*.sql"); do
    if [[ $filename == *"/backup.sql" ]]; then
        continue
    fi
    sqlfmt --input $filename --output $filename --newlines --upper --spaces 4 --comment-pre-space
done
