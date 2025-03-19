#!/bin/bash

printf "Start\n";

while read line; do
  echo "PROCESS event: $line" >&2;
  kill -3 $(cat "/var/run/supervisord.pid")
done < /dev/stdin
