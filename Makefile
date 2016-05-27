GO_SRC = $(wildcard *.go) handlers.go

JSON_SUFFIX = _easyjson.go
JSON_SRC = model/models.go packets.go
JSON_GEN = $(addsuffix $(JSON_SUFFIX), $(basename $(JSON_SRC)))

cord: $(JSON_GEN) $(GO_SRC) check
	@printf " ✔ finished building %s \n" $@

json: $(JSON_GEN)

%$(JSON_SUFFIX): $(JSON_SRC)
ifeq (, $(shell which easyjson))
	@printf " → Installing easyjson\n"
	@go get github.com/mailru/easyjson/easyjson
endif
	@printf " → Generating %s \n" $@
	@rm -f $@
	@easyjson -all $^

handlers.go:
	@go run ./cmd/genhands/main.go \
		CHANNEL_CREATE=Channel \
		CHANNEL_UPDATE=Channel \
		CHANNEL_DELETE=Channel \
		GUILD_CREATE=Guild \
		GUILD_UPDATE=Guild \
		GUILD_DELETE=Guild \
		GUILD_BAN_ADD=Guild \
		GUILD_MEMBER_ADD=Member \
		GUILD_MEMBER_UPDATE=Member \
		GUILD_MEMBER_REMOVE=Member \
		GUILD_ROLE_CREATE=GuildRole \
		GUILD_ROLE_UPDATE=GuildRole \
		GUILD_ROLE_DELETE=GuildRoleDelete \
		GUILD_INTEGRATIONS_UPDATE=GuildIntegrationsUpdate \
		GUILD_EMOJIS_UPDATE=GuildEmojisUpdate \
		MESSAGE_ACK=MessageAck \
		MESSAGE_CREATE=Message \
		MESSAGE_UPDATE=Message \
		MESSAGE_DELETE=Message \
		PRESENCE_UPDATE=PresenceUpdate \
		PRESENCES_REPLACE=PresencesReplace \
		READY=Ready \
		USER_UPDATE=User \
		USER_SETTINGS_UPDATE=UserSettingsUpdate \
		USER_GUILD_SETTINGS_UPDATE=UserGuildSettings \
		TYPING_START=TypingStart \
		VOICE_SERVER_UPDATE=VoiceServerUpdate \
		VOICE_STATE_UPDATE=VoiceState

	@printf " → Generating handler functions \n" $@

check: $(JSON_GEN) $(GO_SRC)
	@go test ./...
	@go vet ./...
	@printf " → Tests are green\n"

clean:
	rm -f $(JSON_GEN) handlers.go

.PHONY: clean check json handlers.go
