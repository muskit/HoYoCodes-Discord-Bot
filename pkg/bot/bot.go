package bot

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// <@%s> = user
// <@!%s> = user (nickname)
// <#%s> = channel
// <@&%s> = role

var (
	GameChoices = []*discordgo.ApplicationCommandOptionChoice {
		{
			Name: "Honkai Impact 3rd",
			Value: "Honkai Impact 3rd",
		},
		{
			Name: "Genshin Impact",
			Value: "Genshin Impact",
		},
		{
			Name: "Honkai Star Rail",
			Value: "Honkai Star Rail",
		},
		{
			Name: "Zenless Zone Zero",
			Value: "Zenless Zone Zero",
		},
	}

	optionalGameChoices = []*discordgo.ApplicationCommandOption {
		{
			Name: "game_1",
			Description: "A game to check codes for.",
			Type: discordgo.ApplicationCommandOptionString,
			Choices: GameChoices,
			Required: false,
		},
		{
			Name: "game_2",
			Description: "A game to check codes for.",
			Type: discordgo.ApplicationCommandOptionString,
			Choices: GameChoices,
			Required: false,
		},
		{
			Name: "game_3",
			Description: "A game to check codes for.",
			Type: discordgo.ApplicationCommandOptionString,
			Choices: GameChoices,
			Required: false,
		},
		{
			Name: "game_4",
			Description: "A game to check codes for.",
			Type: discordgo.ApplicationCommandOptionString,
			Choices: GameChoices,
			Required: false,
		},
	}

	adminCmdFlag int64 = discordgo.PermissionAdministrator

	commands = []*discordgo.ApplicationCommand {
		{
			Name: "help",
			Description: "Introduction on configuring the bot.",
			DefaultMemberPermissions: &adminCmdFlag,
		},
		/// CHANNEL CONFIGURATION ///
		{
			Name: "subscribe_channel",
			Description: "Subscribe this channel to code activity news. Tracks all games by default; use /filter_games to set.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "announce_code_additions",
					Description: "Determines if bot should announce codes being added. Default: `true`",
					Type: discordgo.ApplicationCommandOptionBoolean,
					Required: false,
				},
				{
					Name: "announce_code_removals",
					Description: "Determines if bot should announce codes being removed. Default: `false`",
					Type: discordgo.ApplicationCommandOptionBoolean,
					Required: false,
				},
				{
					Name: "channel",
					Description: "Channel to create a subscription for. Default: the current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
			},
		},
		{
			Name: "filter_games",
			Description: "Set games this channel should be subscribed to. Not specifying games will subscribe to all.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				optionalGameChoices[0],
				optionalGameChoices[1],
				optionalGameChoices[2],
				optionalGameChoices[3],
				{
					Name: "channel",
					Description: "Channel to configure subscribed games for. Default: current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
			},
		},
		{
			Name: "unsubscribe_channel",
			Description: "Unsubscribe a channel from all code announcements. Will leave channel configuration alone.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "channel",
					Description: "Channel to unsubscribe. Default: current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
			},
			DefaultMemberPermissions: &adminCmdFlag,
		},
		{
			Name: "add_ping_role",
			Description: "Adds a role that will be pinged.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "role",
					Description: "Role to ping.",
					Type: discordgo.ApplicationCommandOptionRole,
					Required: true,
				},
				{
					Name: "channel",
					Description: "Channel to add a ping role for. Default: current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
			},
		},
		{
			Name: "remove_ping_role",
			Description: "Remove a role from being pinged.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "role",
					Description: "Role to remove from being pinged.",
					Type: discordgo.ApplicationCommandOptionRole,
					Required: true,
				},
				{
					Name: "channel",
					Description: "Channel to remove a ping role from. Default: current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
			},
		},
		{
			Name: "create_embed",
			Description: "Create an embed that self-updates with active codes. Shows all games if none are specified.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "game",
					Description: "Game to create embed for.",
					Type: discordgo.ApplicationCommandOptionString,
					Choices: GameChoices,
					Required: true,
				},
				{
					Name: "channel",
					Description: "Channel to create the embed. Default: current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
			},
		},
		{
			Name: "delete_embed",
			Description: "Delete a self-updating embed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "message_link",
					Description: "Link to message containing the embed.",
					Type: discordgo.ApplicationCommandOptionString,
					Required: true,
				},
			},
		},
		{
			Name: "show_config",
			Description: "Show subscription configuration for a channel.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "all_channels",
					Description: "Whether to show config for all channels in this server or not. Default: false",
					Type: discordgo.ApplicationCommandOptionBoolean,
					Required: false,
				},
				{
					Name: "channel",
					Description: "Channel to show config for. Default: current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
			},
		},
		/// MISC ///
		{
			Name: "active_codes",
			Description: "Check the current active codes for MiHoYo games. Shows all games if none are specified.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "recents_only",
					Description: "Show only codes that have been released recently. Default: true",
					Type: discordgo.ApplicationCommandOptionBoolean,
					Required: false,
				},
				optionalGameChoices[0],
				optionalGameChoices[1],
				optionalGameChoices[2],
				optionalGameChoices[3],
			},
		},
	}
)

// Command arguments typedef
type CmdOptMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

func parseArgs(options []*discordgo.ApplicationCommandInteractionDataOption) (om CmdOptMap) {
	om = make(CmdOptMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return
}

func handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	HELP_TEXT := (
		"**Welcome to HoyoCodes!**\n\n"+
		"This bot will notify you of new goods code releases for MiHoYo games in two different formats:\n"+
		"- Auto-updating status embeds that list all codes reported to be active and usable\n"+
		"  - `/create_embed`, `/delete_embed`\n"+
		"- Channel subscriptions that notify when new codes are added and/or removed\n"+
		"  - `/subscribe`, `/unsubscribe`, `/filter_games`, `/add_ping_role`, `/remove_ping_role`\n\n"+
		"As you're setting up subscriptions and embeds with these commands, use `/show_config` to check your configuration work so far.\n\n"+
		"Feel free to DM me and set up your own personalized notifications!\n\n"+
		"-# Developed by [muskit](https://muskit.net).\n"+
		"-# Code reporting provided by [PocketTactics](<https://www.pockettactics.com>).")
	RespondPrivate(s, i, HELP_TEXT)
}

func GetChannelID(i *discordgo.InteractionCreate, opts CmdOptMap) uint64 {
	id, _ := strconv.ParseUint(i.ChannelID, 10, 64)
	if val, exists := opts["channel"]; exists {
		id, _ = strconv.ParseUint(val.ChannelValue(nil).ID, 10, 64)
	}
	return id
}

func Respond(s *discordgo.Session, i *discordgo.InteractionCreate, str string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: str,
		},
	})
	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}

func RespondPrivate(s *discordgo.Session, i *discordgo.InteractionCreate, str string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: str,
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}

func RunBot() {
	slog.Info("Starting bot...")
	// read env
	err := godotenv.Load()
	if err != nil {
		slog.Warn(fmt.Sprintf("Could not load .env: %v", err))
	}

	// get vars from env
	token := os.Getenv("token")
	appId := os.Getenv("app_id")

	// init bot
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	// register commands
	if _, err = session.ApplicationCommandBulkOverwrite(appId, "", commands); err != nil {
		log.Fatalf("Could not register commands: %s\n", err)
	} else {
		slog.Info("Successfully registered commands!")
	}

	// EVENT HANDLERS //
	// Bot Interaction
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// only commands
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		// "cast" InteractionData to ApplicationCommandInteractionData
		data := i.ApplicationCommandData()
		opts := parseArgs(data.Options)

		// log.Printf("%s ran %s\n", interactionAuthor(i.Interaction), data.Name)
		// if len(opts) > 0 {
		// 	log.Println("Command options:")
		// 	for name, val := range opts {
		// 		log.Printf("%s=%v\n", name, val)
		// 	}
		// }

		// Command matching
		switch data.Name {
		case "help":
			handleHelp(s, i)
		case "subscribe_channel":
			HandleSubscribe(s, i, opts)
		case "unsubscribe_channel":
			HandleUnsubscribe(s, i, opts)
		case "filter_games":
			HandleFilterGames(s, i, opts)
		case "show_config":
			HandleShowConfig(s, i, opts)
		case "add_ping_role":
			HandleAddPingRole(s, i, opts)
		case "remove_ping_role":
			HandleRemovePingRole(s, i, opts)
		case "create_embed":
			HandleCreateEmbed(s, i, opts)
		case "delete_embed":
			HandleDeleteEmbed(s, i, opts)
		default:
			slog.Warn(fmt.Sprintf("Tried to run an unimplemented command %s!!", data.Name))
			if len(opts) > 0 {
				slog.Debug("Command options:")
				for name, val := range opts {
					slog.Debug(fmt.Sprintf("%s=%v", name, val))
				}
			}
			RespondPrivate(s, i, "command unimplemented")
		}

	})

	// Bot ready
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		slog.Info(fmt.Sprintf("Logged in as %s", r.User.String()))
	})

	// Start Discord bot session with handlers set
	err = session.Open()
	if err != nil {
		log.Fatalf("Could not open Discord session: %s", err)
	}

	go UpdateLoop(session, 2*time.Hour)

	// wait for interrupt
	intrpChan := make(chan os.Signal, 1)
	signal.Notify(intrpChan, os.Interrupt)
	<-intrpChan

	if !UpdatingMutex.TryLock() {
	slog.Info("Waiting until current update finishes to close...")
		UpdatingMutex.Lock() // wait until update is over
	}

	slog.Info("Closing Discord session...")
	err = session.Close()
	if err != nil {
		slog.Warn(fmt.Sprintf("Could not close session gracefully: %v", err))
	}
	slog.Info("Discord session closed!")
}
