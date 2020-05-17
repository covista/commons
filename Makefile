.PHONY: proto

proto: proto/commons.proto
	protoc -I proto/ proto/commons.proto --go_out=plugins=grpc:proto
