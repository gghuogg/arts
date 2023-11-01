package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetPR(consts.PRTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_DOWNLOAD_FILE, &tgDownloadFile{})
}

type tgDownloadFile struct{}

func (t *tgDownloadFile) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetGetDownLoadFileDetail().GetDownloadFile()
	tgDownloadFile := p.PrGetSyncMap(rpc.TG_GET_DOWNLOAD_FILE)
	for k, v := range details {
		value, ok := p.Handler.GetServerMap().Load(k)
		if ok {
			ip := value.(string)
			existingData, _ := tgDownloadFile.Load(ip)
			if existingData != nil {
				tmp := existingData.(map[uint64]*protobuf.DownLoadFileMsg)
				tmp[k] = v
				tgDownloadFile.Store(ip, tmp)
			} else {
				tmp := make(map[uint64]*protobuf.DownLoadFileMsg)
				tmp[k] = v
				tgDownloadFile.Store(ip, tmp)
			}
		}
	}
	return
}
