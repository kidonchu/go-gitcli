#!/bin/bash

VERSION="$1"
if [[ -z "$VERSION" ]]; then
	echo "ERROR: Version number not specified"
	exit 1
fi

go install
zip gitcli_${VERSION}_MacOS-64bit.zip ${GOPATH}/bin/gitcli
