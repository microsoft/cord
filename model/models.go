package model

import (
	"encoding/json"
	"time"
)

// A VoiceRegion stores data for a specific voice region server.
type VoiceRegion struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Hostname string `json:"sample_hostname"`
	Port     int    `json:"sample_port"`
}

// A VoiceICE stores data for voice ICE servers.
type VoiceICE struct {
	TTL     string       `json:"ttl"`
	Servers []*ICEServer `json:"servers"`
}

// A ICEServer stores data for a specific voice ICE server.
type ICEServer struct {
	URL        string `json:"url"`
	Username   string `json:"username"`
	Credential string `json:"credential"`
}

// A Invite stores all data related to a specific Discord Guild or Channel invite.
type Invite struct {
	Guild     *Guild   `json:"guild"`
	Channel   *Channel `json:"channel"`
	Inviter   *User    `json:"inviter"`
	Code      string   `json:"code"`
	CreatedAt string   `json:"created_at"` // TODO make timestamp
	MaxAge    int      `json:"max_age"`
	Uses      int      `json:"uses"`
	MaxUses   int      `json:"max_uses"`
	XkcdPass  bool     `json:"xkcdpass"`
	Revoked   bool     `json:"revoked"`
	Temporary bool     `json:"temporary"`
}

// A Channel holds all data related to an individual Discord channel.
type Channel struct {
	ID                   string                 `json:"id"`
	GuildID              string                 `json:"guild_id"`
	Name                 string                 `json:"name"`
	Topic                string                 `json:"topic"`
	Type                 string                 `json:"type"`
	LastMessageID        string                 `json:"last_message_id"`
	Position             int                    `json:"position"`
	Bitrate              int                    `json:"bitrate"`
	IsPrivate            bool                   `json:"is_private"`
	Recipient            *User                  `json:"recipient"`
	Messages             []*Message             `json:"-"`
	PermissionOverwrites []*PermissionOverwrite `json:"permission_overwrites"`
}

// A PermissionOverwrite holds permission overwrite data for a Channel
type PermissionOverwrite struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Deny  int    `json:"deny"`
	Allow int    `json:"allow"`
}

// Emoji struct holds data related to Emoji's
type Emoji struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Roles         []string `json:"roles"`
	Managed       bool     `json:"managed"`
	RequireColons bool     `json:"require_colons"`
}

// VerificationLevel type defination
type VerificationLevel int

// Constants for VerificationLevel levels from 0 to 3 inclusive
const (
	VerificationLevelNone VerificationLevel = iota
	VerificationLevelLow
	VerificationLevelMedium
	VerificationLevelHigh
)

// A Guild holds all data related to a specific Discord Guild.  Guilds are also
// sometimes referred to as Servers in the Discord client.
type Guild struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Icon              string            `json:"icon"`
	Region            string            `json:"region"`
	AfkChannelID      string            `json:"afk_channel_id"`
	EmbedChannelID    string            `json:"embed_channel_id"`
	OwnerID           string            `json:"owner_id"`
	JoinedAt          string            `json:"joined_at"` // make this a timestamp
	Splash            string            `json:"splash"`
	AfkTimeout        int               `json:"afk_timeout"`
	VerificationLevel VerificationLevel `json:"verification_level"`
	EmbedEnabled      bool              `json:"embed_enabled"`
	Large             bool              `json:"large"` // ??
	Roles             []*Role           `json:"roles"`
	Emojis            []*Emoji          `json:"emojis"`
	Members           []*Member         `json:"members"`
	Presences         []*Presence       `json:"presences"`
	Channels          []*Channel        `json:"channels"`
	VoiceStates       []*VoiceState     `json:"voice_states"`
	Unavailable       *bool             `json:"unavailable"`
}

// A GuildParams stores all the data needed to update discord guild settings
type GuildParams struct {
	Name              string             `json:"name"`
	Region            string             `json:"region"`
	VerificationLevel *VerificationLevel `json:"verification_level"`
}

// A Role stores information about Discord guild member roles.
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Managed     bool   `json:"managed"`
	Hoist       bool   `json:"hoist"`
	Color       int    `json:"color"`
	Position    int    `json:"position"`
	Permissions int    `json:"permissions"`
}

// A VoiceState stores the voice states of Guilds
type VoiceState struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id"`
	Suppress  bool   `json:"suppress"`
	SelfMute  bool   `json:"self_mute"`
	SelfDeaf  bool   `json:"self_deaf"`
	Mute      bool   `json:"mute"`
	Deaf      bool   `json:"deaf"`
}

// A Presence stores the online, offline, or idle and game status of Guild members.
type Presence struct {
	User   *User  `json:"user"`
	Status string `json:"status"`
	Game   *Game  `json:"game"`
}

// PresencesReplace is an array of Presences for an event.
type PresencesReplace []*Presence

func (p *PresencesReplace) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, p)
}

// A Game struct holds the name of the "playing .." game for a user
type Game struct {
	Name string `json:"name"`
}

// A Member stores user information for Guild members.
type Member struct {
	GuildID  string   `json:"guild_id"`
	JoinedAt string   `json:"joined_at"`
	Deaf     bool     `json:"deaf"`
	Mute     bool     `json:"mute"`
	User     *User    `json:"user"`
	Roles    []string `json:"roles"`
}

// A User stores all data for an individual Discord user.
type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Avatar        string `json:"Avatar"`
	Discriminator string `json:"discriminator"`
	Token         string `json:"token"`
	Verified      bool   `json:"verified"`
	Bot           bool   `json:"bot"`
}

// A Settings stores data for a specific users Discord client settings.
type Settings struct {
	RenderEmbeds          bool     `json:"render_embeds"`
	InlineEmbedMedia      bool     `json:"inline_embed_media"`
	EnableTtsCommand      bool     `json:"enable_tts_command"`
	MessageDisplayCompact bool     `json:"message_display_compact"`
	ShowCurrentGame       bool     `json:"show_current_game"`
	Locale                string   `json:"locale"`
	Theme                 string   `json:"theme"`
	MutedChannels         []string `json:"muted_channels"`
}

// An Event provides a basic initial struct for all websocket event.
type Event struct {
	Type      string          `json:"t"`
	State     int             `json:"s"`
	Operation int             `json:"op"`
	Direction int             `json:"dir"`
	RawData   json.RawMessage `json:"d"`
	Struct    interface{}     `json:"-"`
}

// A Ready stores all data for the websocket READY event.
type Ready struct {
	Version           int          `json:"v"`
	SessionID         string       `json:"session_id"`
	HeartbeatInterval uint         `json:"heartbeat_interval"`
	User              *User        `json:"user"`
	ReadState         []*ReadState `json:"read_state"`
	PrivateChannels   []*Channel   `json:"private_channels"`
	Guilds            []*Guild     `json:"guilds"`
}

// A RateLimit struct holds information related to a specific rate limit.
type RateLimit struct {
	Bucket     string        `json:"bucket"`
	Message    string        `json:"message"`
	RetryAfter time.Duration `json:"retry_after"`
}

// A ReadState stores data on the read state of channels.
type ReadState struct {
	MentionCount  int    `json:"mention_count"`
	LastMessageID string `json:"last_message_id"`
	ID            string `json:"id"`
}

// A TypingStart stores data for the typing start websocket event.
type TypingStart struct {
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	Timestamp int    `json:"timestamp"`
}

// A PresenceUpdate stores data for the presence update websocket event.
type PresenceUpdate struct {
	Status  string   `json:"status"`
	GuildID string   `json:"guild_id"`
	Roles   []string `json:"roles"`
	User    *User    `json:"user"`
	Game    *Game    `json:"game"`
}

// A MessageAck stores data for the message ack websocket event.
type MessageAck struct {
	MessageID string `json:"message_id"`
	ChannelID string `json:"channel_id"`
}

// A GuildIntegrationsUpdate stores data for the guild integrations update
// websocket event.
type GuildIntegrationsUpdate struct {
	GuildID string `json:"guild_id"`
}

// A GuildRole stores data for guild role websocket events.
type GuildRole struct {
	Role    *Role  `json:"role"`
	GuildID string `json:"guild_id"`
}

// A GuildRoleDelete stores data for the guild role delete websocket event.
type GuildRoleDelete struct {
	RoleID  string `json:"role_id"`
	GuildID string `json:"guild_id"`
}

// A GuildBan stores data for a guild ban.
type GuildBan struct {
	User    *User  `json:"user"`
	GuildID string `json:"guild_id"`
}

// A GuildEmojisUpdate stores data for a guild emoji update event.
type GuildEmojisUpdate struct {
	GuildID string   `json:"guild_id"`
	Emojis  []*Emoji `json:"emojis"`
}

// A UserGuildSettingsChannelOverride stores data for a channel override for a users guild settings.
type UserGuildSettingsChannelOverride struct {
	Muted                bool   `json:"muted"`
	MessageNotifications int    `json:"message_notifications"`
	ChannelID            string `json:"channel_id"`
}

// A UserGuildSettings stores data for a users guild settings.
type UserGuildSettings struct {
	SupressEveryone      bool                                `json:"suppress_everyone"`
	Muted                bool                                `json:"muted"`
	MobilePush           bool                                `json:"mobile_push"`
	MessageNotifications int                                 `json:"message_notifications"`
	GuildID              string                              `json:"guild_id"`
	ChannelOverrides     []*UserGuildSettingsChannelOverride `json:"channel_overrides"`
}

// UserSettingsUpdate is a map for an event.
type UserSettingsUpdate map[string]interface{}

func (u *UserSettingsUpdate) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, u)
}

// A Message stores all data related to a specific Discord message.
type Message struct {
	ID              string        `json:"id"`
	ChannelID       string        `json:"channel_id"`
	Content         string        `json:"content"`
	Timestamp       string        `json:"timestamp"`
	EditedTimestamp string        `json:"edited_timestamp"`
	Tts             bool          `json:"tts"`
	MentionEveryone bool          `json:"mention_everyone"`
	Author          *User         `json:"author"`
	Attachments     []*Attachment `json:"attachments"`
	Embeds          []*Embed      `json:"embeds"`
	Mentions        []*User       `json:"mentions"`
}

// An Attachment stores data for message attachments.
type Attachment struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url"`
	Filename string `json:"filename"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Size     int    `json:"size"`
}

// An Embed stores data for message embeds.
type Embed struct {
	URL         string `json:"url"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   *struct {
		URL      string `json:"url"`
		ProxyURL string `json:"proxy_url"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
	} `json:"thumbnail"`
	Provider *struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"provider"`
	Author *struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"author"`
	Video *struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"video"`
}

// A VoiceServerUpdate stores the data received during the Voice Server Update
// data websocket event. This data is used during the initial Voice Channel
// join handshaking.
type VoiceServerUpdate struct {
	Token    string `json:"token"`
	GuildID  string `json:"guild_id"`
	Endpoint string `json:"endpoint"`
}

// Resume can be sent over the websocket to continue an existing session.
type Resume struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Sequence  uint64 `json:"seq"`
}

// Resumed is received after a successful Resume packet is sent.
type Resumed struct {
	HeartbeatInterval uint `json:"heartbeat_interval"`
}

// Handshake is sent initially on the first connection to the server.
type Handshake struct {
	Token          string              `json:"token"`
	Properties     HandshakeProperties `json:"properties"`
	Compress       bool                `json:"compress"`
	LargeThreshold int                 `json:"large_threshold"`
}

// HandhsakeProperties are contained within the handshake and describe the
// device conntecting to Discord's server.
type HandshakeProperties struct {
	OS              string `json:"$os"`
	Browser         string `json:"$browser"`
	Device          string `json:"$device"`
	Referer         string `json:"$referer"`
	ReferringDomain string `json:"$referring_domain"`
}
