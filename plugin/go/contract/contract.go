package contract

import (
	"bytes"
	"encoding/binary"
	"log"
	"math/rand"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

var ContractConfig = &PluginConfig{
	Name:    "go_plugin_contract",
	Id:      1,
	Version: 1,
	SupportedTransactions: []string{
		"send",
		"mint_nft",
		"set_stake",
		"send_token",
		"like_post",
	},

TransactionTypeUrls: []string{
    "type.googleapis.com/types.MessageSend",
    "type.googleapis.com/types.MessageMintNFT",
    "type.googleapis.com/types.MessageSetStake",
    "type.googleapis.com/types.MessageSendToken",
	"type.googleapis.com/types.MessageLikePost",
},
	EventTypeUrls: nil,
}

func init() {
	file_account_proto_init()
	file_event_proto_init()
	file_plugin_proto_init()
	file_tx_proto_init()

	var fds [][]byte

	for _, file := range []protoreflect.FileDescriptor{
		anypb.File_google_protobuf_any_proto,
		File_account_proto,
		File_event_proto,
		File_plugin_proto,
		File_tx_proto,
	} {
		fd, _ := proto.Marshal(protodesc.ToFileDescriptorProto(file))
		fds = append(fds, fd)
	}

	ContractConfig.FileDescriptorProtos = fds
}

type Contract struct {
	Config        Config
	FSMConfig     *PluginFSMConfig
	plugin        *Plugin
	fsmId         uint64
	currentHeight uint64
}

func (c *Contract) Genesis(_ *PluginGenesisRequest) *PluginGenesisResponse {
	return &PluginGenesisResponse{}
}

func (c *Contract) BeginBlock(req *PluginBeginRequest) *PluginBeginResponse {
	c.currentHeight = req.Height
	return &PluginBeginResponse{}
}

func (c *Contract) CheckTx(request *PluginCheckRequest) *PluginCheckResponse {
	resp, err := c.plugin.StateRead(c, &PluginStateReadRequest{
		Keys: []*PluginKeyRead{
			{QueryId: rand.Uint64(), Key: KeyForFeeParams()},
		},
	})

	if err == nil {
		err = resp.Error
	}

	if err != nil {
		return &PluginCheckResponse{Error: err}
	}

	minFees := new(FeeParams)
	if err = Unmarshal(resp.Results[0].Entries[0].Value, minFees); err != nil {
		return &PluginCheckResponse{Error: err}
	}

	if request.Tx.Fee < minFees.SendFee {
		return &PluginCheckResponse{Error: ErrTxFeeBelowStateLimit()}
	}

	msg, err := FromAny(request.Tx.Msg)
	if err != nil {
		return &PluginCheckResponse{Error: err}
	}

	switch x := msg.(type) {
	case *MessageSend:
		return c.CheckMessageSend(x)
	case *MessageMintNFT:
		return c.CheckMessageMintNFT(x)
	case *MessageSetStake:
		return c.CheckMessageSetStake(x)
	case *MessageSendToken:

		return c.CheckMessageSendToken(x)
	case *MessageLikePost:

		return c.CheckMessageLikePost(x)
	default:
		return &PluginCheckResponse{Error: ErrInvalidAmount()}
	}
}

func (c *Contract) DeliverTx(request *PluginDeliverRequest) *PluginDeliverResponse {
	msg, err := FromAny(request.Tx.Msg)
	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	switch x := msg.(type) {
	case *MessageSend:
		return c.DeliverMessageSend(x, request.Tx.Fee)
	case *MessageMintNFT:
		return c.DeliverMessageMintNFT(x, request.Tx.Fee)
	case *MessageSetStake:
		return c.DeliverMessageSetStake(x, request.Tx.Fee)
	case *MessageSendToken:

		return c.DeliverMessageSendToken(x)
	case *MessageLikePost:

		return c.DeliverMessageLikePost(x)
	default:
		return &PluginDeliverResponse{Error: ErrInvalidAmount()}
	}
}

func (c *Contract) EndBlock(_ *PluginEndRequest) *PluginEndResponse {
	return &PluginEndResponse{}
}

func (c *Contract) CheckMessageSend(msg *MessageSend) *PluginCheckResponse {
	if len(msg.FromAddress) != 20 {
		return &PluginCheckResponse{Error: ErrInvalidAddress()}
	}

	if len(msg.ToAddress) != 20 {
		return &PluginCheckResponse{Error: ErrInvalidAddress()}
	}

	if msg.Amount == 0 {
		return &PluginCheckResponse{Error: ErrInvalidAmount()}
	}

	return &PluginCheckResponse{
		Recipient:         msg.ToAddress,
		AuthorizedSigners: [][]byte{msg.FromAddress},
	}
}

func (c *Contract) CheckMessageMintNFT(msg *MessageMintNFT) *PluginCheckResponse {
	if len(msg.Creator) != 20 {
		return &PluginCheckResponse{Error: ErrInvalidAddress()}
	}

	if msg.TokenId == "" {
		return &PluginCheckResponse{Error: ErrInvalidAmount()}
	}

	if msg.Name == "" {
		return &PluginCheckResponse{Error: ErrInvalidAmount()}
	}

	if msg.Metadata == "" {
		return &PluginCheckResponse{Error: ErrInvalidAmount()}
	}

	return &PluginCheckResponse{
		AuthorizedSigners: [][]byte{msg.Creator},
	}
}


func (c *Contract) CheckMessageSetStake(msg *MessageSetStake) *PluginCheckResponse {

    if len(msg.Owner) != 20 {
        return &PluginCheckResponse{
            Error: ErrInvalidAddress(),
        }
    }

    if msg.TokenId == "" {
        return &PluginCheckResponse{
            Error: ErrInvalidAmount(),
        }
    }

    return &PluginCheckResponse{
        AuthorizedSigners: [][]byte{
            msg.Owner,
        },
    }

}


func (c *Contract) CheckMessageSendToken(
    msg *MessageSendToken,
) *PluginCheckResponse {

    if len(msg.From) != 20 {
        return &PluginCheckResponse{
            Error: ErrInvalidAddress(),
        }
    }

    if len(msg.To) != 20 {
        return &PluginCheckResponse{
            Error: ErrInvalidAddress(),
        }
    }

    if msg.TokenId == "" {
        return &PluginCheckResponse{
            Error: ErrInvalidAmount(),
        }
    }

    if msg.Amount == 0 {
        return &PluginCheckResponse{
            Error: ErrInvalidAmount(),
        }
    }

    return &PluginCheckResponse{
        Recipient: msg.To,
        AuthorizedSigners: [][]byte{
            msg.From,
        },
    }
}

func (c *Contract) CheckMessageLikePost(
    msg *MessageLikePost,
) *PluginCheckResponse {

    if len(msg.Liker) != 20 {

        return &PluginCheckResponse{
            Error: ErrInvalidAddress(),
        }

    }

    if len(msg.Owner) != 20 {

        return &PluginCheckResponse{
            Error: ErrInvalidAddress(),
        }

    }

    if msg.PostId == 0 {

        return &PluginCheckResponse{
            Error: ErrInvalidAmount(),
        }

    }

    return &PluginCheckResponse{

        Recipient: msg.Owner,

        AuthorizedSigners: [][]byte{
            msg.Liker,
        },

    }

}

func (c *Contract) DeliverMessageSend(msg *MessageSend, fee uint64) *PluginDeliverResponse {
	log.Printf("DeliverMessageSend called: from=%x to=%x amount=%d fee=%d", msg.FromAddress, msg.ToAddress, msg.Amount, fee)

	var (
		fromKey, toKey, feePoolKey         []byte
		fromBytes, toBytes, feePoolBytes   []byte
		fromQueryId, toQueryId, feeQueryId = rand.Uint64(), rand.Uint64(), rand.Uint64()
		from, to, feePool                  = new(Account), new(Account), new(Pool)
	)

	fromKey = KeyForAccount(msg.FromAddress)
	toKey = KeyForAccount(msg.ToAddress)
	feePoolKey = KeyForFeePool(c.Config.ChainId)

	response, err := c.plugin.StateRead(c, &PluginStateReadRequest{
		Keys: []*PluginKeyRead{
			{QueryId: feeQueryId, Key: feePoolKey},
			{QueryId: fromQueryId, Key: fromKey},
			{QueryId: toQueryId, Key: toKey},
		},
	})

	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	if response.Error != nil {
		return &PluginDeliverResponse{Error: response.Error}
	}

	for _, resp := range response.Results {
		if len(resp.Entries) == 0 {
			continue
		}

		switch resp.QueryId {
		case fromQueryId:
			fromBytes = resp.Entries[0].Value
		case toQueryId:
			toBytes = resp.Entries[0].Value
		case feeQueryId:
			feePoolBytes = resp.Entries[0].Value
		}
	}

	amountToDeduct := msg.Amount + fee

	if err = Unmarshal(fromBytes, from); err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	if err = Unmarshal(toBytes, to); err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	if err = Unmarshal(feePoolBytes, feePool); err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	if from.Amount < amountToDeduct {
		return &PluginDeliverResponse{Error: ErrInsufficientFunds()}
	}

	if bytes.Equal(fromKey, toKey) {
		to = from
	}

	from.Amount -= amountToDeduct
	feePool.Amount += fee
	to.Amount += msg.Amount

	fromBytes, err = Marshal(from)
	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	toBytes, err = Marshal(to)
	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	feePoolBytes, err = Marshal(feePool)
	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	var writeResp *PluginStateWriteResponse

	if from.Amount == 0 {
		writeResp, err = c.plugin.StateWrite(c, &PluginStateWriteRequest{
			Sets: []*PluginSetOp{
				{Key: feePoolKey, Value: feePoolBytes},
				{Key: toKey, Value: toBytes},
			},
			Deletes: []*PluginDeleteOp{
				{Key: fromKey},
			},
		})
	} else {
		writeResp, err = c.plugin.StateWrite(c, &PluginStateWriteRequest{
			Sets: []*PluginSetOp{
				{Key: feePoolKey, Value: feePoolBytes},
				{Key: toKey, Value: toBytes},
				{Key: fromKey, Value: fromBytes},
			},
		})
	}

	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	if writeResp.Error != nil {
		return &PluginDeliverResponse{Error: writeResp.Error}
	}

	return &PluginDeliverResponse{}
}

func (c *Contract) DeliverMessageMintNFT(msg *MessageMintNFT, fee uint64) *PluginDeliverResponse {
	log.Printf("DeliverMessageMintNFT called: creator=%x tokenId=%s name=%s fee=%d", msg.Creator, msg.TokenId, msg.Name, fee)

	var (
		nftKey, creatorKey, feePoolKey     []byte
		creatorBytes, feePoolBytes         []byte
		nftQueryId, creatorQueryId, feeQId = rand.Uint64(), rand.Uint64(), rand.Uint64()
		creator, feePool                   = new(Account), new(Pool)
	)

	nftKey = KeyForNFT(msg.TokenId)
	creatorKey = KeyForAccount(msg.Creator)
	feePoolKey = KeyForFeePool(c.Config.ChainId)

	response, err := c.plugin.StateRead(c, &PluginStateReadRequest{
		Keys: []*PluginKeyRead{
			{QueryId: nftQueryId, Key: nftKey},
			{QueryId: creatorQueryId, Key: creatorKey},
			{QueryId: feeQId, Key: feePoolKey},
		},
	})

	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	if response.Error != nil {
		return &PluginDeliverResponse{Error: response.Error}
	}

	for _, resp := range response.Results {
		switch resp.QueryId {
		case nftQueryId:
			if len(resp.Entries) > 0 {
				log.Printf("NFT already exists, overwriting for dev test: tokenId=%s", msg.TokenId)
			}
		case creatorQueryId:
			if len(resp.Entries) > 0 {
				creatorBytes = resp.Entries[0].Value
			}
		case feeQId:
			if len(resp.Entries) > 0 {
				feePoolBytes = resp.Entries[0].Value
			}
		}
	}

	if err = Unmarshal(creatorBytes, creator); err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	if err = Unmarshal(feePoolBytes, feePool); err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	if creator.Amount < fee {
		return &PluginDeliverResponse{Error: ErrInsufficientFunds()}
	}

	creator.Amount -= fee
	feePool.Amount += fee

	nft := &NFT{
		TokenId:       msg.TokenId,
		Creator:       msg.Creator,
		Name:          msg.Name,
		Image:         msg.Image,
		Metadata:      msg.Metadata,
		CreatedHeight: c.currentHeight,
		Staked: false,
	}

	nftBytes, err := Marshal(nft)
	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	creatorBytes, err = Marshal(creator)
	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	feePoolBytes, err = Marshal(feePool)
	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	writeResp, err := c.plugin.StateWrite(c, &PluginStateWriteRequest{
		Sets: []*PluginSetOp{
			{Key: nftKey, Value: nftBytes},
			{Key: creatorKey, Value: creatorBytes},
			{Key: feePoolKey, Value: feePoolBytes},
		},
	})

	if err != nil {
		return &PluginDeliverResponse{Error: err}
	}

	if writeResp.Error != nil {
		return &PluginDeliverResponse{Error: writeResp.Error}
	}

	log.Printf("NFT minted SUCCESS: tokenId=%s", msg.TokenId)

	return &PluginDeliverResponse{}
}


func (c *Contract) DeliverMessageSetStake(
    msg *MessageSetStake,
    fee uint64,
) *PluginDeliverResponse {

    nftKey := KeyForNFT(msg.TokenId)

    response, err := c.plugin.StateRead(c,
        &PluginStateReadRequest{
            Keys: []*PluginKeyRead{
                {
                    QueryId: 1,
                    Key: nftKey,
                },
            },
        })

    if err != nil {
        return &PluginDeliverResponse{
            Error: err,
        }
    }

    if len(response.Results[0].Entries) == 0 {
        return &PluginDeliverResponse{
            Error: ErrInvalidAmount(),
        }
    }

    nft := new(NFT)

    err = Unmarshal(
        response.Results[0].Entries[0].Value,
        nft,
    )

    if err != nil {
        return &PluginDeliverResponse{
            Error: err,
        }
    }

    if !bytes.Equal(nft.Creator, msg.Owner) {
        return &PluginDeliverResponse{
            Error: ErrUnauthorized(),
        }
    }

    nft.Staked = msg.Stake

    nftBytes, _ := Marshal(nft)

    _, err = c.plugin.StateWrite(c,
        &PluginStateWriteRequest{
            Sets: []*PluginSetOp{
                {
                    Key: nftKey,
                    Value: nftBytes,
                },
            },
        })

    if err != nil {
        return &PluginDeliverResponse{
            Error: err,
        }
    }

    return &PluginDeliverResponse{}
}


func (c *Contract) DeliverMessageLikePost(
    msg *MessageLikePost,
) *PluginDeliverResponse {

    log.Printf(
        "LIKE POST | liker=%x owner=%x post=%d liked=%v",
        msg.Liker,
        msg.Owner,
        msg.PostId,
        msg.Liked,
    )

    return &PluginDeliverResponse{}

}

var (
	accountPrefix = []byte{1}
	poolPrefix    = []byte{2}
	nftPrefix     = []byte{8}
	paramsPrefix  = []byte{7}
)

func KeyForAccount(addr []byte) []byte {
	return JoinLenPrefix(accountPrefix, addr)
}

func KeyForNFT(tokenId string) []byte {
	return JoinLenPrefix(nftPrefix, []byte(tokenId))
}

func KeyForFeeParams() []byte {
	return JoinLenPrefix(paramsPrefix, []byte("/f/"))
}

func KeyForFeePool(chainId uint64) []byte {
	return JoinLenPrefix(poolPrefix, formatUint64(chainId))
}

func formatUint64(u uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, u)
	return b
}
