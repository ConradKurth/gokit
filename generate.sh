#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# Find all directories containing at least one prototfile.
# Based on: https://buf.build/docs/migration-prototool#prototool-generate.
for dir in $(find ${DIR}/protos -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq); do
  files=$(find "${dir}" -name '*.proto')
  echo $files
  if [[ $files == *'grpc'* ]]; then
    protoc -I ${DIR} --go-grpc_out=paths=source_relative:${DIR}/protos/gen/go ${files}
    continue
  fi
  protoc -I ${DIR}  --go_out=paths=source_relative:${DIR}/protos/gen/go ${files}
done
