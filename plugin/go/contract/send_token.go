package contract

import "log"

func (c *Contract) DeliverMessageSendToken(
    msg *MessageSendToken,
) *PluginDeliverResponse {

    log.Printf(
        "SendToken from=%x to=%x token=%s amount=%d",
        msg.From,
        msg.To,
        msg.TokenId,
        msg.Amount,
    )

    // sementara hanya loloskan transaksi
    return &PluginDeliverResponse{}
}