#!/bin/bash

echo "Generating $VALIDATORS_COUNT containers"

re='^[0-9]+$'
if ! [[ $VALIDATORS_COUNT =~ $re ]] ; then
   echo "usage: set VALIDATORS_COUNT to desired number of validator nodes" >&2; exit 1
fi


echo "version: '3'">docker-compose.yml
echo "services:">>docker-compose.yml


for (( i=0; i < $VALIDATORS_COUNT; i++ ))
do
    ports_start=$((26656+2*i))
    ports_end=$(($ports_start+1))
    echo "">> docker-compose.yml
    echo "  rdnode$i:">> docker-compose.yml
    echo "    container_name: rdnode$i">> docker-compose.yml
    echo "    image: \"tendermint/rdnode\"">> docker-compose.yml
    echo "    environment:">> docker-compose.yml
    echo "      - ID=$i">> docker-compose.yml
    echo "      - LOG=\${LOG:-rd.log}">> docker-compose.yml
    echo "      - HOME=/rd/node$i/rd">> docker-compose.yml
    echo "    ports:">> docker-compose.yml
    echo "      - \"$ports_start-$ports_end:26656-26657\"">> docker-compose.yml
    echo "    volumes:">> docker-compose.yml
    echo "      - ./build:/rd:Z">> docker-compose.yml
    echo "    networks:">> docker-compose.yml
    echo "      localnet:">> docker-compose.yml
    echo "        ipv4_address: 192.168.10.$((i+2))">> docker-compose.yml
done

cat docker-compose.yml.template >> docker-compose.yml