package cmdhelp

import (
	"github.com/creachadair/twig/command"
)

var Command = &command.C{
	Name:  "help",
	Usage: "[topic/command]",
	Help:  `Print help for the specified command or topic.`,

	CustomFlags: true,
	Run:         command.RunHelp,
}

var Topics = []*command.C{
	{
		Name: "expansions",
		Help: `
Expansions available for query results.

By default, result objects refer to other objects by reference (ID).
The server will expand certain references on request.
Use the syntax "@name" on the command-line, e.g. "@author_id".

author_id                      : expand a user object for the author of a tweet
referenced_tweets.id           : expand referenced tweets (retweets, replies, quotes)
in_reply_to_user_id            : expand a user object for the author of a replied-to tweet
attachments.media_keys         : expand media objects (videos, images) referenced in a tweet
attachments.poll_ids           : expand poll objects defined in a tweet
geo.place_id                   : expand location data tagged in a tweet
entities.mentions.username     : expand user objects for users mentioned in a tweet
referenced_tweets.id.author_id : expand user objects for the authors of referenced tweets
pinned_tweet_id                : expand a tweet object for the pinned ID in a user profile
`,
	},
	{
		Name: "tweet.fields",
	},
	{
		Name: "user.fields",
	},
}
