# DGaming RandApp

DGaming RandApp is Cosmos-based SDK, designed for easier building blockchain 
applications using on-chain DKG algorithm in Golang.


More about [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).


Notable source code files to check out about DKG:

1. https://github.com/dgamingfoundation/tendermint/blob/dcr-random/types/random.go
2. https://github.com/dgamingfoundation/tendermint/blob/dcr-random/consensus/dkg.go
3. https://github.com/dgamingfoundation/tendermint/blob/dcr-random/consensus/dkg_dealer.go


On-chain DKG works the same way as the off-chain version but writes its messages to blocks, 
which allows us to slash a validator that refuses to participate in a DKG round.

More about DKG and lib, which implements corresponding algorithm 
see [dkglib](https://github.com/dgamingfoundation/dkglib).

# Running a node

To build and run a node with a script:

```bash
./run.sh
```

## Step by step run
(Re)build and (re)install node and client:
```
rm -rf ~/.r*
rm $GOPATH/bin/rcli
rm $GOPATH/bin/rd
make install
```

Initialize configuration files and genesis file:
```
rd init --chain-id [chain_id]
```
Add validator:
```
rcli keys add [validator_name]:
```
Add both accounts, with coins to the genesis file:
```
rd add-genesis-account $(rcli keys show [validator_name] -a) 1000nametoken,100000000stake
```
Configure your CLI to eliminate need for chain-id flag:
```
rcli config chain-id [chain_id]
rcli config output json
rcli config indent true
rcli config trust-node true
```
Start the node:
```
chmod +w ~/.rd/config
rd start
```
