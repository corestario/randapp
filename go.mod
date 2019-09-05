module github.com/dgamingfoundation/randapp

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.28.2-0.20190827131926-5aacf454e1b6
	github.com/dgamingfoundation/marketplace v0.0.0-20190829103843-b5b8e1c556ac
	github.com/gorilla/mux v1.7.0
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/common v0.4.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.32.2
	github.com/tendermint/tm-db v0.1.1
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

replace github.com/tendermint/tendermint => github.com/dgamingfoundation/tendermint v0.27.4-0.20190902075848-ce9fbae9b910
