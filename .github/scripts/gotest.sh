#!/bin/bash
go test -v ./... >/tmp/gotest.txt 2>/tmp/depdownload.txt
RESULT=$?
if [ "$(cat /tmp/depdownload.txt | wc -l)" -ne 0 ]; then
  echo -e "::group::\e[0;34m📥 Dependency downloads\e[0m"
  cat /tmp/depdownload.txt | sed -e "s/go: downloading /   📦 /"
  echo "::endgroup::"
fi
cat /tmp/gotest.txt | grep -v "go: downloading"
exit $RESULT