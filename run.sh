#!/bin/bash

# chmod +x run.sh

FIRST_ARG=$1
SECOND_ARG=$2

if [ $# -eq 0 ]; then
    echo "
~ Available commands
    pooler          # Start the pooler
    api             # Start the API on localhost:4444
    ports           # List all ports in use
    cloc            # Count lines of code
    test go         # Run Go tests"
fi

case $FIRST_ARG in

"pooler")
    echo " * Starting the pooler"
    cd syro && reflex -r '\.go' -s -- sh -c "go run cmd/pooler/main.go"
    ;;

"api")
    echo " * Starting the api"
    cd syro && reflex -r '\.go' -s -- sh -c "go run cmd/api/main.go"
    ;;

"exec")
    cd syro && go run cmd/exec/main.go
    ;;

"flamegraph")
    cd syro
    go build -o out cmd/flamegraph/main.go
    ./out

    # create and open the flamegraph svg
    go tool pprof -raw -output=cpu.txt ./out cpu.prof
    stackcollapse-go.pl cpu.txt | flamegraph.pl > flame.svg && open flame.svg
    ;;

"test")
    case $SECOND_ARG in
    "go")
        echo " * Running Go tests"
        #   -v          means verbose (print the output to the console)
        #   -count=1    means run the tests only once (don't cache the results)
        cd syro
        GO_TEST_MODE=full GO_CONF_PATH="$(pwd)/conf/config.dev.toml" go test ./... -count=1
        ;;
    esac
    ;;

"cloc")
    # find . -name '*.go' | xargs wc -l | sort -nr
    # |_test\.go
    # --not-match-f='\.js$|\.txt$|\.sh$|\.csv$|\.parquet$|\.xsl$|\.xslx$|\.md$|controllers\.gen\.go$|\.yml$|\.d.ts$|\.ipynb$' \
    cloc \
        --not-match-f='\.js$|\.txt$|\.sh$|\.csv$|\.parquet$|\.xsl$|\.xslx$|\.md$|controllers\.gen\.go$|\.yml$|\.d.ts$|\.ipynb$|_test\.go' \
        --exclude-dir=out \
        --exclude-dir=dist \
        --exclude-dir=.vscode \
        --exclude-dir=swagger \
        --exclude-dir=backup \
        --exclude-dir=docs \
        --exclude-dir=node_modules \
        --exclude-dir=scripts \
        --exclude-dir=py \
        --exclude-dir=.venv \
        .
    ;;

"ports")
    sudo lsof -i -P -n | grep LISTEN
    ;;

esac
