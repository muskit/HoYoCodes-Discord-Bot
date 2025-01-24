package bot

import (
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
)

// <@%s> = user
// <@!%s> = user (nickname)
// <#%s> = channel
// <@&%s> = role

var (
	//go:embed help_texts/intro.md
	helpIntro string
	//go:embed help_texts/subscriptions.md
	helpSubscriptions string
	//go:embed help_texts/tickers.md
	helpTickers string

	helpTexts = map[string]string {
		"intro": helpIntro,
		"subscriptions": helpSubscriptions,
		"tickers": helpTickers,
	}

	helpChoices = []*discordgo.ApplicationCommandOptionChoice {
		{
			Name: "intro",
			Value: "intro",
		},
		{
			Name: "subscriptions",
			Value: "subscriptions",
		},
		{
			Name: "tickers",
			Value: "tickers",
		},
	}

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

	integrations = []discordgo.ApplicationIntegrationType {
		discordgo.ApplicationIntegrationUserInstall,
		discordgo.ApplicationIntegrationGuildInstall,
	}

	adminCmdFlag int64 = discordgo.PermissionAdministrator

	commands = []*discordgo.ApplicationCommand {
		{
			Name: "help",
			Description: "Get help on using the bot.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "page",
					Description: "What aspect of the bot to get help on. Default: intro",
					Type: discordgo.ApplicationCommandOptionString,
					Required: false,
					Choices: helpChoices,
				},
			},
			IntegrationTypes: &integrations,
		},
		/// SUBSCRIPTIONS ///
		{
			Name: "subscribe",
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
			},
		},
		{
			Name: "unsubscribe",
			Description: "Unsubscribe a channel from all code announcements. Will leave subscription settings alone.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
			},
		},
		{
			Name: "add_ping_role",
			Description: "Adds a role that will be pinged.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "role",
					Description: "Role to ping.",
					Type: discordgo.ApplicationCommandOptionRole,
					Required: true,
				},
			},
		},
		{
			Name: "remove_ping_role",
			Description: "Remove a role from being pinged.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "role",
					Description: "Role to remove from being pinged.",
					Type: discordgo.ApplicationCommandOptionRole,
					Required: true,
				},
			},
		},
		{
			Name: "check_subscription",
			Description: "Show subscription configuration for the current channel.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "all_channels",
					Description: "Show subscriptions for all channels in this server. Default: false",
					Type: discordgo.ApplicationCommandOptionBoolean,
					Required: false,
				},
			},
		},
		/// TICKERS ///
		{
			Name: "create_ticker",
			Description: "Create an ticker that self-updates with active codes. Shows all games if none are specified.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "game",
					Description: "Game to create ticker for.",
					Type: discordgo.ApplicationCommandOptionString,
					Choices: GameChoices,
					Required: true,
				},
			},
		},
		{
			Name: "delete_ticker",
			Description: "Delete a self-updating ticker.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "message_link",
					Description: "Link to message.",
					Type: discordgo.ApplicationCommandOptionString,
					Required: true,
				},
			},
		},
		{
			Name: "check_tickers",
			Description: "Show all tickers present in the server.",
		},
		/// MISC ///
		{
			Name: "active_codes",
			Description: "Privately get the current active codes for a game.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "game",
					Description: "A game to check codes for.",
					Type: discordgo.ApplicationCommandOptionString,
					Choices: GameChoices,
					Required: true,
				},
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

func interactionAuthor(i *discordgo.Interaction) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
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

func handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	page := "intro"
	if ch, exists := opts["page"]; exists {
		page = ch.StringValue()
	}
	text := helpTexts[page]
	slog.Debug("pulled help page:\n", "page", page, "text", text)
	RespondPrivate(s, i, helpTexts[page])
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
	// Bot Added to Server
	// FIXME: event fires when bot goes online
	// session.AddHandler(func (s *discordgo.Session, gc *discordgo.GuildCreate) {
	// 	sysChan := gc.SystemChannelID
	// 	if sysChan != "" {
	// 		// send to server's system msgs
	// 		s.ChannelMessageSend(sysChan, helpIntro)
	// 	} else {
	// 		// TODO: send dm to admins
	// 	}
	// })

	// Bot Interaction
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		slog.Debug(fmt.Sprintf("Received interaction of type %v", i.Type))
		// only commands
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		// "cast" InteractionData to ApplicationCommandInteractionData
		data := i.ApplicationCommandData()
		opts := parseArgs(data.Options)

		slog.Debug(fmt.Sprintf("%s ran %s\n", interactionAuthor(i.Interaction), data.Name))
		if len(opts) > 0 {
			slog.Debug("Command options:")
			for name, val := range opts {
				slog.Debug(fmt.Sprintf("%s=%v\n", name, val))
			}
		}

		// Command matching
		switch data.Name {
		case "help":
			handleHelp(s, i, opts)
		case "subscribe":
			HandleSubscribe(s, i, opts)
		case "unsubscribe":
			HandleUnsubscribe(s, i, opts)
		case "filter_games":
			HandleFilterGames(s, i, opts)
		case "check_subscription":
			HandleCheckSubscription(s, i, opts)
		case "add_ping_role":
			HandleAddPingRole(s, i, opts)
		case "remove_ping_role":
			HandleRemovePingRole(s, i, opts)
		case "create_ticker":
			HandleCreateTicker(s, i, opts)
		case "delete_ticker":
			HandleDeleteTicker(s, i, opts)
		case "active_codes":
			HandleActiveCodes(s, i, opts)
		case "check_tickers":
			HandleGetTickers(s, i)
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

	// Start Discord bot session with handlers set //
	err = session.Open()
	if err != nil {
		log.Fatalf("Could not open Discord session: %s", err)
	}

	session.UpdateCustomStatus("Use `/help` to learn about how I work!")

	// wait for interrupt
	intrpChan := make(chan os.Signal, 1)
	signal.Notify(intrpChan, os.Interrupt)
	go UpdateRoutine(session, intrpChan)
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

	db.Close()
}
