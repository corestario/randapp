# CoreStarIO RandApp

CoreStarIO RandApp is Cosmos-based SDK, designed for easier building blockchain 
applications using on-chain DKG algorithm in Golang.


More about [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).


Notable source code files to check out about DKG:

1. https://github.com/corestario/tendermint/blob/dcr-random/types/random.go
2. https://github.com/corestario/tendermint/blob/dcr-random/consensus/dkg.go
3. https://github.com/corestario/tendermint/blob/dcr-random/consensus/dkg_dealer.go


On-chain DKG works the same way as the off-chain version but writes its messages to blocks, 
which allows us to slash a validator that refuses to participate in a DKG round.

More about DKG and lib, which implements corresponding algorithm 
see [dkglib](https://github.com/corestario/dkglib).

# Requirements:
* [Tendermint (corestario fork)](https://github.com/corestario/tendermint)
* [DKGLib](https://github.com/corestario/dkglib)
* [Cosmos-utils](https://github.com/corestario/cosmos-utils)
* [Cosmos-sdk (corestario fork)](https://github.com/corestario/cosmos-sdk)
* [Cosmos-modules (corestario fork)](https://github.com/corestario/modules)

# Running a local-testnet

#### Ensure that you have docker!
#### Ensure that dkglib, tendermint, cosmos-sdk and cosmos-utils, modules folders are next to your randapp folder!
##### It is necessary for building docker images and starting testnet! 

Example:

/ home

**/ projects

****/ randapp

****/ dkglib

****/ tendermint

****/ cosmos-sdk

****/ cosmos-utils

****/ modules

#### Run a local-testnet:
```shell script
sudo rm -rf ./build && sudo make build-docker-rdnode && make build-linux && sudo make localnet-stop && sudo make localnet-start-without-bls-keys
```

#### How to view logs on a randapp node:
```shell script
docker logs -f rdnode0
```