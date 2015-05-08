rm -rf raftlogs
rm -rf puddlestore/raftlogs
go build cli-client.go main-client.go shell.go
go build cli-node.go main-node.go shell.go
