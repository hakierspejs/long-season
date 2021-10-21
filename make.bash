#!/bin/bash
#
# make.bash is a simple and portable utility script
# for running development tasks.
#
# Run "./make.bash help" for list of commands.

SERVER_NAME="long-season"
CLI_NAME="short-season"
FILENAME=$0

ls_cmd() {
    CMD="$*"
    echo "$CMD"
    $CMD
    STATUS=$?
    if [ ! $STATUS -eq 0 ]; then
        echo "command \"$CMD\" exited with $STATUS"
        exit 1
    fi
}

ls_test() {
    ls_cmd go test -race ./...
}

ls_build() {
    ls_cmd go build -o $SERVER_NAME cmd/long-season/*.go
    ls_cmd go build -o $CLI_NAME cmd/short-season/*.go
}

ls_lint() {
    ls_cmd go vet ./...
}

ls_clean() {
    ls_cmd rm -rf $SERVER_NAME $CLI_NAME
}

ls_run() {
    ls_build
    ls_cmd "./$SERVER_NAME"
}

ls_watch() {
    git ls-files | entr -r ./$FILENAME run
}

print_help() {
    echo "Usage: $FILENAME [help|build|clean|lint|test|run|watch]"
}

case $1 in
    "build")
        ls_build
        ;;
    "clean")
        ls_clean
        ;;
    "lint")
        ls_lint
        ;;
    "test")
        ls_test
        ;;
    "run")
        ls_run
        ;;
    "watch")
        ls_watch
        ;;
    ""|"help")
        print_help
        ;;
    *) echo "Doesn't recognize \"$1\". Use help command for more info."
        ;;
esac
