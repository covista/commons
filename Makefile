.PHONY: proto

proto: proto/commons.proto
	protoc -I proto/ -I grpc-gateway/third_party/googleapis proto/commons.proto --go_out=plugins=grpc:proto --grpc-gateway_out=logtostderr=true:proto --swagger_out=logtostderr=true:swagger
