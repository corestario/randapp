#!/bin/bash

# begin

reseed_interval=$1
#use for reseeding module access
sender_node_id=$2

echo "reseed every $reseed_interval blocks"

function get_last_block_height() {
  local hs=$(rcli q block --trust-node | jq '.block_meta.header.height | tonumber')
  echo $hs
}


function get_btc_hash() {
  local hsh=$(curl https://blockchain.info/latestblock | jq .hash)
  echo $hsh
  if [ -z $hsh ]
  then
    return 1
  else
    return 0
  fi
}

prevHeight=$(get_last_block_height)
echo last block height = $prevHeight


function is_time_to_reseed() {
  local lb=$(get_last_block_height)
  local pH=$prevHeight
  local diff=$(($lb - $pH))
  echo $lb
  if [ $diff -ge $reseed_interval ]
  then
    return 0
  else
    return 1
  fi
}


while [ 1 ]
do
  r=$(is_time_to_reseed)
  r2=$?
  if [ $r2 -eq 0 ]
  then
    echo "time to reseed"
    prevHeight=$r
    seed=$(get_btc_hash)
    s2=$?
    if [ $s2 -eq 0 ]
    then
      docker exec -ti "rdnode$sender_node_id" bash -c "./rcli tx reseeding send $seed --keyring-backend=test --chain-id=rchain --from node$sender_node_id -y"
    fi
  fi
  echo "Waiting for reseeding time..."
  sleep 2
done

# end
