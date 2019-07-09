#!/usr/bin/env bash

##
## Input parameters
##
BINARY=/rd/${BINARY:-rd}
UNARY=/rd/${UNARY:-rcli}
ID=${ID:-0}
LOG=${LOG:-rd.log}

##
## Assert linux binary
##
if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'rd' E.g.: -e BINARY=rd"
	exit 1
fi
BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"
if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

##
## Run binary with all parameters
##
export RDHOME="/rd/node${ID}"


#ls /root -a
#ls /root/.rd -a
#ls /root/.rd/config -a
#cat /root/.rd/config/genesis.json

if [ -d "`dirname /root/.rd/config`" ]; then
	#echo 'PHASE0'
	#$BINARY init --chain-id rchain validator${ID}
	#$UNARY keys add validator${ID} <<< "12345678"
	#$BINARY add-genesis-account $($UNARY keys show validator${ID} -a) 1000nametoken,100000000stake

	$UNARY config chain-id rchain
	$UNARY config output json
    $UNARY config indent true
    $UNARY config trust-node true

	#echo 'PHASE1!'
	#$BINARY gentx --name validator${ID} <<< '12345678'
	#echo 'PHASE2!'
	#$BINARY collect-gentxs
	#echo 'PHASE3!'
	#$BINARY validate-genesis
	ls /root/.rd/config/ -a
	cat /root/.rd/config/genesis.json
fi

if [ -d "`dirname ${RDHOME}/${LOG}`" ]; then
  "$BINARY" "$@" | tee "${RDHOME}/${LOG}"
else
  "$BINARY" "$@"
fi

echo validator${ID}
echo validator${ID}
echo validator${ID}
echo validator${ID}

chmod 777 -R /rd


#$BINARY collect-gentxs
#$BINARY validate-genesis