include Makefile.ledger
all: install
install: go.sum
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/rd
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/rcli
go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		#GO111MODULE=on go mod verify

BUILD_TAGS?='rd'
OUTPUT?=build/rd
OUTPUTCLI?=build/rcli

get_tools:
	@echo "--> Installing tools"
	./scripts/get_tools.sh

build-linux: get_tools
	GOOS=linux GOARCH=amd64 $(MAKE) build

build-docker-localnode:
	@cd networks/local && make

build:
	CGO_ENABLED=0 go build -tags $(build_tags) -o $(OUTPUT) ./cmd/rd/
	CGO_ENABLED=0 go build -tags $(build_tags) -o $(OUTPUTCLI) ./cmd/rcli/

localnet-start:
	@if ! [ -f build/node0/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/rd:Z randapp/localnode testnet --o . --populate-persistent-peers --starting-ip-address 192.167.10.2 ; mv $(CURDIR)/build/docker-compose.yml $(CURDIR) ; fi
	make localnet-stop
	cp ./run.sh build/
	docker-compose up

# Stop testnet
localnet-stop:
	docker-compose down
