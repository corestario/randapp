#!/usr/bin/env sh

##
## Input parameters
##
BINARY=/rd/${BINARY:-rd}
ID=${ID:-0}
LOG=${LOG:-rd.log}

yes '12345678' | /rd/rd init --chain-id rchain validator
yes '12345678' | /rd/rcli keys add validator
valName=$(yes '12345678' | /rd/rcli keys show validator -a)
echo $valName
yes '12345678' | /rd/rd add-genesis-account $valName 1000nametoken,100000000stake

/rd/rcli config chain-id rchain
/rd/rcli config output json
/rd/rcli config indent true
/rd/rcli config trust-node true

##
## Assert linux binary
##
if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'rd' E.g.: -e BINARY=rd_my_test_version"
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
export RDHOME="/rd/node${ID}/rd"

if [ -d "$(dirname "${RDHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${RDHOME}" "$@" | tee "${RDHOME}/${LOG}"
else
  "${BINARY}" --home "${RDHOME}" "$@"
fi



``