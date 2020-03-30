package discover

// ServerInfo server info with etcd full path and val (ip info)
type ServerInfo struct {
	key string
	val string
}

// GetKey GetKey
func (s *ServerInfo) GetKey() string { return s.key }

// GetValue GetValue
func (s *ServerInfo) GetValue() string { return s.val }
