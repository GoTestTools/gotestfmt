#!/bin/bash
gofmt -l . | grep -v _testsource >/tmp/gofmt.files
if [ "$(cat /tmp/gofmt.files | wc -l)" -ne 0 ]; then
  echo -e "::group::\e[0;31m❌ One or more files are not formatted properly\e[0m"
  cat /tmp/gofmt.files | xargs -i echo "   📝 {}"
  echo "::endgroup::"
  exit 1
else
  echo -e "\e[0;32m✅ All files are formatted correctly\e[0m"
fi