protoc:
	@protoc -I="shared" --go_out="server" "shared/packets.proto"
	@echo "Proto filed compiled successfully"
sqlgen:
	@sqlc generate -f server/internal/server/db/config/sqlc.yml