package discover

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFormatPath(t *testing.T) {
	s1 := formatPath("/")
	if s1 != "/" {
		t.Fail()
	}
	s2 := formatPath("/1")
	if s2 != "/1" {
		t.Fail()
	}
	s3 := formatPath("/1/")
	if s3 != "/1" {
		t.Fail()
	}
}

func TestRegister(t *testing.T) {
	s, err := NewService("test", "test_type", "test1", "test_ip", []string{"127.0.0.1:2379"})
	if err != nil {
		fmt.Println("new service err:", err)
	}
	ser, _ := s.(*Service)
	fmt.Println("path:", ser.servicePath)
	fmt.Println("type:", ser.serviceType)
	fmt.Println("name:", ser.name)
	fmt.Println("fullpath:", ser.fullPath)
	s.Register()
	time.Sleep(5 * time.Second)
	res, err := ser.client.Get(context.TODO(), ser.fullPath)
	if err != nil {
		fmt.Println("client.Get err:", err)
	}
	fmt.Printf("res get :%#v\n", res.Kvs)
	if len(res.Kvs) != 1 {
		t.Fail()
	}
	kv := res.Kvs[0]
	if string(kv.Key) != ser.fullPath {
		t.Fail()
	}
	if string(kv.Value) != "test_ip" {
		t.Fail()
	}
}

func TestWatchAfter(t *testing.T) {
	s1, err := NewService("test", "test_type", "test1", "test_ip1", []string{"127.0.0.1:2379"})
	if err != nil {
		fmt.Println("new service err:", err)
	}
	s2, err := NewService("test", "test_type", "test2", "test_ip2", []string{"127.0.0.1:2379"})
	if err != nil {
		fmt.Println("new service err:", err)
	}
	s2.WatchNodes("/test/test_type")

	s1.Register()
	time.Sleep(5 * time.Second)

	serverString := s2.GetServiceType("test_type")

	if serverString != "test_ip1" {
		t.Fail()
	}
}

func TestWatchBefore(t *testing.T) {
	s1, err := NewService("test", "test_type", "test1", "test_ip1", []string{"127.0.0.1:2379"})
	if err != nil {
		fmt.Println("new service err:", err)
	}
	s1.Register()
	s2, err := NewService("test", "test_type", "test2", "test_ip2", []string{"127.0.0.1:2379"})
	if err != nil {
		fmt.Println("new service err:", err)
	}
	s2.WatchNodes("/test/test_type")
	time.Sleep(5 * time.Second)

	serverString := s2.GetServiceType("test_type")

	if serverString != "test_ip1" {
		t.Fail()
	}
}
