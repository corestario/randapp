#!/usr/bin/env bash

randapp_path=/go/src/github.com/corestario/randapp
num=$1
rm -rf /root/.rcli
rm -rf /root/.rd

rd init moniker --chain-id rchain

rcli config chain-id rchain
rcli config output json
rcli config indent true
rcli config trust-node true

mkdir -p $HOME/.rd/config
rm -rf /root/.rcli
cp /root/tmp/genesis.json /root/.rd/config
cp /root/tmp/config.toml /root/.rd/config
cp -r /root/tmp/.rcli /root/.rcli
cp -r /root/tmp/.rcli /root/.rcli${num}

sed -i 's/moniker = "moniker"/moniker = "node-'"$num"'"/' /root/.rd/config/config.toml
