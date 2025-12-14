package static

import "embed"

//go:embed *
var StaticFiles embed.FS

// SQLFiles is the ordered list of SQL files to execute for database setup.
// Order matters: settings -> tables -> functions -> procedures -> events -> triggers
// Tables must be in dependency order (e.g., DECK before CARD).
var SQLFiles = []string{
	// database
	"sql/settings.sql",

	// tables
	"sql/tables/USER.sql",
	"sql/tables/DECK.sql",
	"sql/tables/CARD.sql",
	"sql/tables/LOBBY.sql",
	"sql/tables/DRAW_PILE.sql",
	"sql/tables/PLAYER.sql",
	"sql/tables/JUDGE.sql",
	"sql/tables/HAND.sql",
	"sql/tables/RESPONSE.sql",
	"sql/tables/RESPONSE_CARD.sql",
	"sql/tables/REVIEW_CARD.sql",
	"sql/tables/WIN.sql",
	"sql/tables/KICK.sql",
	"sql/tables/USER_ACCESS_DECK.sql",
	"sql/tables/USER_ACCESS_LOBBY.sql",
	"sql/tables/LOGIN_ATTEMPT.sql",
	"sql/tables/LOG_CREDITS_SPENT.sql",
	"sql/tables/LOG_DISCARD.sql",
	"sql/tables/LOG_SKIP.sql",
	"sql/tables/LOG_RESPONSE_CARD.sql",
	"sql/tables/LOG_WIN.sql",
	"sql/tables/LOG_KICK.sql",
	"sql/tables/AUDIT_CARD.sql",
	"sql/tables/AUDIT_DECK.sql",
	"sql/tables/AUDIT_USER.sql",

	// functions
	"sql/functions/FN_GET_DRAW_PILE_CARD_ID.sql",
	"sql/functions/FN_GET_LOBBY_JUDGE_PLAYER_ID.sql",
	"sql/functions/FN_GET_LOGIN_ATTEMPT_IS_ALLOWED.sql",
	"sql/functions/FN_GET_PLAYER_IS_WINNING.sql",
	"sql/functions/FN_USER_HAS_DECK_ACCESS.sql",
	"sql/functions/FN_USER_HAS_LOBBY_ACCESS.sql",

	// procedures
	"sql/procedures/SP_ADD_EXTRA_RESPONSE.sql",
	"sql/procedures/SP_ADD_EXTRA_RESPONSE_UNDO.sql",
	"sql/procedures/SP_ALERT_LOBBY.sql",
	"sql/procedures/SP_BET_ON_WIN.sql",
	"sql/procedures/SP_BET_ON_WIN_UNDO.sql",
	"sql/procedures/SP_BLOCK_RESPONSE.sql",
	"sql/procedures/SP_DISCARD_CARD.sql",
	"sql/procedures/SP_DRAW_HAND.sql",
	"sql/procedures/SP_GAMBLE_CREDITS.sql",
	"sql/procedures/SP_GET_READABLE_DECKS.sql",
	"sql/procedures/SP_PICK_RANDOM_WINNER.sql",
	"sql/procedures/SP_PICK_WINNER.sql",
	"sql/procedures/SP_PURCHASE_CREDITS.sql",
	"sql/procedures/SP_RESET_RESPONSES.sql",
	"sql/procedures/SP_RESPOND_WITH_CARD.sql",
	"sql/procedures/SP_RESPOND_WITH_FIND_CARD.sql",
	"sql/procedures/SP_RESPOND_WITH_STEAL_CARD.sql",
	"sql/procedures/SP_RESPOND_WITH_SURPRISE_CARD.sql",
	"sql/procedures/SP_RESPOND_WITH_WILD_CARD.sql",
	"sql/procedures/SP_SET_LOSING_STREAK.sql",
	"sql/procedures/SP_SET_MISSING_JUDGE_CARD.sql",
	"sql/procedures/SP_SET_MISSING_JUDGE_PLAYER.sql",
	"sql/procedures/SP_SET_NEXT_JUDGE_CARD.sql",
	"sql/procedures/SP_SET_NEXT_JUDGE_PLAYER.sql",
	"sql/procedures/SP_SET_PLAYER_ACTIVE.sql",
	"sql/procedures/SP_SET_PLAYER_INACTIVE.sql",
	"sql/procedures/SP_SET_RESPONSE_COUNT.sql",
	"sql/procedures/SP_SET_RESPONSES_LOBBY.sql",
	"sql/procedures/SP_SET_RESPONSES_PLAYER.sql",
	"sql/procedures/SP_SET_WINNING_STREAK.sql",
	"sql/procedures/SP_SKIP_JUDGE.sql",
	"sql/procedures/SP_SKIP_PROMPT.sql",
	"sql/procedures/SP_START_NEW_ROUND.sql",
	"sql/procedures/SP_VOTE_TO_KICK.sql",
	"sql/procedures/SP_VOTE_TO_KICK_UNDO.sql",
	"sql/procedures/SP_WITHDRAW_CARD.sql",
	"sql/procedures/SP_WITHDRAW_LOBBY.sql",
	"sql/procedures/SP_WITHDRAW_RESPONSE.sql",

	// events
	"sql/events/EVT_CLEAN_BAD_PROMPT_CARDS.sql",
	"sql/events/EVT_CLEAN_BAD_RESPONSE_CARDS.sql",
	"sql/events/EVT_CLEAN_LOGIN_ATTEMPTS.sql",

	// triggers
	"sql/triggers/TR_AUDIT_CARD_DELETE.sql",
	"sql/triggers/TR_AUDIT_CARD_UPDATE.sql",
	"sql/triggers/TR_AUDIT_DECK_DELETE.sql",
	"sql/triggers/TR_AUDIT_DECK_UPDATE.sql",
	"sql/triggers/TR_AUDIT_USER_DELETE.sql",
	"sql/triggers/TR_AUDIT_USER_UPDATE.sql",
	"sql/triggers/TR_LOBBY_AFTER_DELETE.sql",
	"sql/triggers/TR_LOBBY_AFTER_INSERT.sql",
	"sql/triggers/TR_LOBBY_AFTER_UPDATE.sql",
	"sql/triggers/TR_PLAYER_AFTER_INSERT.sql",
	"sql/triggers/TR_PLAYER_AFTER_UPDATE.sql",
	"sql/triggers/TR_PLAYER_BEFORE_INSERT.sql",
	"sql/triggers/TR_REVOKE_ACCESS_AF_UP_DECK.sql",
	"sql/triggers/TR_REVOKE_ACCESS_AF_UP_LOBBY.sql",
	"sql/triggers/TR_SET_CHANGED_ON_DATE_BF_UP_CARD.sql",
	"sql/triggers/TR_SET_CHANGED_ON_DATE_BF_UP_DECK.sql",
	"sql/triggers/TR_SET_CHANGED_ON_DATE_BF_UP_USER.sql",
}
