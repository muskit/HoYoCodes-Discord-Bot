package consts

import "time"

var Games = []string{
	"Honkai Impact 3rd",
	"Genshin Impact",
	"Honkai Star Rail",
	"Zenless Zone Zero",
}

var RedeemURL = map[string]string{
	"Genshin Impact": "https://genshin.hoyoverse.com/en/gift",
	"Honkai Star Rail": "https://hsr.hoyoverse.com/gift",
	"Zenless Zone Zero": "https://zenless.hoyoverse.com/redemption",
}

var ArticleURL map[string]string = map[string]string{
	"Honkai Impact 3rd": "https://www.pockettactics.com/honkai-impact/codes",
	"Genshin Impact": "https://www.pockettactics.com/genshin-impact/codes",
	"Honkai Star Rail": "https://www.pockettactics.com/honkai-star-rail/codes",
	"Zenless Zone Zero": "https://www.pockettactics.com/zenless-zone-zero/codes",
}

const UpdateInterval = 2 * time.Hour
const RecentSinceLatestThreshold = 36 * time.Hour
const RecentThreshold = 7*24*time.Hour

// guild_id, channel_id, message_id
const MessageLinkTemplate = "https://discord.com/channels/%v/%v/%v"