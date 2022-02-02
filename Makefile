gen:
	protoc --proto_path=api/proto/v1 --proto_path=third_party --go_out=plugins=grpc:pkg/api/v1 todo-service.proto
	protoc --proto_path=api/proto/v1 --proto_path=third_party --grpc-gateway_out=logtostderr=true:pkg/api/v1 todo-service.proto
	protoc --proto_path=api/proto/v1 --proto_path=third_party --swagger_out=logtostderr=true:api/swagger/v1 todo-service.proto

buildapi:
	cd cmd/server && go build .
	
runapi: buildapi
	cd cmd/server && ./server.exe \
		-grpc-port=9090 -http-port=8080 -db-host=localhost:3306 -db-user=root \
		-db-password=password -db-schema=todo -log-level=-1 -log-time-format=2006-01-02T15:04:05.999999999Z07:00

run-client-grpc:
	cd cmd/client-grpc && go build . && ./client-grpc.exe -server=localhost:9090
	
run-client-rest:
	cd cmd/client-rest && go build . && ./client-rest.exe -server=http://localhost:8080