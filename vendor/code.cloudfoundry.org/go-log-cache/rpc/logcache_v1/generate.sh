#!/bin/bash

set -eu

dir_resolve() {
  cd "$1" 2>/dev/null || return $? # cd to desired directory; if fail, quell any error messages but return exit status
  echo "$(pwd -P)"                 # output full, link-resolved path
}

set -e

TARGET=$(dirname $0)
TARGET=$(dir_resolve $TARGET)
cd $TARGET

# move vendor out of the way to avoid vendoring anything due to the following go gets and installs
mv ../../vendor ../../vendor.bak
trap "mv ../../vendor.bak ../../vendor" EXIT

go get -d github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
go get -d google.golang.org/protobuf/cmd/protoc-gen-go
go get -d google.golang.org/grpc/cmd/protoc-gen-go-grpc

go install \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
  google.golang.org/protobuf/cmd/protoc-gen-go \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc

# install dependencies to a temp dir
tmp_dir=$(mktemp -d)
trap "rm -rf $tmp_dir" EXIT

git clone https://github.com/cloudfoundry/log-cache $tmp_dir/log-cache
mv $tmp_dir/log-cache/api/v1 $tmp_dir/logcache_v1

# v1.14.6 still contains the third_party/googleapi protos imported in
# egress.proto and promql.proto so we need this old version for the include path
git clone https://github.com/grpc-ecosystem/grpc-gateway -b v1.14.6 $tmp_dir/grpc-gateway

git clone https://github.com/cloudfoundry/loggregator-api $tmp_dir/loggregator-api

package=code.cloudfoundry.org/go-log-cache/rpc/logcache_v1

# invoke the generator using a distinct path for the proto files
# keeping them separate from ingress and egress.proto from loggregator-api,
# used by go-loggregator
protoc \
  logcache_v1/ingress.proto \
  logcache_v1/egress.proto \
  logcache_v1/orchestration.proto \
  logcache_v1/promql.proto \
  --proto_path=$tmp_dir/ \
  --go_out=Mlogcache_v1/promql.proto=$package,Mlogcache_v1/orchestration.proto=$package,Mlogcache_v1/ingress.proto=$package,Mlogcache_v1/egress.proto=$package,Mv2/envelope.proto=code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2:.. \
  --go_opt=paths=source_relative \
  --go-grpc_out=Mlogcache_v1/promql.proto=$package,Mlogcache_v1/orchestration.proto=$package,Mlogcache_v1/ingress.proto=$package,Mlogcache_v1/egress.proto=$package,Mv2/envelope.proto=code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2:.. \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=logtostderr=true,Mlogcache_v1/promql.proto=$package,Mlogcache_v1/orchestration.proto=$package,Mlogcache_v1/ingress.proto=$package,Mlogcache_v1/egress.proto=$package,Mv2/envelope.proto=code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2:.. \
  --grpc-gateway_opt=paths=source_relative \
  -I=$tmp_dir/grpc-gateway/third_party/googleapis \
  -I=$tmp_dir/loggregator-api
