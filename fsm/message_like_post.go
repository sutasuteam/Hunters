package fsm

import "github.com/canopy-network/canopy/lib"

type MessageLikePost struct {

    Liker lib.HexBytes `json:"liker"`

    Owner lib.HexBytes `json:"owner"`

    PostID uint64 `json:"postId"`

    Liked bool `json:"liked"`
}