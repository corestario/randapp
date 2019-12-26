package randapp

import (
	types "github.com/corestario/dkglib/lib/alias"
)

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
