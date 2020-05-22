PROTO_GENFILES = proto/commons.pb.go proto/commons.pb.gw.go

commons-server: $(PROTO_GENFILES) cmd/commons/main.go
	go build -o commons-server cmd/commons/main.go
	cp commons-server docker/commons-server/.

$(PROTO_GENFILES): proto/commons.proto
	protoc -I proto/ -I grpc-gateway/third_party/googleapis proto/commons.proto --go_out=plugins=grpc:proto --grpc-gateway_out=logtostderr=true:proto --swagger_out=logtostderr=true:swagger
	python3 -m grpc_tools.protoc -I proto -I grpc-gateway/third_party/googleapis --python_out=simulation/. --grpc_python_out=simulation/. proto/commons.proto
