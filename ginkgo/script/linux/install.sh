TOOL_ROOT=$(dirname $(dirname $(dirname $(readlink -fm $0))))
echo ${TOOL_ROOT}

cd ${TOOL_ROOT}
go mod tidy
go mod download
go build -o /usr/local/bin/solar-ginkgo main.go

