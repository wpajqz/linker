package discover

import (
	"context"
	"errors"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// IServicer interface of a service
type IServicer interface {
	Stop()
	Register()
	WatchNodes(string)
	GetServiceType(string) string
}

// Service a service
type Service struct {
	ipInfo      string //etcd val
	stop        chan bool
	leaseid     clientv3.LeaseID
	client      *clientv3.Client
	servicePath string //register path
	serviceType string //my service type
	name        string //my service name
	fullPath    string //fullpath=s.servicePath + s.serviceType + s.name, etcd key

	// all services been watched except me
	// map[type]map[fullPath]ServerInfo
	services map[string]map[string]ServerInfo
	mu       sync.RWMutex
}

// NewService init with a parameter check,t is a srever type
func NewService(path, t, name, ip string, endpoints []string) (IServicer, error) {
	if path == "" {
		return nil, ErrorServicePathNil
	}
	if t == "" {
		return nil, ErrorServiceTypeNil
	}
	if name == "" {
		return nil, ErrorServiceNameNil
	}
	if ip == "" {
		return nil, ErrorServiceIPInfoNil
	}
	path = formatPath(path)
	t = formatPath(t)
	name = formatPath(name)
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 3 * time.Second,
	}
	etcdcli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Service{
		client:      etcdcli,
		stop:        make(chan bool, 1),
		ipInfo:      ip,
		servicePath: path,
		name:        name,
		serviceType: t,
		fullPath:    path + t + name,
		services:    map[string]map[string]ServerInfo{},
	}, nil

}

func (s *Service) Register() {
	go func() {
		_ = s.register()
	}()
}

// Register register me
func (s *Service) register() error {
	ch, err := s.keepAlive()
	if err != nil {
		return err
	}

	for {
		select {
		case <-s.stop:
			return s.revoke()
		case <-s.client.Ctx().Done():
			return errors.New("server closed")
		case _, ok := <-ch:
			if !ok {
				return s.revoke()
			}
		}
	}
}

// Stop 停止
func (s *Service) Stop() {
	s.stop <- true
}

// keepAlive 保持连接
func (s *Service) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {

	resp, err := s.client.Grant(context.TODO(), 5)
	if err != nil {
		return nil, err
	}

	_, err = s.client.Put(context.TODO(), s.fullPath, s.ipInfo, clientv3.WithLease(resp.ID))
	if err != nil {
		return nil, err
	}
	s.leaseid = resp.ID

	return s.client.KeepAlive(context.TODO(), resp.ID)
}

// revoke 取消
func (s *Service) revoke() error {
	_, err := s.client.Revoke(context.TODO(), s.leaseid)
	return err
}

func (s *Service) WatchNodes(path string) {
	go func() {
		_ = s.watchNodes(path)
	}()
}

// WatchNodes WatchNodes
func (s *Service) watchNodes(path string) error {
	if err := s.getServerFirst(path); err != nil {
		return err
	}
	rch := s.client.Watch(context.TODO(), path, clientv3.WithPrefix())

	for wresp := range rch {
		for _, ev := range wresp.Events {
			if string(ev.Kv.Key) == s.fullPath {
				continue
			}
			switch ev.Type {
			case clientv3.EventTypePut:
				s.addService(string(ev.Kv.Key), string(ev.Kv.Value))
			case clientv3.EventTypeDelete:
				s.removeService(string(ev.Kv.Key))
			}
		}
	}
	return nil
}

// add a service
func (s *Service) addService(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fullPath == key {
		return
	}

	// name check
	serviceName := strings.Replace(key, "\\", "/", -1)
	serviceType := filepath.Base(filepath.Dir(serviceName))

	// try new service kind init
	if s.services[serviceType] == nil {
		s.services[serviceType] = map[string]ServerInfo{}
	}

	service := s.services[serviceType]
	if _, ok := service[key]; ok {
		return
	}
	service[key] = ServerInfo{
		key: key,
		val: value,
	}
}

// remove a service
func (s *Service) removeService(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// name check
	serviceName := strings.Replace(key, "\\", "/", -1)
	serviceType := filepath.Base(filepath.Dir(serviceName))
	// check service kind
	service := s.services[serviceType]
	if service == nil {
		return
	}

	// remove a service
	_, ok := service[key]
	if !ok {
		return
	}
	delete(service, serviceName)
}

// GetServiceType get ip by server type return empty if don't have it
func (s *Service) GetServiceType(t string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.services[t]
	if !ok {
		return ""
	}
	list := []string{}
	for _, item := range m {
		list = append(list, item.GetValue())
	}
	return list[rand.Intn(len(list))]
}

// getServerOnce get watch path at first time
func (s *Service) getServerFirst(path string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	resp, err := s.client.Get(ctx, path, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return err
	}

	if resp != nil && resp.Kvs != nil {
		for _, item := range resp.Kvs {
			s.addService(string(item.Key), string(item.Value))
		}
	}
	return nil
}

func formatPath(path string) string {
	s1 := strings.TrimSuffix(path, "/")
	s2 := strings.TrimPrefix(s1, "/")
	return "/" + s2
}
