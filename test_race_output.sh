#!/usr/bin/env sh

i=1
while [ "$i" -le 10000 ]; do
  echo "# $i - execution"
  go run -race tripper.go
  i=$(($i+1))
done
