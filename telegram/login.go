package telegram

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"arthas/telegram/kv"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/fatih/color"
	"github.com/go-kit/log/level"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
	"strings"
	"time"
)

func (m *Manager) loginWithCode(detail LoginDetail) {
	tgUserLoginCallback := make([]callback.TgUserLoginCallback, 0)

	value, ok := m.accountsSync.Load(detail.PhoneNumber)
	if !ok {
		level.Info(m.logger).Log("msg", "TG Account is not exist", "user", detail.PhoneNumber)
		return
	}
	account := value.(*Account)
	account.proxyUrl = detail.ProxyUrl
	if account.IsLogin == true {
		level.Info(m.logger).Log("msg", "TG Account is online,will not continue", "user", &detail.PhoneNumber)
		return
	}
	var c *telegram.Client

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
		//Logger:  log.Named("gaps"),
	})

	appId, err := uint64ToIntSafe(account.AppId)
	if err != nil {
		level.Info(m.logger).Log("msg", "uint64toInt", "err", err)
	}
	m.kvd.WithNs(detail.PhoneNumber)

	// 接收消息，私聊，群聊
	d.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok1 := update.Message.(*tg.Message)
		if !ok1 {
			return nil
		}
		self, err := c.Self(ctx)
		if err != nil {
			return err
		}
		//fmt.Println("msg:", msg)
		m.sendTgMsgCallback([]callback.TgMsgCallback{covertMessage(self, msg)})
		return nil
	})

	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		msg, ok1 := update.Message.(*tg.Message)
		if !ok1 {
			return nil
		}
		self, err := c.Self(ctx)
		if err != nil {
			return err
		}

		//fmt.Println("msg:", msg)
		m.sendTgMsgCallback([]callback.TgMsgCallback{covertMessage(self, msg)})
		return nil
	})

	c, kvd, err := m.Initialization(m.context, true, appId, account.AppHash, account, gaps)
	if err != nil {
		level.Error(m.logger).Log("msg", "init client fail", "err", err)
		if detail.ResChan != nil {
			res := LoginDetailRes{
				LoginStatus: protobuf.AccountStatus_FAIL,
				Comment:     "login fail:" + err.Error(),
			}
			detail.ResChan <- res
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := c.Run(ctx, func(ctx context.Context) error {
		if err = c.Ping(ctx); err != nil {
			return err
		}
		status, err := c.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if status.Authorized {

		} else {

			flow := auth.NewFlow(&tgAuth{phoneNum: detail.PhoneNumber, watchChan: m.whatChan, loginId: detail.LoginId, sendCodeChan: m.sendCodeChan, loginDetailResChan: detail.ResChan}, auth.SendCodeOptions{})

			if err = c.Auth().IfNecessary(ctx, flow); err != nil {
				level.Error(m.logger).Log("msg", "auth necessary fail", "err", err)
				resChan := flow.Auth.(*tgAuth).ResChan
				if resChan != nil {
					// 将login的chan返回给设为空，那么下面就不会再进返回多一次了
					detail.ResChan = nil
					resChan <- LoginDetailRes{
						LoginStatus: protobuf.AccountStatus_LOGIN_CODE_FAIL,
						Comment:     "login fail:" + err.Error(),
					}
				}
				cancel()
				return err
			}
			resChan := flow.Auth.(*tgAuth).ResChan
			if resChan != nil {
				// 将login的chan返回给设为空，那么下面就不会再进返回多一次了
				detail.ResChan = nil
				resChan <- LoginDetailRes{
					LoginStatus: protobuf.AccountStatus_SUCCESS,
					Comment:     "login success..",
				}
				cancel()
			}

		}

		color.Yellow("WARN: Using the built-in APP_ID & APP_HASH may increase the probability of blocking")
		color.Blue("Login...")

		user, err := c.Self(ctx)
		if err != nil {
			level.Error(m.logger).Log("msg", "get user info fail", "err", err)
			return err
		}

		if err = kvd.Set(key.App(), []byte(consts.AppBuiltin)); err != nil {
			level.Error(m.logger).Log("msg", "get user info fail", "err", err)
			return err
		}

		account.IsLogin = true
		account.UserID = user.ID
		account.UserName = user.Username
		account.LastName = user.LastName
		account.FirstName = user.FirstName
		m.recordConnection(true)
		m.etcdAccountLoginState(account.PhoneNumber, account.IsLogin, detail.ProxyUrl, user.ID, user.Username, user.FirstName, user.LastName)
		m.accountsSync.Store(detail.PhoneNumber, account)

		color.Blue("Login successfully! ID: %d, Username: %s", user.ID, user.Username)

		// 登录回调
		msgCallback := callback.TgUserLoginCallback{
			IsSuccess:    1,
			TgId:         user.ID,
			Username:     user.Username,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Phone:        user.Phone,
			IsOnline:     1,
			ProxyAddress: detail.ProxyUrl,
			Comment:      "login success...",
		}
		tgUserLoginCallback = append(tgUserLoginCallback, msgCallback)
		if len(tgUserLoginCallback) > 0 {
			m.loginMsgCallback(tgUserLoginCallback)
		}

		if detail.ResChan != nil {
			res := LoginDetailRes{
				LoginStatus: protobuf.AccountStatus_SUCCESS,
				Comment:     "login success..",
			}
			detail.ResChan <- res
		}

		m.clientMap.Store(detail.PhoneNumber, c)
		//num, _ := strconv.ParseUint(detail.PhoneNumber, 10, 64)
		cancel()

		return gaps.Run(ctx, c.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				level.Info(m.logger).Log("Gaps started")
			},
		})

	}); err != nil {
		cancel()
		//errMsg := strings.Split(err.Error(), ":")
		//if errMsg[6] == "PHONE_CODE_INVALID"{
		//
		//}
		msgCallback := callback.TgUserLoginCallback{
			AccountStatus: 1,
			IsSuccess:     2,
			Phone:         detail.PhoneNumber,
			IsOnline:      2,
			ProxyAddress:  detail.ProxyUrl,
			Comment:       "login fali:" + err.Error(),
		}
		tgUserLoginCallback = append(tgUserLoginCallback, msgCallback)
		if len(tgUserLoginCallback) > 0 {
			m.loginMsgCallback(tgUserLoginCallback)
		}

		level.Error(m.logger).Log("msg", "login fail", "err", err)
		if detail.ResChan != nil {
			res := LoginDetailRes{}
			if strings.Contains(err.Error(), "400: PHONE_NUMBER_INVALID") {
				// 无效手机号
				res.LoginStatus = protobuf.AccountStatus_NOT_EXIST
				res.Comment = "login fail:" + err.Error()
			} else if strings.Contains(err.Error(), "400: PHONE_NUMBER_BANNED") {
				res.LoginStatus = protobuf.AccountStatus_SEAL
				res.Comment = "login fail:" + err.Error()
				fmt.Println("删除etcd online账号" + detail.PhoneNumber)
				_ = m.OffOnlineAccount(detail.PhoneNumber) // 删除etcd启动登录的账号
			} else if strings.Contains(err.Error(), "corrupted key") {
				res.LoginStatus = protobuf.AccountStatus_LOGIN_SESSION_CORRUPTED
				res.Comment = "login fail:" + err.Error()
				_ = m.OffOnlineAccount(detail.PhoneNumber)
			} else {
				res.LoginStatus = protobuf.AccountStatus_FAIL
				res.Comment = "login fail:" + err.Error()
				_ = m.OffOnlineAccount(detail.PhoneNumber)

			}

			detail.ResChan <- res
			return
		}
	}
	return

}

func (m *Manager) Initialization(ctx context.Context, login bool, appId int, appHash string, account *Account, d *updates.Manager, middlewares ...telegram.Middleware) (*telegram.Client, kv.KV, error) {
	account.AppHash = appHash
	account.AppId = uint64(appId)
	mode, err := m.kvd.Get(key.App())
	if err != nil {
		mode = []byte(consts.AppBuiltin)
	} else {
		if string(mode) != "" {
			login = false
		}
	}
	// etcd获取device信息
	m.getAccountDevice(account)

	opts := telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{
			Dial: utils.Proxy.GetDial(account.proxyUrl).DialContext,
			//Dial: utils.Proxy.GetDial("asdasdas").DialContext,
		}),
		ReconnectionBackoff: func() backoff.BackOff {
			b := backoff.NewExponentialBackOff()

			b.Multiplier = 1.1
			b.MaxElapsedTime = viper.GetDuration(consts.FlagReconnectTimeout)
			//b.Clock = _clock
			return b
		},
		Device:         account.Device,
		SessionStorage: storage.NewSession(m.kvd, login),
		RetryInterval:  5 * time.Second,
		MaxRetries:     5,
		DialTimeout:    10 * time.Second,
		Middlewares:    middlewares,

		UpdateHandler: d,
	}
	client := telegram.NewClient(int(account.AppId), account.AppHash, opts)
	//if opts.SessionStorage != nil {
	//	client.SetClientStorage(&store.ArtsTgLoader{storage.NewSession(m.kvd, login)})
	//}
	//client = telegram.SetSessionStorage(client, &store.ArtsTgLoader{storage.NewSession(m.kvd, login)})
	return client, m.kvd, nil
}

func (m *Manager) getAccountDevice(account *Account) *Account {
	key := fmt.Sprintf("/tg/%v/device/json", account.PhoneNumber)
	deviceRes := &AccountDevice{}
	getReq := etcd.GetReq{
		Key:     key,
		Options: nil,
		ResChan: make(chan etcd.GetRes),
	}
	m.getChan <- getReq
	getResult := <-getReq.ResChan
	if getResult.Result.Kvs != nil {
		// 如果不为空，则赋值给account
		_ = json.Unmarshal(getResult.Result.Kvs[0].Value, deviceRes)
		account.AppId = deviceRes.AppId
		account.AppHash = deviceRes.AppHash
		account.Device = telegram.DeviceConfig{
			DeviceModel:    deviceRes.DeviceModel,
			SystemVersion:  deviceRes.SystemVersion,
			AppVersion:     deviceRes.AppVersion,
			LangCode:       deviceRes.LangCode,
			SystemLangCode: deviceRes.SystemLangCode,
			LangPack:       deviceRes.LangPack,
		}
	} else {
		// 如果没有，传入固定的,这是传入到etcd的数据

		deviceRes.DeviceModel = consts.Device.DeviceModel
		deviceRes.SystemVersion = consts.Device.SystemVersion
		deviceRes.AppVersion = consts.Device.AppVersion
		deviceRes.DeviceModel = consts.Device.DeviceModel
		deviceRes.LangPack = consts.Device.LangPack
		deviceRes.SystemLangCode = consts.Device.SystemLangCode
		deviceRes.AppId = account.AppId
		deviceRes.AppHash = account.AppHash

		if account.AppHash == "" {
			account.AppHash = appHash
			deviceRes.AppHash = appHash
		}
		if account.AppId == 0 {
			account.AppId = appID
			deviceRes.AppId = appID
		}

		// 将信息传入到map内存
		a := &AccountDetail{
			PhoneNumber: account.PhoneNumber,
			AppId:       account.AppId,
			AppHash:     account.AppHash,
		}
		m.accountsSync.Store(account.PhoneNumber, a)

		deviceJ, _ := json.Marshal(deviceRes)

		putReq := etcd.PutReq{
			Key:     key,
			Value:   string(deviceJ),
			Options: nil,
			ResChan: make(chan etcd.PutRes),
		}
		m.putChan <- putReq

		// 再赋值给account,传入的是默认的设备值
		account.Device = consts.Device
	}

	return account
}

func (m *Manager) recordConnection(re bool) {
	if re == true {
		key := fmt.Sprintf("/service/%v/%v/%v/%v/%v/%v", m.env, m.ns, "count", m.appName, m.vs, m.ListenAddress)
		//记录连接数
		putValue := &protobuf.ServiceInfo{}
		getRes := &protobuf.ServiceInfo{}
		getReq := etcd.GetReq{
			Key:     key,
			Options: nil,
			ResChan: make(chan etcd.GetRes),
		}
		m.getChan <- getReq
		getResult := <-getReq.ResChan
		if getResult.Result.Kvs != nil {
			proto.Unmarshal(getResult.Result.Kvs[0].Value, getRes)
			tmp := &protobuf.ServiceInfo{
				IP:          m.ListenAddress,
				Connections: getRes.Connections + 1,
			}
			putValue = tmp
		}

		putvalueb, _ := proto.Marshal(putValue)

		putReq := etcd.PutReq{
			Key:     key,
			Value:   string(putvalueb),
			Options: nil,
			ResChan: make(chan etcd.PutRes),
		}
		m.putChan <- putReq
	}

}
