package randapp

import "github.com/dgamingfoundation/dkglib/lib/types"

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
