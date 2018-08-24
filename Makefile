gopath = .
buildPath = ./bin
projectPath = ${gopath}

all:build

.PHONY: help clean cleanall html pdf deps deploy

help:
	@echo "Please use 'make <target>' where <target> is one of"
	@echo "  docker  Docker image based on centos"
	@echo "  linux   Compile executables and files based on Linux"
	@echo "  build  Build based on source code"
	@echo "  buildsrc  Build based on source code"
	@echo "  clean   Delete executable files in bin directory"
	@echo
	@echo " use 'make <target>'"


docker:linux
	docker build -t sander .

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${buildPath}/crawler -ldflags -w ${projectPath}/cmd/crawler
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${buildPath}/indexer -ldflags -w ${projectPath}/cmd/indexer
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${buildPath}/migrator -ldflags -w ${projectPath}/cmd/migrator
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${buildPath}/main -ldflags -w ${projectPath}/cmd/main

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

