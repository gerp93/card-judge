# Card Judge

Version: 1.29.0

Card Judge is an open source online multiplayer party game.

Built on [gameshell-framework](https://github.com/gerp93/gameshell-framework)
(see `src/go.mod` for the pinned version) — the shared user/lobby/realtime
platform this and other gameshell-framework games run on. Card Judge was the
first game built on it, and the framework was later extracted out of this repo.

## Definitions

- A **card** is text categorized as either a *Prompt* or *Response*.
- A **deck** is a grouping of cards.
- A **lobby** is an active game taking place.
- A **draw pile** is the available cards within a lobby.
- A **user** is an account created on the site.
- A **player** is a user in a lobby.
- A **judge** is a role given to a single player within a lobby.

## Rules

### Start Playing

To start playing, create/join a lobby of at least three players
(including yourself). The player who created the lobby will start as the
judge.

### Playing a Round

The prompt card will be displayed in the middle of the game board for
all players to read.

All players (except the judge) answer the question or fill in the blanks
of the prompt card by selecting the amount of required response cards
from their hand. Players can withdraw their play as long not all other
players have played.

The judge has the option to skip the current prompt card. Any response
cards already played will be returned to the players hand.

Once all players have played the required amount of cards, the answers
will be put in alphabetical order to ensure randomness. The judge will
be able to reveal each response one at a time. The judge can rule-out
any responses they do not like.

Once all cards are revealed the judge can select the winner. The judge
should choose the response that they believe to be the best played
response for the given prompt card. A random option is provided if the
judge cannot decide between the remaining non-ruled-out responses.

The judge choosing a winner concludes that round. The player who won is
awarded a point on the scoreboard. The judge rotates based on the order
players joined the lobby.

Gameplay can continue until there are no more cards to draw from the
draw pile or when players agree to finish. The player with the most
points is the winner.

### Credits/Specials/Perks

A lobby will have a set free credits for each player. Credits can be
used to purchase Specials/Perks at varying costs. Specials vary from
gambling credits to potentially earn more, to playing Special Cards
which are an alternative way to play a turn. Each Special has a help
icon next to it to display what it is. Perks last the rest of the game,
as opposed to the round. Perks will provide a long term advantage in the
game.

Credits can be earned/lost through win/lose streaks. The streak amount
can be set per lobby. This is to help balance and allow those who may be
winning more to lose credits, and allow those losing more to gain
credits and catch up.

A handicap is in place to help balance the game. A player's handicap is
the amount of other players in the lobby they are beating. All fixed
cost specials/perks will have the handicap added to the price.

## Deployment

Create/restore and backup/delete of Card Judge instances on Digital Ocean is
handled by the shared [gameshell-deploy](https://github.com/gerp93/gameshell-deploy)
control plane — this repo has no deploy scripts, `deploy.conf`, or `backups/`
of its own; that config and data live in gameshell-deploy under
`games/card-judge/`. See gameshell-deploy's README for prerequisites
(Digital Ocean account, `doctl`, GPG) and `create.sh`/`delete.sh` usage.

## Environment Variables

The following environment variables are needed to run your own instance:

```
// MySQL/MariaDB
CARD_JUDGE_SQL_HOST // ip address of server
CARD_JUDGE_SQL_DATABASE // database name
CARD_JUDGE_SQL_USER // database username
CARD_JUDGE_SQL_PASSWORD // database username password

// Port
CARD_JUDGE_PORT // [optional] port to serve (defaults to 2016)

// Redirect Logs
CARD_JUDGE_LOG_FILE // [optional] path to log file (defaults to stdout)

// HTTPS Certificates
CARD_JUDGE_CERT_FILE // [optional] path to cert file
CARD_JUDGE_KEY_FILE // [optional] path to key file
```
