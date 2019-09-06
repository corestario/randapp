#!/usr/bin/env bash

randapp_path=/go/src/github.com/dgamingfoundation/randapp

rm -rf $HOME/.rcli
rm -rf $HOME/.rd

rd init moniker --chain-id randappchain

rcli config chain-id randappchain
rcli config output json
rcli config indent true
rcli config trust-node true

mkdir -p $HOME/.rd/config

cp $HOME/tmp/genesis.json $HOME/.rd/config
cp $HOME/tmp/config.toml $HOME/.rd/config
cp $HOME/tmp/bls_key.json $HOME/.rd/config
cp -r $HOME/tmp/keys $HOME/.rcli/

sed -i 's/moniker = "moniker"/moniker = "node-'"$1"'"/' $HOME/.rd/config/config.toml
