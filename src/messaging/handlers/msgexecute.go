package handlers

import (
	"context"
	"encoding/json"

	"github.com/blocklessnetworking/b7s/src/enums"
	"github.com/blocklessnetworking/b7s/src/models"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

func HandleMsgExecute(ctx context.Context, message []byte) {
	msgExecute := &models.MsgExecute{}
	json.Unmarshal(message, msgExecute)
	msgExecute.From = ctx.Value("peerID").(peer.ID)
	log.WithFields(log.Fields{
		"message": string(message),
	}).Info("message from peer")

	channel := ctx.Value(enums.ChannelMsgExecute).(chan models.MsgExecute)
	channel <- *msgExecute
}
