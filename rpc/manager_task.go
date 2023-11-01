package rpc

import (
	"arthas/callback"
	"arthas/callback/consts"
	"arthas/etcd"
	"arthas/protobuf"
	"arthas/util"
	"context"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/encoding/gjson"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"strconv"
	"strings"
	"time"
)

func (h *Handler) watchProxy(env, ns string) {
	get := etcd.GetReq{
		Key:     "/service/" + env + "/" + ns + "/arthas/",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	h.getChan <- get
	result := <-get.ResChan

	if result.Result.Kvs != nil {
		tgList := make([]string, 0)
		for _, v := range result.Result.Kvs {
			serviceInfo := &protobuf.ServiceInfo{}
			proto.Unmarshal(v.Value, serviceInfo)
			level.Info(h.logger).Log("msg", "ETCD: Server is Online", "name", strings.Split(string(v.Key), "/")[5], "ip", serviceInfo.IP)
			if strings.Split(string(v.Key), "/")[5] == "telegram" {
				tgList = append(tgList, serviceInfo.IP)
			}
			stream := h.connect(serviceInfo.IP)
			// 初始化登录账号到内存里
			appName := strings.Split(string(v.Key), "/")[5]
			h.initTgLoginAccountToMap(string(v.Key), appName)
			h.heart(stream)

		}

		if len(tgList) > 0 {
			h.restartOnline("telegram", tgList)
		}
	}

	go func() {
		for {
			select {
			case <-h.ExitChan:
				return
			default:

				reqModel := etcd.WatchReq{
					Key:     "/service/" + env + "/" + ns + "/arthas/",
					Options: []clientv3.OpOption{clientv3.WithPrefix()},
					ResChan: make(chan etcd.WatchRes),
				}
				h.watchChan <- reqModel

				for res := range reqModel.ResChan {
					for _, ev := range res.Events {
						switch ev.Type {
						case mvccpb.DELETE:
							level.Info(h.logger).Log("msg", "ETCD: Server is Offline", "name", strings.Split(string(ev.Kv.Key), "/")[5], "ip", strings.Split(string(ev.Kv.Key), "/")[7])
						case mvccpb.PUT:
							serviceInfo := &protobuf.ServiceInfo{}
							proto.Unmarshal(ev.Kv.Value, serviceInfo)
							level.Info(h.logger).Log("msg", "ETCD: Server is Online", "name", strings.Split(string(ev.Kv.Key), "/")[5], "ip", serviceInfo.IP)

							streamConnectClient := h.connect(serviceInfo.IP)
							h.heart(streamConnectClient)

							appName := strings.Split(string(ev.Kv.Key), "/")[5]
							list, getListErr := h.GetServiceListByName(appName)
							if getListErr != nil {
								level.Error(h.logger).Log("msg", "Failed to IP LISt", "err", getListErr)
							}
							h.restartOnline(appName, list)
						}
					}
				}

			}
		}
	}()

}

func (h *Handler) watchService(env, ns string) {
	get := etcd.GetReq{
		Key:     "/service/" + env + "/" + ns + "/arts_manager/",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	h.getChan <- get
	result := <-get.ResChan

	if result.Result.Kvs != nil {
		for _, v := range result.Result.Kvs {
			level.Info(h.logger).Log("msg", "ETCD: Management is Online", "ip", strings.Split(string(v.Key), "/")[6])
		}
	}

	go func() {
		for {
			select {
			case <-h.ExitChanArt:
				return
			default:
				reqModel := etcd.WatchReq{
					Key:     "/service/" + env + "/" + ns + "/arts_manager/",
					Options: []clientv3.OpOption{clientv3.WithPrefix()},
					ResChan: make(chan etcd.WatchRes),
				}
				h.watchChan <- reqModel

				for res := range reqModel.ResChan {
					for _, ev := range res.Events {
						switch ev.Type {
						case mvccpb.DELETE:
							level.Info(h.logger).Log("msg", "ETCD: Management is Offline", "ip", strings.Split(string(ev.Kv.Key), "/")[6])
						case mvccpb.PUT:
							level.Info(h.logger).Log("msg", "ETCD: Management is Online", "ip", strings.Split(string(ev.Kv.Key), "/")[6])
						}
					}
				}

			}
		}
	}()

}
func (h *Handler) RegisterServiceToEtcd(env, ns, version, servername string, serviceAddress string, management string, appName string) {
	endpoints := util.CalculateListenedEndpoints(serviceAddress)
	req := etcd.LeaseReq{ResChan: make(chan etcd.LeaseRes)}
	h.leaseChan <- req
	res := <-req.ResChan
	lease := res.Result

	leaseGrant, err := lease.Grant(h.context, 20)
	if err != nil {
		level.Error(h.logger).Log("msg", "leaseGrant err", "err", err)
	}

	if servername != "" {
		var key string
		if appName != "" {
			key = fmt.Sprintf("/service/%v/%v/%v/%v/%v/%v", env, ns, servername, appName, version, endpoints[0].String())
		}

		putValue := &protobuf.ServiceInfo{
			IP:          endpoints[0].String(),
			Connections: 0,
		}

		putvalueb, _ := proto.Marshal(putValue)

		putReq := etcd.PutReq{
			Key:     key,
			Value:   string(putvalueb),
			Options: []clientv3.OpOption{clientv3.WithLease(leaseGrant.ID)}, //并且绑定租约
			ResChan: make(chan etcd.PutRes),
		}
		h.putChan <- putReq
	}

	if management != "" {
		mKey := fmt.Sprintf("/service/%v/%v/%v/%v/%v", env, ns, management, version, endpoints[0].String())
		level.Info(h.logger).Log("msg", "management start registration", "management service", endpoints[0].String())
		putReq := etcd.PutReq{
			Key:     mKey,
			Value:   "",
			Options: []clientv3.OpOption{clientv3.WithLease(leaseGrant.ID)}, //并且绑定租约
			ResChan: make(chan etcd.PutRes),
		}
		h.putChan <- putReq
	}

	keepRespChan, err := lease.KeepAlive(h.context, leaseGrant.ID)
	if err != nil {
		level.Error(h.logger).Log("msg", "lease keepAlive err", "err", err)
	}

	go func() {
		for {
			select {
			case resp := <-keepRespChan:
				if resp == nil {
					level.Info(h.logger).Log("msg", "lease keepAlive fail")
					// 服务挂了回调
					artsServerDown := callback.ArtsServerDown{
						Message: "lease keepAlive fail...",
						Ip:      endpoints[0].String(),
					}
					h.callbackChan.ImCallbackChan <- callback.ImCallBackChan{
						Topic: consts.CallbackTopicArtsServerDown,
						Callback: callback.ImCallback{
							Type: consts.CallbackTypeTg,
							Data: gjson.MustEncode(artsServerDown),
						},
					}
					return
				}
			}
		}
	}()
}

func (h *Handler) record(env, ns, version, serviceAddress, appName string) {
	endpoints := util.CalculateListenedEndpoints(serviceAddress)
	key := fmt.Sprintf("/service/%v/%v/%v/%v/%v/%v", env, ns, "count", appName, version, endpoints[0].String())

	putValue := &protobuf.ServiceInfo{
		IP:          endpoints[0].String(),
		Connections: 0,
	}

	putvalueb, _ := proto.Marshal(putValue)

	putReq := etcd.PutReq{
		Key:     key,
		Value:   string(putvalueb),
		Options: nil,
		ResChan: make(chan etcd.PutRes),
	}
	h.putChan <- putReq
}

func (h *Handler) initTgLoginAccountToMap(key string, imType string) {

	splits := strings.Split(key, "/")
	ipPort := splits[len(splits)-1]
	get := etcd.GetReq{
		Key:     "/service/" + h.Env + "/" + h.NameSpace + "/online/" + ipPort + "/",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	h.getChan <- get
	result := <-get.ResChan
	for _, kv := range result.Result.Kvs {
		//fmt.Println(kv)
		ipToAccount := strings.Split(string(kv.Key), "/")
		accountPhone := ipToAccount[len(ipToAccount)-1]

		userJid, _ := strconv.ParseUint(accountPhone, 10, 64)
		h.serverMap.Store(userJid, ipPort)
	}

}

func (h *Handler) connect(ip string) protobuf.ArthasStream_ConnectClient {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancelServerStream = cancel

	conn, err := grpc.Dial(ip, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		level.Error(h.logger).Log("msg", "grpc dial err", "err", err)
	}

	client := protobuf.NewArthasStreamClient(conn)
	streamConnectClient, err := client.Connect(ctx)
	if err != nil {
		level.Error(h.logger).Log("msg", "client connect err", "err", err)
	}

	h.StreamMap.Store(ip, streamConnectClient)

	return streamConnectClient
}

func (h *Handler) heart(streamConnectClient protobuf.ArthasStream_ConnectClient) {

	if streamConnectClient != nil {
		go func() {
			for {
				recv, err := streamConnectClient.Recv()
				ip, err := getIPFromContext(streamConnectClient.Context())
				if err != nil {
					level.Error(h.logger).Log("msg", "Failed to resolve IP", "err", streamConnectClient.Context().Err())
				}
				if err != nil {
					//服务端断开，删除关联的流
					h.StreamMap.Delete(ip)
					level.Info(h.logger).Log("msg", "Server is Offline", "ip", ip)
					level.Error(h.logger).Log("msg", "streamConnectClient receive err", "err", err)
					return
				}

				if recv.GetPongMessage() != "" {
					heartRes := strings.Split(recv.PongMessage, "/")
					timeStamp := time.Now().Unix()
					put := etcd.PutReq{
						Key:     "/heart/" + heartRes[1] + "/" + heartRes[0],
						Value:   strconv.FormatInt(timeStamp, 10),
						Options: nil,
						ResChan: make(chan etcd.PutRes),
					}
					h.putChan <- put

					//level.Info(h.logger).Log("msg", "Server is Online", "ip", ip)
					//fmt.Println("收到：", ip)
				} else if recv != nil {
					h.finalResponse <- recv
				}

				//get := etcd.GetReq{
				//	Key:     "/heart/pong/" + ip,
				//	Options: nil,
				//	ResChan: make(chan etcd.GetRes),
				//}
				//h.getChan <- get
				//res := <-get.ResChan
				//if res.Result.Kvs != nil {
				//	oldStamp, _ := strconv.ParseInt(string(res.Result.Kvs[0].Value), 10, 64)
				//	nowStamp := time.Now().Unix()
				//	if checkTimeDifference(oldStamp, nowStamp) {
				//		fmt.Println("时间戳相差不超过31秒")
				//	}
				//}

			}

		}()

		go func() {
			for {
				select {
				case <-h.ticker.C:
					err := streamConnectClient.Send(&protobuf.RequestMessage{
						Action: protobuf.Action_PING_PONG,
						ActionDetail: &protobuf.RequestMessage_PingMessage{
							PingMessage: "ping",
						},
					})
					if err != nil {
						ip, err := getIPFromContext(streamConnectClient.Context())
						if err != nil {
							level.Error(h.logger).Log("msg", "Failed to resolve IP", "err", streamConnectClient.Context().Err())
						}

						level.Info(h.logger).Log("ip", ip, "msg", "The connection has been disconnected")
						level.Error(h.logger).Log("msg", "stream send err", "err", err)
						return
					} else {
						ip, _ := getIPFromContext(streamConnectClient.Context())
						h.StreamMap.Store(ip, streamConnectClient)
					}
				}
			}
		}()
	}

}

func (h *Handler) telegramWatcher() {

	//fmt.Println("监听开始。。。")

	lockKey := "/services/watch/" + h.Env + "/" + h.NameSpace + "/lock"
	se := etcd.SessionReq{
		ResChan: make(chan etcd.SessionRes),
	}
	h.sessionChan <- se

	session := <-se.ResChan
	//fmt.Println("session:", session.Result)
	defer session.Result.Close()

	lock := concurrency.NewMutex(session.Result, lockKey)

	for {

		err := lock.Lock(context.Background())
		//fmt.Println("err:", err)
		if err != nil {
			// 无法获取锁，可以选择重试或者处理其他逻辑
			level.Info(h.logger).Log("msg", "get etcd mutex fail,continue to watch")
			continue
		}

		reqModel := etcd.WatchReq{
			Key:     "/services/" + h.Env + "/" + h.NameSpace,
			Options: []clientv3.OpOption{clientv3.WithPrefix()},
			ResChan: make(chan etcd.WatchRes),
		}
		h.watchChan <- reqModel

		for res := range reqModel.ResChan {
			for _, ev := range res.Events {
				switch ev.Type {
				case mvccpb.DELETE:
					//fmt.Println(p.Handler.Env, p.Handler.NameSpace)
					userJid, _ := strconv.ParseUint(strings.Split(string(ev.Kv.Key), "/")[5], 10, 64)
					appName := strings.Split(string(ev.Kv.Key), "/")[4]
					value, ok := h.serverMap.Load(userJid)
					if ok {
						v1, ok1 := h.StreamMap.Load(value.(string))
						if !ok1 {
							level.Info(h.logger).Log("msg", "initiate failover")

							list, _ := h.GetServiceListByName(appName)
							if len(list) > 0 {
								least := h.DiscoverServicesFindLeast(h.NameSpace, h.Version, h.Env, appName)
								//在新的服务端重新登录
								if least != nil {
									h.serverMap.Store(userJid, least.IP)

									h.LoginAgain(least.IP, userJid, appName)
								}

							}

						} else {
							streamConnectClient1 := v1.(protobuf.ArthasStream_ConnectClient)
							req := &protobuf.RequestMessage{
								Action: protobuf.Action_PING_PONG,
								ActionDetail: &protobuf.RequestMessage_PingMessage{
									PingMessage: "ping",
								},
							}
							sendErr := streamConnectClient1.Send(req)
							ip, err := getIPFromContext(streamConnectClient1.Context())
							if err != nil {
								level.Error(h.logger).Log("msg", "Failed to resolve IP", "err", streamConnectClient1.Context().Err())
							}
							if sendErr != nil {
								h.StreamMap.Delete(ip)
								level.Info(h.logger).Log("msg", "Server is Offline", "ip", ip)

								list, _ := h.GetServiceListByName(appName)
								if len(list) > 0 {
									least := h.DiscoverServicesFindLeast(h.NameSpace, h.Version, h.Env, appName)

									//在新的服务端重新登录
									if least != nil {
										h.serverMap.Store(userJid, least.IP)

										h.LoginAgain(least.IP, userJid, appName)
									}

								}
							}

						}
					}

				case mvccpb.PUT:
					tmp := &protobuf.AccountLogin{}
					proto.Unmarshal(ev.Kv.Value, tmp)
					h.serverMap.Store(tmp.UserJid, tmp.GrpcServer)
				}
			}
		}
		lock.Unlock(context.Background())
	}
}

func (h *Handler) whatsappWatcher() {
	for {
		reqModel := etcd.WatchReq{
			Key:     "/services/" + h.Env + "/" + h.NameSpace + "/whatsapp/",
			Options: []clientv3.OpOption{clientv3.WithPrefix()},
			ResChan: make(chan etcd.WatchRes),
		}
		h.watchChan <- reqModel
		for res := range reqModel.ResChan {
			for _, ev := range res.Events {
				switch ev.Type {
				case mvccpb.DELETE:
					userJid, _ := strconv.ParseUint(strings.Split(string(ev.Kv.Key), "/")[5], 10, 64)
					value, ok := h.serverMap.Load(userJid)
					if ok {
						v1, ok1 := h.StreamMap.Load(value.(string))
						if !ok1 {
							level.Info(h.logger).Log("msg", "get stream fail")
						}
						streamConnectClient1 := v1.(protobuf.ArthasStream_ConnectClient)

						if streamConnectClient1 == nil {
							//故障转移
							//services := p.DiscoverServices(p.Handler.NameSpace, p.Handler.Version, p.Handler.Env, "telegram")
							//exists := p.exists(services, func(s string) bool {
							//	return s == account.GrpcServer
							//})
							//if exists == false {
							least := h.DiscoverServicesFindLeast(h.NameSpace, h.Version, h.Env, "whatsapp")
							h.serverMap.Store(userJid, least)

							//在新的服务端重新登录
							if least != nil {
								h.LoginAgain(least.IP, userJid, "whatsapp")
							}

						} else {
							req := &protobuf.RequestMessage{
								Action: protobuf.Action_PING_PONG,
								ActionDetail: &protobuf.RequestMessage_PingMessage{
									PingMessage: "ping",
								},
							}
							sendErr := streamConnectClient1.Send(req)
							if sendErr != nil {
							}
						}

					}

				case mvccpb.PUT:
					//fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
					tmp := &protobuf.AccountLogin{}
					proto.Unmarshal(ev.Kv.Value, tmp)
					h.serverMap.Store(tmp.UserJid, tmp.GrpcServer)
				}
			}
		}
	}
}

func (h *Handler) LoginAgain(ip string, userJid uint64, appName string) {
	time.Sleep(2 * time.Second)
	h.SyncInfo(ip, userJid, appName)

	time.Sleep(2 * time.Second)

	h.Login(ip, userJid, appName)
}
func (h *Handler) SyncInfo(ip string, userJid uint64, appName string) {
	if appName == "telegram" {

		v1, ok := h.StreamMap.Load(ip)
		if !ok {
			level.Info(h.logger).Log("msg", "get stream fail")
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			key := fmt.Sprintf("/arthas/tg_accountdetail/%v", strconv.FormatUint(userJid, 10))
			get := etcd.GetReq{
				Key:     key,
				Options: []clientv3.OpOption{clientv3.WithPrefix()},
				ResChan: make(chan etcd.GetRes),
			}
			h.getChan <- get
			result := <-get.ResChan

			if result.Result.Kvs != nil {
				tmp := &protobuf.TgAccountDetail{}
				proto.Unmarshal(result.Result.Kvs[0].Value, tmp)

				appData := make(map[uint64]*protobuf.AppData)
				user, _ := strconv.ParseUint(tmp.PhoneNumber, 10, 64)
				appData[user] = &protobuf.AppData{AppId: tmp.AppId, AppHash: tmp.AppHash}

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_SYNC_APP_INFO,
					Type:   appName,
					ActionDetail: &protobuf.RequestMessage_SyncAppAction{
						SyncAppAction: &protobuf.SyncAppInfoAction{
							AppData: appData,
						},
					},
				}
				if err := streamConnectClient.Send(req); err != nil {
					level.Error(h.logger).Log("msg", "streamConnectClient send", "err", err)
				}
			}
		}

	} else if appName == "whatsapp" {

		v1, ok := h.StreamMap.Load(ip)
		if !ok {
			level.Info(h.logger).Log("msg", "get stream fail")
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			key := fmt.Sprintf("/arthas/accountdetail/%v", strconv.FormatUint(userJid, 10))
			get := etcd.GetReq{
				Key:     key,
				Options: []clientv3.OpOption{clientv3.WithPrefix()},
				ResChan: make(chan etcd.GetRes),
			}
			h.getChan <- get
			result := <-get.ResChan

			if result.Result.Kvs != nil {
				tmp := &protobuf.AccountDetail{}
				proto.Unmarshal(result.Result.Kvs[0].Value, tmp)

				keyData := make(map[uint64]*protobuf.KeyData)

				pk := tmp.PrivateKey
				pkm := tmp.PrivateMsgKey
				pb := tmp.PublicKey
				pbm := tmp.PublicMsgKey
				identify := tmp.Identify
				user := tmp.UserJid
				keyData[user] = &protobuf.KeyData{Privatekey: pk, PrivateMsgKey: pkm, Publickey: pb, PublicMsgKey: pbm, Identify: identify}

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_SYNC_ACCOUNT_KEY,
					Type:   "whatsapp",
					ActionDetail: &protobuf.RequestMessage_SyncAccountKeyAction{
						SyncAccountKeyAction: &protobuf.SyncAccountKeyAction{
							KeyData: keyData,
						},
					},
				}
				if err := streamConnectClient.Send(req); err != nil {
					level.Error(h.logger).Log("msg", "streamConnectClient send", "err", err)
				}
			}
		}

	}

}

func (h *Handler) Login(ip string, userJid uint64, appName string) {

	v1, ok := h.StreamMap.Load(ip)
	if !ok {
		level.Info(h.logger).Log("msg", "get stream fail")
	} else {
		streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)
		//登录信息
		get := etcd.GetReq{
			Key:     "/service/proxyaddr/" + strconv.FormatUint(userJid, 10),
			Options: []clientv3.OpOption{clientv3.WithPrefix()},
			ResChan: make(chan etcd.GetRes),
		}
		h.getChan <- get
		result := <-get.ResChan

		loginDetail := make(map[uint64]*protobuf.LoginDetail)
		ld := &protobuf.LoginDetail{
			ProxyUrl: string(result.Result.Kvs[0].Value),
		}
		loginDetail[userJid] = ld
		login := &protobuf.RequestMessage{
			Action: protobuf.Action_LOGIN,
			Type:   appName,
			ActionDetail: &protobuf.RequestMessage_OrdinaryAction{
				OrdinaryAction: &protobuf.OrdinaryAction{
					LoginDetail: loginDetail,
				},
			},
		}

		if err := streamConnectClient.Send(login); err != nil {
			level.Error(h.logger).Log("msg", "streamConnectClient send", "err", err)
		}
	}

}

func (h *Handler) restartOnline(appName string, list []string) {
	getOnline := etcd.GetReq{
		Key:     "/service/" + h.Env + "/" + h.NameSpace + "/online/",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	h.getChan <- getOnline
	onlineResult := <-getOnline.ResChan
	if onlineResult.Result.Kvs != nil {
		for _, onlineValue := range onlineResult.Result.Kvs {
			curIp := strings.Split(string(onlineValue.Key), "/")[5]
			userString := strings.Split(string(onlineValue.Key), "/")[6]
			user, _ := strconv.ParseUint(userString, 10, 64)

			getTg := etcd.GetReq{
				Key:     "/services/" + h.Env + "/" + h.NameSpace + "/telegram/" + userString,
				Options: nil,
				ResChan: make(chan etcd.GetRes),
			}
			h.getChan <- getTg
			getResult := <-getTg.ResChan
			if getResult.Result.Kvs == nil && len(list) != 0 {
				availableList := ContainsString(list, curIp)
				if len(list) == len(availableList) {
					index := RandomInt(len(list))
					h.serverMap.Store(user, list[index])
					h.LoginAgain(list[index], user, appName)
				} else if len(list) > len(availableList) && len(availableList) != 0 {
					index := RandomInt(len(availableList))
					h.serverMap.Store(user, availableList[index])
					h.LoginAgain(availableList[index], user, appName)
				} else if len(list) > len(availableList) && len(availableList) == 0 {
					h.serverMap.Store(user, curIp)
					h.LoginAgain(curIp, user, appName)
				}
			}

		}
	}

}
