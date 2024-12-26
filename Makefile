protoc:
	@protoc -I="shared" --go_out="server" "shared/packets.proto"
	@echo "Proto filed compiled successfully"