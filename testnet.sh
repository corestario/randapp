#!/usr/bin/env bash

cur_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )

while test $# -gt 0; do
  case "$1" in
    -h|--help)
      echo "testnet - run randapp testnet"
      echo " "
      echo "testnet [options]"
      echo " "
      echo "options:"
      echo "-h, --help                    show brief help"
      echo "-n, --maximum_nodes=n         specify maximum node count"
      echo "-c, --node_count=c            specify node count"
      echo "-p, --bots_per_node=p         specify bots per node count"
      echo "--no_rebuild                  run without rebuilding docker images"
      echo "--kill                        stop and remove testnet containers"
      echo "--ruin                        force stop containers 1 and 2 after 5 seconds running dkg"
      exit 0
      ;;
    -p|--bots_per_node)
      shift
      if test $# -gt 0; then
        export bots_per_node=$1
      else
        echo "no bots_per_node specified"
        exit 1
      fi
      shift
      ;;
    -n|--maximum_nodes)
      shift
      if test $# -gt 0; then
        export n=$1
      else
        echo "no maximum_nodes specified"
        exit 1
      fi
      shift
      ;;
    -c|--node_count)
      shift
      if test $# -gt 0; then
        export node_count=$1
      else
        echo "no node_count specified"
        exit 1
      fi
      shift
      ;;
    --no_rebuild)
      NOREBUILD=true
      shift
      ;;
    --kill)
      nodeArray=$(cat nodeArray.txt)
      docker stop ${nodeArray[@]}
      docker rm ${nodeArray[@]}
      rm -rf $cur_path/node0_config
      rm $cur_path/nodeArray.txt
      exit 0
      shift
      ;;
    --ruin)
      FORCERUIN=true
      shift
      ;;
    *)
      break
      ;;
  esac
done

if [[ -n $bots_per_node ]] && [[ -n $node_count ]]
then
      n=$(( $bots_per_node*$node_count ))
fi

if [[ -z $n ]]
then
      n=4
fi

if [[ -z $bots_per_node ]]
then
      bots_per_node=1
      node_count=$n
fi


if [[ -z $node_count ]]
then
      node_count=$(($n/$bots_per_node))
fi

if [[ -z $t2 ]]
then
      t2=$(((n*2)/3))
fi

if [[ -z $t1 ]]
then
      t1=$t2
fi

echo "params: $t1 $t2 $n"
echo "node_count: $node_count"
echo "bots_per_node: $bots_per_node"

sleep 3

cd $cur_path
rm -rf ./vendor

rm -rf ./node0_config
mkdir ./node0_config

gopath=$(whereis go | grep -oP '(?<=go: )(\S*)(?= .*)' -m 1)
PATH=$gopath:$gopath/bin:$PATH

echo $GOBIN

if [[ $NOREBUILD ]]
then
  echo "no rebuild"
else
  make prepare
  GO111MODULE=off

  cd $cur_path/../dkglib
  ./testnet.sh
  cd $cur_path
  docker build -t randapp_testnet .
fi

RAPATH=/go/src/github.com/dgamingfoundation/randapp

echo "run node0"

node0_full_id=$(docker run -d randapp_testnet /bin/bash -c "$RAPATH/scripts/init_chain_full.sh $n;
 sed -i 's/timeout_commit = \"5s\"/timeout_commit = \"1s\"/' /root/.rd/config/config.toml;
 rd start > /root/rd_start.log")
node0_id=${node0_full_id:0:12}

echo "node0: $node0_id"
echo

while  ! docker exec $node0_id /bin/bash -c "[[ -d /root/.rd ]]" ; do
sleep 2
echo "waiting ..."
done

sleep 15

docker cp $node0_id:/root/.rd ./node0_config/.rd
docker cp $node0_id:/root/.rcli ./node0_config/.rcli

chmod -R 0777 ./node0_config

node0_addr=$(cat ./node0_config/.rd/config/genesis.json | jq '.app_state.genutil.gentxs[0].value.memo')

echo node0_addr
echo $node0_addr

if [[ -z $node0_addr ]] || [[ $node0_addr == "null" ]] || [[ $node0_addr == null ]]
then
  echo "ERROR"
  exit 1
fi

sed -i "s/seeds = \"\"/seeds = $node0_addr/" ./node0_config/.rd/config/config.toml
sed -i "s/persistent_peers = \"\"/persistent_peers = $node0_addr/" ./node0_config/.rd/config/config.toml

nodeArray=($node0_id)

for ((i=1;i<$node_count;i++));
do
    nodeN_full_id=$(docker create -t randapp_testnet /bin/bash -c "$RAPATH/scripts/init_chain.sh $i > /root/inch.log && rd start > /root/rd_start.log")
    nodeN_id=${nodeN_full_id:0:12}

    nodeArray+=($nodeN_id)

    docker cp ./node0_config/.rd/config/config.toml $nodeN_id:/root/tmp/
    docker cp ./node0_config/.rd/config/genesis.json $nodeN_id:/root/tmp/
    docker cp ./node0_config/.rcli $nodeN_id:/root/tmp/.rcli

    docker start $nodeN_id

    echo "node_num: $i, node_id: $nodeN_id"

done

sleep 10

echo "${nodeArray[@]}" > nodeArray.txt

chmod 0777 ./nodeArray.txt

echo "${nodeArray[@]}"
echo "all nodes started"
echo "run run_clients"

sleep 10

for ((i=0;i<$node_count;i++));
do
  nodeN_id=${nodeArray[$i]}
  docker exec -d $nodeN_id /bin/bash -c "dkglib -num=$i > /root/dkglib.log" &
  echo "node_num: $i, node_id: $nodeN_id"
done

if [[ $FORCERUIN ]]
then
  sleep 5
  docker stop ${nodeArray[1]}
  docker stop ${nodeArray[2]}
fi