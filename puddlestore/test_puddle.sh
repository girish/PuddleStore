rm -rf raftlogs
rm -rf puddlestore/raftlogs
go test puddlestore/client.go \
        puddlestore/guid.go   \
        puddlestore/inode.go  \
        puddlestore/listener.go \
        puddlestore/logging.go \
        puddlestore/puddle_local_impl.go \
        puddlestore/puddle_rpc_api.go \
        puddlestore/puddle_rpc_impl.go \
        puddlestore/puddlestore.go \
        puddlestore/puddlestore_test.go \
        puddlestore/util.go
