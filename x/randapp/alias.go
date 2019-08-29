package randapp

import "github.com/tendermint/tendermint/types"

type (
	DKGDataType = types.DKGDataType
)

const (
	DKGPubKey            = types.DKGPubKey
	DKGDeal              = types.DKGDeal
	DKGResponse          = types.DKGResponse
	DKGJustification     = types.DKGJustification
	DKGCommits           = types.DKGCommits
	DKGComplaint         = types.DKGComplaint
	DKGReconstructCommit = types.DKGReconstructCommit

	ModuleName = "randapp"
	RouterKey  = ModuleName
	StoreKey   = ModuleName
)
