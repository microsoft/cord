EVENTS = events/events.go
GO_SRC = $(wildcard *.go) $(EVENTS)

JSON_SUFFIX = _easyjson.go
JSON_SRC = model/models.go packets.go
JSON_GEN = $(addsuffix $(JSON_SUFFIX), $(basename $(JSON_SRC)))

all: events $(JSON_GEN) $(GO_SRC) check
	@printf " ✔ Finished %s \n" $@

json: $(JSON_GEN)

%$(JSON_SUFFIX): $(JSON_SRC)
ifeq (, $(shell which easyjson))
	@printf " → Installing easyjson\n"
	@go get github.com/mailru/easyjson/easyjson
endif
	@printf " → Generating %s \n" $@
	@rm -f $@
	@easyjson -all $^

events:
	@printf " → Generating %s \n" $@
	@go run ./cmd/genevents/main.go \
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
		VOICE_STATE_UPDATE=VoiceState > $(EVENTS)


check: $(JSON_GEN) $(GO_SRC)
	@printf " → Running tests\n"
	@go test -race ./...
	@go vet ./...

clean:
	rm -f $(JSON_GEN) $(EVENTS)

.PHONY: clean check json events
