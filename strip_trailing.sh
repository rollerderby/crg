#!/bin/bash

find . -type f -name '*.go' -exec sed --in-place 's/[[:space:]]\+$//' {} \+
find . -type f -name '*.html' -exec sed --in-place 's/[[:space:]]\+$//' {} \+
find . -type f -name '*.js' -exec sed --in-place 's/[[:space:]]\+$//' {} \+
find . -type f -name '*.css' -exec sed --in-place 's/[[:space:]]\+$//' {} \+
