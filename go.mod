module github.com/corestario/randapp

go 1.12

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/corestario/dkglib v1.0.0
	github.com/cosmos/cosmos-sdk v0.34.4-0.20191013030331-92ea174ea6e6
	github.com/cosmos/modules/incubator/reseeding v1.0.0
	github.com/gorilla/mux v1.7.3
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/common v0.7.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.6.1
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.32.8
	github.com/tendermint/tm-db v0.3.0
	github.com/corestario/cosmos-utils/client v0.0.0-20200319080016-b6dea902750d
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

replace github.com/tendermint/tendermint => ./../tendermint

replace github.com/corestario/dkglib => ./../dkglib

replace github.com/cosmos/cosmos-sdk => ./../cosmos-sdk

replace github.com/corestario/randapp => ./

replace go.dedis.ch/kyber/v3 => ./../kyber

replace github.com/cosmos/modules/incubator/reseeding => ./../modules/incubator/reseeding
