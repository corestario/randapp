# Messages

## MsgSendDKGData 

`MsgSendDKGData` message is the only message used by the module. It holds the seed and sender address. 

| **Field** | **Type**                                                 | **Description**                                              |
|:----------|:---------------------------------------------------------|:-------------------------------------------------------------|
| Owner     | `sdk.AccAddress`                                         | The account address of the user sending the message.         |
| Data      | `github.com/corestario/dkglib/lib/alias.DKGData`         | DKG data itself (see below)                                  |

``` go
type MsgSendDKGData struct {
	Data  *alias.DKGData `json:"data"`
	Owner sdk.AccAddress `json:"owner"`
}

type DKGData struct {
	Type        DKGDataType
	Addr        []byte
	RoundID     int
	Data        []byte // Data is going to keep serialized kyber objects.
	ToIndex     int    // ID of the participant for whom the message is; might be not set
	NumEntities int    // Number of sub-entities in the Data array, sometimes required for unmarshaling.
	Signature   []byte //Signature for verifying data
}
```