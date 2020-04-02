include Makefile.ledger

VALIDATORS_COUNT ?= 4
val = 1
exit: exit $(val)

all: lint install

install: go.sum
		go install $(BUILD_FLAGS) ./cmd/randappd
		go install $(BUILD_FLAGS) ./cmd/randappcli

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
	go build -mod=readonly $(BUILD_FLAGS) -o build/randappd ./cmd/randappd
	go build -mod=readonly $(BUILD_FLAGS) -o build/randappcli ./cmd/randappcli

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-docker-randappnode:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: build-linux localnet-stop
	@if ! [ -f build/node0/randappd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/randappd:Z tendermint/randappnode testnet --v $(VALIDATORS_COUNT) -o . --starting-ip-address 192.168.10.2 --keyring-backend=test ;	fi
	docker-compose up -d

localnet-start-without-bls-keys: build-linux localnet-stop
	@if ! [ -f build/node0/randappd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/randappd:Z tendermint/randappnode testnet --v $(VALIDATORS_COUNT) -o . --starting-ip-address 192.168.10.2 --without-bls-keys --keyring-backend=test ;	fi
	docker-compose up -d

localnet-start-with-dkg-in-5-blocks: build-linux localnet-stop
	@if ! [ -f build/node0/randappd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/randappd:Z tendermint/randappnode testnet --v 4 -o . --starting-ip-address 192.168.10.2 --dkg-num-blocks 5 --keyring-backend=test ;	fi
		docker-compose up -d

# Stop testnet
localnet-stop:
	docker-compose down
.PHONY: test