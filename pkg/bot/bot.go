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
		/// CONFIGURATION COMMANDS ///
		{
			Name: "subscribe_channel",
			Description: "Subscribe this channel to code activity news. Tracks all games by default; use `/filter_games` to set.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name: "ping_role",
					Description: "Role to ping when codes (for game if specified) have been updated.",
					Type: discordgo.ApplicationCommandOptionRole,
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
			Description: "Set games this channel should be subscribed to. Run command without games to subscribe to all.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: optionalGameChoices,
		},
		{
			Name: "unsubscribe_channel",
			Description: "Unsubscribe a channel from all code announcements.",
			DefaultMemberPermissions: &adminCmdFlag,
		},
		{
			Name: "create_embed",
			Description: "Create an embed that updates with active codes. Shows all MiHoYo games if none are specified.",
			DefaultMemberPermissions: &adminCmdFlag,
			Options: optionalGameChoices,
		},
		/// ON-DEMAND RUN COMMANDS ///
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
		log.Printf("%s = %v\n", opt.Name, opt)
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
	_, err = session.ApplicationCommandBulkOverwrite(appId, "", commands)
	if err != nil {
		log.Fatalf("could not register commands: %s", err)
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

		// Command matching
		switch data.Name {
		case "echo":
			handleEcho(s, i, options)
		case "active_codes":
			
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
		log.Printf("could not close session gracefully: %s", err)
	}
}
