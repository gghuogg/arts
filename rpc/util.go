package rpc

import (
	"arthas/etcd"
	"arthas/protobuf"
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ContainsString(slice []string, target string) []string {
	var result []string
	found := false // 用于标记是否找到目标元素

	for _, element := range slice {
		if element != target {
			result = append(result, element)
		} else {
			found = true
		}
	}

	if found {
		return result
	}

	// 如果目标元素未找到，返回原始切片
	return slice

}

func RandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	rand.Seed(time.Now().UnixNano()) // 使用当前时间作为随机种子
	return rand.Intn(max)
}

func getIPFromContext(ctx context.Context) (string, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return "", errors.New("failed to get peer info from context")
	}

	addr := peer.Addr
	switch a := addr.(type) {
	case *net.TCPAddr:
		// IPv4 or IPv6
		return a.AddrPort().String(), nil
	default:
		return "", errors.New("unsupported address type")
	}
}

func (p *ProxyServer) getSyncMap(key string) *sync.Map {
	p.mu.Lock()
	defer p.mu.Unlock()

	m, exists := p.taskSyncMaps[key]
	if !exists {
		m = &sync.Map{}
		p.taskSyncMaps[key] = m
	}
	return m
}

func (p *ProxyServer) DiscoverServices(space, version, env, name string) []string {
	var services []string
	if name != "" {
		get := etcd.GetReq{
			Key:     "/service/" + env + "/" + space + "/arthas/" + name,
			Options: []clientv3.OpOption{clientv3.WithPrefix()},
			ResChan: make(chan etcd.GetRes),
		}
		//fmt.Println(get.Key)
		p.Handler.getChan <- get
		result := <-get.ResChan

		if result.Result.Kvs != nil {
			for _, v := range result.Result.Kvs {
				serviceInfo := &protobuf.ServiceInfo{}
				proto.Unmarshal(v.Value, serviceInfo)
				services = append(services, serviceInfo.IP)
			}
		}
	} else {
		get := etcd.GetReq{
			Key:     "/service/" + env + "/" + space + "/arthas/",
			Options: []clientv3.OpOption{clientv3.WithPrefix()},
			ResChan: make(chan etcd.GetRes),
		}
		p.Handler.getChan <- get
		result := <-get.ResChan

		if result.Result.Kvs != nil {
			for _, v := range result.Result.Kvs {
				serviceInfo := &protobuf.ServiceInfo{}
				proto.Unmarshal(v.Value, serviceInfo)
				services = append(services, serviceInfo.IP)
			}
		}
	}

	return services
}

func (p *ProxyServer) GetServiceListByName(name string) ([]string, error) {

	services := p.DiscoverServices(p.Handler.NameSpace, p.Handler.Version, p.Handler.Env, name)

	if len(services) == 0 {
		return nil, errors.New("no services available")
	}

	return services, nil
}

func (h *Handler) DiscoverServices(space, version, env, name string) []string {
	var services []string
	if name != "" {
		get := etcd.GetReq{
			Key:     "/service/" + env + "/" + space + "/arthas/" + name,
			Options: []clientv3.OpOption{clientv3.WithPrefix()},
			ResChan: make(chan etcd.GetRes),
		}
		//fmt.Println(get.Key)
		h.getChan <- get
		result := <-get.ResChan

		if result.Result.Kvs != nil {
			for _, v := range result.Result.Kvs {
				serviceInfo := &protobuf.ServiceInfo{}
				proto.Unmarshal(v.Value, serviceInfo)
				services = append(services, serviceInfo.IP)
			}
		}
	} else {
		get := etcd.GetReq{
			Key:     "/service/" + env + "/" + space + "/arthas/",
			Options: []clientv3.OpOption{clientv3.WithPrefix()},
			ResChan: make(chan etcd.GetRes),
		}
		h.getChan <- get
		result := <-get.ResChan

		if result.Result.Kvs != nil {
			for _, v := range result.Result.Kvs {
				serviceInfo := &protobuf.ServiceInfo{}
				proto.Unmarshal(v.Value, serviceInfo)
				services = append(services, serviceInfo.IP)
			}
		}
	}

	return services
}

func (h *Handler) DiscoverServicesFindLeast(space, version, env, name string) *protobuf.ServiceInfo {

	var leastConnectedService *protobuf.ServiceInfo
	getAlive := etcd.GetReq{
		Key:     "/service/" + env + "/" + space + "/arthas/",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	h.getChan <- getAlive
	resultAlive := <-getAlive.ResChan

	get := etcd.GetReq{
		Key:     "/service/" + env + "/" + space + "/count/" + name,
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	h.getChan <- get
	result := <-get.ResChan

	if resultAlive.Result.Kvs != nil && result.Result.Kvs != nil {
		minConnections := uint32(math.MaxUint32)
		for _, v := range result.Result.Kvs {
			for _, v1 := range resultAlive.Result.Kvs {
				serviceInfo := &protobuf.ServiceInfo{}
				proto.Unmarshal(v.Value, serviceInfo)

				serviceInfo1 := &protobuf.ServiceInfo{}
				proto.Unmarshal(v1.Value, serviceInfo1)

				if serviceInfo.IP == serviceInfo1.IP {
					if serviceInfo.Connections < minConnections {
						minConnections = serviceInfo.Connections
						leastConnectedService = serviceInfo
					}
				}
			}

		}
	}

	return leastConnectedService
}

func (h *Handler) GetServiceListByName(name string) ([]string, error) {

	services := h.DiscoverServices(h.NameSpace, h.Version, h.Env, name)

	if len(services) == 0 {
		return nil, errors.New("no services available")
	}

	return services, nil
}

func (h *Handler) FindAccountList() (result []uint64) {
	result = make([]uint64, 0)

	get := etcd.GetReq{
		Key:     "/service/" + h.Env + "/" + h.NameSpace + "/online/",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	h.getChan <- get
	etcdResult := <-get.ResChan

	if etcdResult.Result.Kvs != nil {
		for _, v := range etcdResult.Result.Kvs {
			tmp := strings.Split(string(v.Key), "/")[6]
			phone, _ := strconv.ParseUint(tmp, 10, 64)
			result = append(result, phone)
		}
	}
	return result
}
