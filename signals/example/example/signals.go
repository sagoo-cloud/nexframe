package example

import "github.com/sagoo-cloud/nexframe/signals"

// 异步信号
var RecordCreated = signals.New[Record]()
var RecordUpdated = signals.New[Record]()
var RecordDeleted = signals.New[Record]()

// 同步信号
var RecordCreatedSync = signals.NewSync[Record]()
var RecordUpdatedSync = signals.NewSync[Record]()
var RecordDeletedSync = signals.NewSync[Record]()
