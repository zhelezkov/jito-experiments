protoc --go_out=./pkg/jito/gen --go_opt=paths=source_relative \
  --go-grpc_out=./pkg/jito/gen --go-grpc_opt=paths=source_relative \
  --proto_path=./mev-protos-master auth.proto block.proto block_engine.proto bundle.proto packet.proto searcher.proto shared.proto