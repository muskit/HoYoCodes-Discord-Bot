package main

import (
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var gameChoices = []*discordgo.ApplicationCommandOptionChoice {
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

var optionalGameChoices = []*discordgo.ApplicationCommandOption {
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

var commands = []*discordgo.ApplicationCommand {
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
		Description: "Subscribe a channel to automatically announce code activity changes.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name: "channel",
				Description: "The channel to auto-announce codes to.",
				Type: discordgo.ApplicationCommandOptionChannel,
				Required: true,
			},
			{
				Name: "game",
				Description: "Game to subscribe for codes to. Don't specify to announce for all games.",
				Type: discordgo.ApplicationCommandOptionString,
				Choices: gameChoices,
				Required: false,
			},
			{
				Name: "ping_role",
				Description: "Role to ping when codes (for game if specified) have been updated.",
				Type: discordgo.ApplicationCommandOptionRole,
				Required: false,
			},
			{
				Name: "announce_code_removals",
				Description: "Determines if bot should announce codes being removed. Default: false",
				Type: discordgo.ApplicationCommandOptionBoolean,
				Required: false,
			},
			{
				Name: "announce_code_additions",
				Description: "Determines if bot should announce codes being added. Default: true",
				Type: discordgo.ApplicationCommandOptionBoolean,
				Required: false,
			},
		},
	},
	{
		Name: "unsubscribe_channel",
		Description: "Unsubscribe a channel from all code announcements.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name: "channel",
				Description: "The channel to unsubscribe all announcement codes from.",
				Type: discordgo.ApplicationCommandOptionChannel,
				Required: true,
			},
		},
	},
	{
		Name: "create_embed",
		Description: "Create an embed that automatically updates with active codes. Shows all games if none are specified.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name: "channel",
				Description: "The channel in which to create the embed.",
				Type: discordgo.ApplicationCommandOptionChannel,
				Required: true,
			},
			optionalGameChoices[0],
			optionalGameChoices[1],
			optionalGameChoices[2],
			optionalGameChoices[3],
		},
	},
	/// ON-DEMAND RUN COMMANDS ///
	{
		Name: "active_codes",
		Description: "Check the current active codes for MiHoYo games. Shows all games if none are specified.",
		Options: optionalGameChoices,
	},
}

// Command arguments typedef
type CMDArgsMap = map[string]*discordgo.ApplicationCommandInteractionDataOption
func parseArgs(options []*discordgo.ApplicationCommandInteractionDataOption) (om CMDArgsMap) {
	om = make(CMDArgsMap)
	for _, opt := range options {
		log.Printf("%s = %s\n", opt.Name, opt)
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
		// Type: discordgo.InteractionResponseModal,
		// Data: &discordgo.InteractionResponseData{
		// 	Title: "echo",
		// 	CustomID: "fuck",
		// 	Content: builder.String(),
		// 	Components: []discordgo.MessageComponent {
		// 		discordgo.ActionsRow{
		// 			Components: []discordgo.MessageComponent {
		// 				discordgo.TextInput{
		// 					CustomID: "some_input",
		// 					Style: discordgo.TextInputShort,
		// 					Label: "Message",
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	})

	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}

func RunBot() {
	log.Println("Starting bot...")

	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("could not load .env: %s", err)
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
			break
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
