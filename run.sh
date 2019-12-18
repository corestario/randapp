#!/usr/bin/env bash

rm -rf ~/.r*
rm $GOPATH/bin/rcli
rm $GOPATH/bin/rd
make install

# Initialize configuration files and genesis file
rd init --chain-id rchain validator0

rcli keys add validator0 <<< "12345678"
echo $(rcli keys show validator0 -a)
rcli keys add validator1 <<< "12345678"
rcli keys add validator2 <<< "12345678"
rcli keys add validator3 <<< "12345678"
#rcli keys add validator4 <<< "12345678"

# Add both accounts, with coins to the genesis file
#rd add-genesis-account $(rcli keys show validator0 -a) 1000nametoken,1000validator0coin
#rd add-genesis-account $(rcli keys show validator1 -a) 1000nametoken,1000validator1coin
#rd add-genesis-account $(rcli keys show validator2 -a) 1000nametoken,1000validator2coin
#rd add-genesis-account $(rcli keys show validator3 -a) 1000nametoken,1000validator3coin

rd add-genesis-account $(rcli keys show validator0 -a) 1000nametoken,100000000stake
rd add-genesis-account $(rcli keys show validator1 -a) 1000nametoken,100000000stake
rd add-genesis-account $(rcli keys show validator2 -a) 1000nametoken,100000000stake
rd add-genesis-account $(rcli keys show validator3 -a) 1000nametoken,100000000stake
#rd add-genesis-account $(rcli keys show validator4 -a) 1000nametoken,100000000stake


# Configure your CLI to eliminate need for chain-id flag
rcli config chain-id rchain
rcli config output json
rcli config indent true
rcli config trust-node true

cp -r ~/.rcli ~/.rcli0
cp -r ~/.rcli ~/.rcli1
cp -r ~/.rcli ~/.rcli2
cp -r ~/.rcli ~/.rcli3

chmod +w ~/.rd/config

rd gentx --name validator0 <<< "12345678"

rd collect-gentxs

rd validate-genesis

rd start