#!/bin/bash
TEST_LIST=$(go list ./... | grep -v -e "sample" | grep -v -e "examples" | grep -e "^github")
# shellcheck disable=SC2086
gotestsum --junitfile ~/junit/unit-tests.xml -- -short $TEST_LIST
