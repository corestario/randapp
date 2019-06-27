module github.com/dgamingfoundation/randapp

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.28.2-0.20190616100639-18415eedaf25
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/spf13/cobra v0.0.4
	github.com/spf13/viper v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.6
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

replace github.com/tendermint/tendermint => github.com/dgamingfoundation/tendermint v0.27.4-0.20190604195457-d66632d1761e
