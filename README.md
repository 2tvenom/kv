### Build 
`go build ./cmd/kv-server`
`go build ./cmd/kv-bench`

### Run test
`./genca.sh`

`go test -v ./...`

Auth certificates

`./genca.sh`
### Run
`./kv-server`

### Bench
`./kv-server`

`./kv-bench`
### Requests
#### http / tcp (ncat required)
Set key

`curl -d 'SET key value' http://localhost:4500`

`echo "SET key value" | ncat 127.0.0.1 4501`

Set with ttl

`curl -d 'SET key 10 value' http://localhost:4500`

`echo "SET key 10 value" | ncat 127.0.0.1 4501`

Get key
 
`curl -d 'GET key' http://localhost:4500`

`echo "GET key" | ncat 127.0.0.1 4501`

Set list

`curl -d 'SETLIST key aa bb hh hh' http://localhost:4500`

`echo "SETLIST key aa bb hh hh" | ncat 127.0.0.1 4501`

Get list

`curl -d 'GETLIST key' http://localhost:4500`

`echo "GETLIST key" | ncat 127.0.0.1 4501`

Get list element

`curl -d 'GETLISTELEM key 1' http://localhost:4500`

`echo "GETLISTELEM key 1" | ncat 127.0.0.1 4501`

Set dictionary

`curl -d 'SETDICT key foo:aa baz:bar bar:foo zz:hello a:first' http://localhost:4500`

`echo "SETDICT key foo:aa baz:bar bar:foo zz:hello a:first" | ncat 127.0.0.1 4501`

Get dictionary

`curl -d 'GETDICT key' http://localhost:4500`

`echo "GETDICT key" | ncat 127.0.0.1 4501`

Get dictionary element

`curl -d 'GETDICTELEM key bar' http://localhost:4500`

`curl -d 'GETDICTELEM key a' http://localhost:4500`

`echo "GETDICTELEM key bar" | ncat 127.0.0.1 4501`

`echo "GETDICTELEM key a" | ncat 127.0.0.1 4501`

Get keys

`curl -d 'KEYS' http://localhost:4500`

`echo "KEYS" | ncat 127.0.0.1 4501`

Remove key

`curl -d 'REMOVE key' http://localhost:4500`

`echo "REMOVE key" | ncat 127.0.0.1 4501`


#### Auth request

Run server with keys  
`./kv-server -secure -cert-path ca.crt -key-path ca.key`

`curl -d "SETDICT aaa foo:baz bar:foo baz:bazbaz aa:foo_baz" --cert client.crt --key client.key -k "https://localhost:4500/"`