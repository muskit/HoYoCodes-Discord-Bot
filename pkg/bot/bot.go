package bot

import (
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	unimplementedResponse discordgo.InteractionResponse = discordgo.InteractionResponse {
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "command unimplemented",
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}

	gameChoices = []*discordgo.ApplicationCommandOptionChoice {
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
			Choices: gameChoices,
			Required: false,
		},
		{
			Name: "game_2",
			Description: "A game to check codes for.",
			Type: discordgo.ApplicationCommandOptionString,
			Choices: gameChoices,
			Required: false,
		},
		{
			Name: "game_3",
			Description: "A game to check codes for.",
			Type: discordgo.ApplicationCommandOptionString,
			Choices: gameChoices,
			Required: false,
		},
		{
			Name: "game_4",
			Description: "A game to check codes for.",
			Type: discordgo.ApplicationCommandOptionString,
			Choices: gameChoices,
			Required: false,
		},
	}

	adminCmdFlag int64 = discordgo.PermissionAdministrator

	commands = []*discordgo.ApplicationCommand {
		/// TEST ///
		{
			Name:        "echo",
			Description: "Say something through the bot",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "message",
					Description: "Contents of the message",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "author",
					Description: "Whether to prepend message's author",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
			},
		},
		/// CHANNEL CONFIGURATION ///
		{
			Name: "subscribe_channel",
			Description: "Subscribe this channel to code activity news. Tracks all games by default; use /filter_games to set.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "channel",
					Description: "Channel to create a subcription for. Default: the current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
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
				{
					Name: "channel",
					Description: "Channel to configure subscribed games for. Default: current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
				optionalGameChoices[0],
				optionalGameChoices[1],
				optionalGameChoices[2],
				optionalGameChoices[3],
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
			Name: "create_embed",
			Description: "Create an embed that updates with active codes. Shows all MiHoYo games if none are specified.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "channel",
					Description: "Channel to create the embed. Default: current channel.",
					Type: discordgo.ApplicationCommandOptionChannel,
					Required: false,
				},
				optionalGameChoices[0],
				optionalGameChoices[1],
				optionalGameChoices[2],
				optionalGameChoices[3],
			},
		},
		/// MISC ///
		{
			Name: "active_codes",
			Description: "Check the current active codes for MiHoYo games. Shows all games if none are specified.",
			Options: optionalGameChoices,
		},
	}
)

// Command arguments typedef
type CMDArgsMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

func parseArgs(options []*discordgo.ApplicationCommandInteractionDataOption) (om CMDArgsMap) {
	om = make(CMDArgsMap)
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

// EXAMPLE: echo cmd handler
func handleEcho(s *discordgo.Session, i *discordgo.InteractionCreate, opts CMDArgsMap) {
	builder := new(strings.Builder)

	author := interactionAuthor(i.Interaction)
	builder.WriteString("**" + author.String() + "** says: ")
	builder.WriteString(opts["message"].StringValue())

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}


func RunBot() {
	log.Println("Starting bot...")
	// read env
	err := godotenv.Load()
	if err != nil {
		log.Printf("WARNING: could not load .env: %v", err)
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
		log.Println("Successfully registered commands!")
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
		log.Printf("%s ran %s\n", interactionAuthor(i.Interaction), data.Name)

		options := parseArgs(data.Options)
		if len(options) > 0 {
			log.Println("Command options:")
			for name, val := range options {
				log.Printf("%s=%v\n", name, val)
			}
		}

		// Command matching
		switch data.Name {
		case "echo":
			handleEcho(s, i, options)
		case "subscribe_channel":
			HandleSubscribe(s, i, options)
		case "unsubscribe_channel":
			HandleUnsubscribe(s, i, options)
		case "active_codes":
			s.InteractionRespond(i.Interaction, &unimplementedResponse)
		}

	})

	// Bot ready
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})


	// Run with callbacks configured! //
	err = session.Open()
	if err != nil {
		log.Fatalf("could not open session: %s", err)
	}

	// wait for interrupt
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	// close session gracefully
	err = session.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %v", err)
	}
}
