gopath = .
buildPath = ./bin
projectPath = ${gopath}

all:build

build:
	go build -o ${buildPath}/crawler -ldflags -w ${projectPath}/cmd/crawler
	go build -o ${buildPath}/indexer -ldflags -w ${projectPath}/cmd/indexer
	go build -o ${buildPath}/migrator -ldflags -w ${projectPath}/cmd/migrator
	go build -o ${buildPath}/main -ldflags -w ${projectPath}/cmd/main
		
buildsrc:
	go build -o ${buildPath}/crawler   ${projectPath}/cmd/crawler
	go build -o ${buildPath}/indexer   ${projectPath}/cmd/indexer
	go build -o ${buildPath}/migrator   ${projectPath}/cmd/migrator
	go build -o ${buildPath}/main   ${projectPath}/cmd/main

clean:
	rm -rf ${buildPath}/crawler ${buildPath}/indexer ${buildPath}/migrator ${buildPath}/main
