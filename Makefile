include Makefile.ledger


val = 1
exit: exit $(val)

all: lint install

install: go.sum
		go install $(BUILD_FLAGS) ./cmd/rd
		go install $(BUILD_FLAGS) ./cmd/rcli

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

prepare: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify
		GO111MODULE=on go mod vendor

lint:
	golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify

test:
	@echo "Start testing:"
	go test ./...

build: go.sum
	go build -mod=readonly $(BUILD_FLAGS) -o build/rd ./cmd/rd
	go build -mod=readonly $(BUILD_FLAGS) -o build/rcli ./cmd/rcli

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-docker-rdnode:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: build-linux localnet-stop
	@if ! [ -f build/node0/rd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/rd:Z tendermint/rdnode testnet --v 4 -o . --starting-ip-address 192.168.10.2 ;	fi
	docker-compose up -d

localnet-start-without-bls-keys: build-linux localnet-stop
	@if ! [ -f build/node0/rd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/rd:Z tendermint/rdnode testnet --v 4 -o . --starting-ip-address 192.168.10.2 --without-bls-keys ;	fi
	docker-compose up -d

localnet-start-with-dkg-in-10-blocks:
	@if ! [ -f build/node0/rd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/rd:Z tendermint/rdnode testnet --v 4 -o . --starting-ip-address 192.168.10.2 --dkg-num-blocks 10 ;	fi
		docker-compose up -d

# Stop testnet
localnet-stop:
	docker-compose down
.PHONY: test
