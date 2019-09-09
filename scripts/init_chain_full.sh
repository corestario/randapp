#!/usr/bin/env bash

n=$1

dir_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )
randapp_path=$dir_path/..

pwrd="12345678"

rm -rf $HOME/.rcli
rm -rf $HOME/.rd

rd init moniker --chain-id rchain

for (( i=0; i<$n; i++ ))
do
    rcli keys add "validator$i" <<< $pwrd

    rd add-genesis-account $(rcli keys show "validator$i" -a) 1000nametoken,100000000stake
done

rcli config chain-id rchain
rcli config output json
rcli config indent true
rcli config trust-node true

cp -r /root/.rcli /root/.rcli0

rd gentx --name validator0 <<< $pwrd

rd collect-gentxs

rd validate-genesis

