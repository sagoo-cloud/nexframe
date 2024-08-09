package socketeer

import (
	"errors"
	"time"
)

var (
	pongWait           = 60 * time.Second
	pingPeriod         = (pongWait * 9) / 10
	writeWait          = 10 * time.Second
	maxMessageSize     = int64(512)
	maxReadBufferSize  = 1024
	maxWriteBufferSize = 1024
)
var ConnectionIdDoestExist = errors.New("ConnectionId Does not Exist")
