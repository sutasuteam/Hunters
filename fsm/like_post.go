package fsm

import "github.com/canopy-network/canopy/lib"

func (f *FSM) ExecuteLikePost(

    msg *MessageLikePost,

) lib.ErrorI {

    post, err := f.GetPost(

        msg.Owner,

        msg.PostID,

    )

    if err != nil {

        return err
    }

    if msg.Liked {

        post.Likes++

    } else {

        if post.Likes > 0 {

            post.Likes--
        }
    }

    return f.SetPost(post)
}