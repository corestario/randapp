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

# Running a local-testnet

#### Ensure that you have docker!
#### Ensure that dkglib folder is next to your randapp folder!
##### It is necessary for building docker images and starting testnet! 

Example:

/ home

**/ projects

****/ randapp

****/ dkglib 
 

Run from randapp folder:
```bash
./testnet.sh
```
You might need to run it as superuser.

Flags:
```
      -h, --help                    show brief help
      -n, --node_count=n            specify node count
      --no_rebuild                  run without rebuilding docker images"
      --kill                        stop and remove testnet containers; remove additional files"
      --ruin                        force stop containers 1 and 2 after 5 seconds running dkg
```

Application logs are stored in docker container in ```/root/``` folder.

randapp - ```/root/rd_start.log```

dkglib - ```/root/dkglib.log```

To access logs use

```
docker exec ContainerID /bin/bash -c "cat /root/rd_start.log"
```

and

```
docker exec ContainerID /bin/bash -c "cat /root/dkglib.log"
```


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
