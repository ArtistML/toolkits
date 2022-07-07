package grpc

import (
	"math"

	"github.com/dustin/go-humanize"
	"github.com/gogf/gf/frame/g"
)

// GetReceiveAndSendMsgSize 获取消息大小的配置信息
func GetReceiveAndSendMsgSize() (int, int) {
	receiveStr := g.Cfg().GetString("grpc.message.receive")
	sendStr := g.Cfg().GetString("grpc.message.send")
	g.Log("").Infof("receiveStr = %s, sendStr = %s", receiveStr, sendStr)
	var receiveSize = 1024 * 1024 * 4
	var sendSize = math.MaxInt32
	if receiveStr != "" {
		parseBytes, err := humanize.ParseBytes(receiveStr)
		if err == nil {
			receiveSize = int(parseBytes)
		}
	}
	if sendStr != "" {
		parseBytes, err := humanize.ParseBytes(sendStr)
		if err == nil {
			sendSize = int(parseBytes)
		}
	}
	return receiveSize, sendSize
}
