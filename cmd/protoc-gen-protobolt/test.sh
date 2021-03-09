
go install .;
mkdir -p test;
protoc -I="../../proto/v1" -I=./ --go_out=paths=source_relative:./test/ \
  --protobolt_out=source_relative:./test/ ./test.proto;

# Check that the generated packages contains valid go code.
go build ./test;
