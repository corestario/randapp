#!/usr/bin/env bash

randapp_path=/go/src/github.com/dgamingfoundation/randapp
num=$1
rm -rf $HOME/.rcli
rm -rf $HOME/.rd

rd init moniker --chain-id rchain

rcli config chain-id rchain
rcli config output json
rcli config indent true
rcli config trust-node true

mkdir -p $HOME/.rd/config

cp $HOME/tmp/genesis.json $HOME/.rd/config
cp $HOME/tmp/config.toml $HOME/.rd/config
cp -r $HOME/tmp/.rcli $HOME/.rcli
cp -r $HOME/tmp/.rcli $HOME/.rcli${num}

sed -i 's/moniker = "moniker"/moniker = "node-'"$num"'"/' $HOME/.rd/config/config.toml
