package cmdhelp

import (
	"github.com/creachadair/command"
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
List of expansions available for query results.

By default, result objects refer to other objects by reference (ID).
The server will expand certain references on request.
Use the syntax "@name" on the command-line, e.g. "@author_id".

author_id
   : expand a user object for the author of a tweet

referenced_tweets.id           (shortcut: tweets, ref_tweets)
   : expand referenced tweets (retweets, replies, quotes)

in_reply_to_user_id            (shortcut: reply_to_user)
   : expand a user object for the author of a replied-to tweet

attachments.media_keys         (shortcut: media_keys)
   : expand media objects (videos, images) referenced in a tweet

attachments.poll_ids           (shortcut: poll_ids)
   : expand poll objects defined in a tweet

geo.place_id                   (shortcut: place_id)
   : expand location data tagged in a tweet

entities.mentions.username     (shortcut: mentions)
   : expand user objects for users mentioned in a tweet

referenced_tweets.id.author_id (shortcut: ref_author)
   : expand user objects for the authors of referenced tweets

pinned_tweet_id                (shortcut: pinned_tweet)
   : expand a tweet object for the pinned ID in a user profile
`,
	},
	{
		Name: "tweet.fields",
		Help: `
List of optional Tweet field parameters.

By default, tweet objects are returned with a minimal set of default
fields (ID and text). Additional fields can be requested in the query.
Use the syntax "tweet:name" on the command-line, e.g., "tweet:author_id".

The following field tags are available:

  attachments          geo                  promoted_metrics
  author_id            in_reply_to_user_id  public_metrics
  context_annotations  lang                 referenced_tweets
  conversation_id      non_public_metrics   source
  created_at           organic_metrics      withheld
  entities             possibly_sensitive
`,
	},
	{
		Name: "user.fields",
		Help: `
List of optional User field parameters.

By default, user objects are returned with a minimal set of default
fields (ID, name, and username). Additional fields can be requested in
the query. Use the syntax "user:name", e.g., "user:created_at".

The following field tags are available:

  created_at   pinned_tweet_id    url
  description  profile_image_url  verified
  entities     protected          withheld
  location     public_metrics
`,
	},
	{
		Name: "poll.fields",
		Help: `
List of optional Poll field parameters.

By default, poll objects are returned with a minimal set of default
fields (ID and options). Additional fields can be requested in
the query. Use the syntax "poll:name", e.g., "poll:duration_minutes".

The following field tags are available:

  attachments
  duration_minutes
  end_datetime
  voting_status
`,
	},
	{
		Name: "media.fields",
		Help: `
List of optional Media field parameters.

By default, media objects are returned with a minimal set of default
fields (media_key and type). Additional fields can be requested in
the query. Use the syntax "media:name", e.g., "media:width".

The following field tags are available:

  attachments  non_public_metrics  promoted_metrics
  duration_ms  organic_metrics     public_metrics
  height       preview_image_url   width
`,
	},
}
