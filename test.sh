set -ex

SERVER_PORT=$(./bin/random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(./bin/random tempfile)

./bin/metricstest -test.v -test.run=^TestIteration1$ \
    -binary-path=./bin/server

./bin/metricstest -test.v -test.run=^TestIteration2[AB]*$ \
    -source-path=. \
    -agent-binary-path=./bin/agent

./bin/metricstest -test.v -test.run=^TestIteration3[AB]*$ \
    -source-path=. \
    -agent-binary-path=./bin/agent \
    -binary-path=./bin/server

./bin/metricstest -test.v -test.run=^TestIteration4$ \
    -agent-binary-path=./bin/agent \
    -binary-path=./bin/server \
    -server-port=$SERVER_PORT \
    -source-path=.

./bin/metricstest -test.v -test.run=^TestIteration5$ \
    -agent-binary-path=./bin/agent \
    -binary-path=./bin/server \
    -server-port=$SERVER_PORT \
    -source-path=.

./bin/metricstest -test.v -test.run=^TestIteration6$ \
    -agent-binary-path=./bin/agent \
    -binary-path=./bin/server \
    -server-port=$SERVER_PORT \
    -source-path=.

./bin/metricstest -test.v -test.run=^TestIteration7$ \
    -agent-binary-path=./bin/agent \
    -binary-path=./bin/server \
    -server-port=$SERVER_PORT \
    -source-path=.

./bin/metricstest -test.v -test.run=^TestIteration8$ \
    -agent-binary-path=./bin/agent \
    -binary-path=./bin/server \
    -server-port=$SERVER_PORT \
    -source-path=.

./bin/metricstest -test.v -test.run=^TestIteration9$ \
    -agent-binary-path=./bin/agent \
    -binary-path=./bin/server \
    -file-storage-path=$TEMP_FILE \
    -server-port=$SERVER_PORT \
    -source-path=.
