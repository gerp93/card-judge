package database

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"uuid"
)

type StatUser struct {
	UserName string

	GamePlayCount            int
	GameWinCount             int
	RoundPlayCount           int
	RoundWinCount            int
	ResponseCardPlayCount    int
	ResponseCardDiscardCount int
	PromptCardPlayCount      int
	PromptCardSkipCount      int
	CreditsSpentCount        int
	CreditsEarnedCount       int
	LobbyKickCount           int
	FlipTableCount           int
}

type Achievement struct {
	Name     string
	Progress int
}

type StatCard struct {
	DeckName string
	Category string
	Text     string

	PlayCount    int
	WinCount     int
	DiscardCount int
	SkipCount    int
}

func GetStatsLeaderboard(userId uuid.UUID, timeframe string, topic string, subject string) ([]string, [][]string, error) {
	resultHeaders := make([]string, 0)
	resultRows := make([][]string, 0)
	params := make([]any, 0)

	var timeframeDateString string
	switch timeframe {
	case "all":
		timeframeDateString = "1970-01-01"
	case "year":
		timeframeDateString = "DATE_SUB(CURRENT_DATE(), INTERVAL 1 YEAR)"
	case "quarter":
		timeframeDateString = "DATE_SUB(CURRENT_DATE(), INTERVAL 1 QUARTER)"
	case "month":
		timeframeDateString = "DATE_SUB(CURRENT_DATE(), INTERVAL 1 MONTH)"
	case "day":
		timeframeDateString = "DATE_SUB(CURRENT_DATE(), INTERVAL 1 DAY)"
	default:
		return resultHeaders, resultRows, errors.New("invalid time frame provided")
	}

	var sqlString string
	switch topic {
	case "game-win-ratio":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Games Played")
			resultHeaders = append(resultHeaders, "Games Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				WITH GAMES_PLAYED AS (
						SELECT
							PLAYER_USER_ID AS USER_ID,
							COUNT(DISTINCT LOBBY_ID) AS PLAY_COUNT
						FROM LOG_RESPONSE_CARD
						WHERE CREATED_ON_DATE >= %s
						GROUP BY PLAYER_USER_ID
					),
					GAME_WINS AS (
						SELECT
							USER_ID,
							COUNT(DISTINCT LOBBY_ID) AS WIN_COUNT
						FROM V_GAME_WINNER
						WHERE TIMESTAMP >= %s
						GROUP BY USER_ID
					)
				SELECT
					GP.PLAY_COUNT,
					COALESCE(GW.WIN_COUNT, 0) AS WIN_COUNT,
					COALESCE((GW.WIN_COUNT * 1.0) / (GP.PLAY_COUNT * 1.0), 0.0) AS WIN_RATIO,
					U.NAME AS NAME
				FROM USER AS U
					INNER JOIN GAMES_PLAYED AS GP ON GP.USER_ID = U.ID
					LEFT JOIN GAME_WINS AS GW ON GW.USER_ID = U.ID
				ORDER BY WIN_RATIO DESC,
					PLAY_COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "game-win":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Games Won")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				WITH GAME_WINS AS (
						SELECT
							USER_ID,
							COUNT(DISTINCT LOBBY_ID) AS WIN_COUNT
						FROM V_GAME_WINNER
						WHERE TIMESTAMP >= %s
						GROUP BY USER_ID
					)
				SELECT
					GW.WIN_COUNT AS COUNT,
					U.NAME AS NAME
				FROM USER AS U
					INNER JOIN GAME_WINS AS GW ON GW.USER_ID = U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "game-play":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Games Played")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.LOBBY_ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN USER AS U ON U.ID = LRC.PLAYER_USER_ID
				WHERE LRC.CREATED_ON_DATE >= %s
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Games Played")
			resultHeaders = append(resultHeaders, "Card")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.LOBBY_ID) AS COUNT,
					COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
				WHERE LRC.CREATED_ON_DATE >= %s
				GROUP BY C.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "special-category":
			resultHeaders = append(resultHeaders, "Games Played")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.LOBBY_ID) AS COUNT,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
				WHERE LRC.CREATED_ON_DATE >= %s
				GROUP BY LRC.SPECIAL_CATEGORY
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "round-win-ratio":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					PLAY_COUNT,
					WIN_COUNT,
					COALESCE((WIN_COUNT * 1.0) / (PLAY_COUNT * 1.0), 0.0) AS WIN_RATIO,
					NAME
				FROM (
						SELECT
							COUNT(DISTINCT LRC.ROUND_ID) AS PLAY_COUNT,
							COUNT(DISTINCT LW.ID) AS WIN_COUNT,
							U.NAME AS NAME
						FROM LOG_RESPONSE_CARD AS LRC
							LEFT JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
							INNER JOIN USER AS U ON U.ID = LRC.PLAYER_USER_ID
						WHERE LRC.CREATED_ON_DATE >= %s
						GROUP BY U.ID
					) AS T
				ORDER BY WIN_RATIO DESC,
					PLAY_COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					PLAY_COUNT,
					WIN_COUNT,
					COALESCE((WIN_COUNT * 1.0) / (PLAY_COUNT * 1.0), 0.0) AS WIN_RATIO,
					NAME
				FROM (
						SELECT
							COUNT(DISTINCT LRC.ROUND_ID) AS PLAY_COUNT,
							COUNT(DISTINCT LW.ID) AS WIN_COUNT,
							COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
						FROM LOG_RESPONSE_CARD AS LRC
							LEFT JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
							LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
						WHERE LRC.CREATED_ON_DATE >= %s
							AND FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
						GROUP BY C.ID
					) AS T
				ORDER BY WIN_RATIO DESC,
					PLAY_COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "special-category":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = fmt.Sprintf(`
				SELECT
					PLAY_COUNT,
					WIN_COUNT,
					COALESCE((WIN_COUNT * 1.0) / (PLAY_COUNT * 1.0), 0.0) AS WIN_RATIO,
					NAME
				FROM (
						SELECT
							COUNT(DISTINCT LRC.ROUND_ID) AS PLAY_COUNT,
							COUNT(DISTINCT LW.ID) AS WIN_COUNT,
							COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
						FROM LOG_RESPONSE_CARD AS LRC
							LEFT JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
						WHERE LRC.CREATED_ON_DATE >= %s
						GROUP BY LRC.SPECIAL_CATEGORY
					) AS T
				ORDER BY WIN_RATIO DESC,
					PLAY_COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "round-win":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT RW.ROUND_ID) AS COUNT,
					U.NAME AS NAME
				FROM V_ROUND_WINNER AS RW
					INNER JOIN USER AS U ON U.ID = RW.USER_ID
				WHERE RW.TIMESTAMP >= %s
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "special-category":
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
				WHERE LRC.CREATED_ON_DATE >= %s
				GROUP BY LRC.SPECIAL_CATEGORY
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "round-play":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN USER AS U ON U.ID = LRC.PLAYER_USER_ID
				WHERE LRC.CREATED_ON_DATE >= %s
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "special-category":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
				WHERE LRC.CREATED_ON_DATE >= %s
				GROUP BY LRC.SPECIAL_CATEGORY
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "response-card-play":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN USER AS U ON U.ID = LRC.PLAYER_USER_ID
				WHERE LRC.CREATED_ON_DATE >= %s
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ID) AS COUNT,
					COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "special-category":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ID) AS COUNT,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
				WHERE LRC.CREATED_ON_DATE >= %s
				GROUP BY LRC.SPECIAL_CATEGORY
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "response-card-discard":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Cards Discarded")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LD.ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_DISCARD AS LD
					INNER JOIN USER AS U ON U.ID = LD.USER_ID
				WHERE LD.CREATED_ON_DATE >= %s
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Cards Discarded")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LD.ID) AS COUNT,
					COALESCE(C.TEXT, 'Unknown') AS NAME
				FROM LOG_DISCARD AS LD
					LEFT JOIN CARD AS C ON C.ID = LD.CARD_ID
				WHERE LD.CREATED_ON_DATE >= %s
					AND FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "prompt-card-play":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN USER AS U ON U.ID = LRC.JUDGE_USER_ID
				WHERE LRC.CREATED_ON_DATE >= %s
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					COALESCE(C.TEXT, 'Unknown') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					LEFT JOIN CARD AS C ON C.ID = LRC.JUDGE_CARD_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "prompt-card-skip":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Cards Skipped")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LS.ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_SKIP AS LS
					INNER JOIN USER AS U ON U.ID = LS.USER_ID
				WHERE LS.CREATED_ON_DATE >= %s
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Cards Skipped")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(DISTINCT LS.ID) AS COUNT,
					COALESCE(C.TEXT, 'Unknown') AS NAME
				FROM LOG_SKIP AS LS
					LEFT JOIN CARD AS C ON C.ID = LS.CARD_ID
				WHERE LS.CREATED_ON_DATE >= %s
					AND FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "picked-judge":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Judge Picking")
			resultHeaders = append(resultHeaders, "Player")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					UJ.NAME AS JUDGE_NAME,
					UP.NAME AS NAME,
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UJ ON UJ.ID = LRC.JUDGE_USER_ID
					INNER JOIN USER AS UP ON UP.ID = LRC.PLAYER_USER_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND UJ.ID = ?
				GROUP BY LRC.JUDGE_USER_ID,
					LRC.PLAYER_USER_ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Judge Picking")
			resultHeaders = append(resultHeaders, "Card")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					UJ.NAME AS JUDGE_NAME,
					CP.TEXT AS NAME,
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UJ ON UJ.ID = LRC.JUDGE_USER_ID
					INNER JOIN CARD AS CP ON CP.ID = LRC.PLAYER_CARD_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND UJ.ID = ?
				GROUP BY LRC.JUDGE_USER_ID,
					LRC.PLAYER_CARD_ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "special-category":
			resultHeaders = append(resultHeaders, "Judge Picking")
			resultHeaders = append(resultHeaders, "Special Category")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					UJ.NAME AS JUDGE_NAME,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME,
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UJ ON UJ.ID = LRC.JUDGE_USER_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND UJ.ID = ?
				GROUP BY LRC.JUDGE_USER_ID,
					LRC.SPECIAL_CATEGORY
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "picked-player":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Winner")
			resultHeaders = append(resultHeaders, "Judge Who Picked")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					UP.NAME AS PLAYER_NAME,
					UJ.NAME AS NAME,
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UJ ON UJ.ID = LRC.JUDGE_USER_ID
					INNER JOIN USER AS UP ON UP.ID = LRC.PLAYER_USER_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND UP.ID = ?
				GROUP BY LRC.JUDGE_USER_ID,
					LRC.PLAYER_USER_ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "card":
			resultHeaders = append(resultHeaders, "Winner")
			resultHeaders = append(resultHeaders, "Card Played")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					UP.NAME AS PLAYER_NAME,
					CJ.TEXT AS NAME,
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN CARD AS CJ ON CJ.ID = LRC.PLAYER_CARD_ID
					INNER JOIN USER AS UP ON UP.ID = LRC.PLAYER_USER_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND UP.ID = ?
				GROUP BY LRC.PLAYER_CARD_ID,
					LRC.PLAYER_USER_ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		case "special-category":
			resultHeaders = append(resultHeaders, "Winner")
			resultHeaders = append(resultHeaders, "Special Category Played")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = fmt.Sprintf(`
				SELECT
					UP.NAME AS PLAYER_NAME,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME,
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UP ON UP.ID = LRC.PLAYER_USER_ID
				WHERE LRC.CREATED_ON_DATE >= %s
					AND UP.ID = ?
				GROUP BY LRC.SPECIAL_CATEGORY,
					LRC.PLAYER_USER_ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-spent":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Spent")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					SUM(LCS.AMOUNT) AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.AMOUNT > 0
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-earned":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Earned")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					SUM(LCS.AMOUNT) * -1 AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.AMOUNT < 0
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-spent-category":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Spent")
			resultHeaders = append(resultHeaders, "Category")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					SUM(LCS.AMOUNT) AS COUNT,
					LCS.CATEGORY AS CATEGORY,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.AMOUNT > 0
				GROUP BY U.ID,
					LCS.CATEGORY
				ORDER BY COUNT DESC,
					NAME ASC,
					CATEGORY ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-earned-category":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Earned")
			resultHeaders = append(resultHeaders, "Category")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					SUM(LCS.AMOUNT) * -1 AS COUNT,
					LCS.CATEGORY AS CATEGORY,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.AMOUNT < 0
				GROUP BY U.ID,
					LCS.CATEGORY
				ORDER BY COUNT DESC,
					NAME ASC,
					CATEGORY ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-spent-game":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Spent in a Game")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					SUM(LCS.AMOUNT) AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.AMOUNT > 0
				GROUP BY U.ID,
					LCS.LOBBY_ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-earned-game":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Earned in a Game")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					SUM(LCS.AMOUNT) * -1 AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.AMOUNT < 0
				GROUP BY U.ID,
					LCS.LOBBY_ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "gamble":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Gamble")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					MAX(LCS.AMOUNT) AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.CATEGORY = 'GAMBLE'
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "gamble-win":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Gamble Win")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					MIN(LCS.AMOUNT) * -1 AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.CATEGORY = 'GAMBLE-WIN'
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "bet":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Bet")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					MAX(LCS.AMOUNT) AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.CATEGORY = 'BET'
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "bet-win":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Bet Win")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					MIN(LCS.AMOUNT) * -1 AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CREATED_ON_DATE >= %s
					AND LCS.CATEGORY = 'BET-WIN'
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "kick":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Kicked")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(*) AS COUNT,
					U.NAME AS NAME
				FROM LOG_KICK AS LK
					INNER JOIN USER AS U ON U.ID = LK.USER_ID
				WHERE LK.CREATED_ON_DATE >= %s
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "flip-table":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Flipped Tables")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = fmt.Sprintf(`
				SELECT
					COUNT(*) AS COUNT,
					U.NAME AS NAME
				FROM LOG_FLIP_TABLE AS LFT
					INNER JOIN USER AS U ON U.ID = LFT.USER_ID
				WHERE LFT.CREATED_ON_DATE >= %s
				GROUP BY U.ID
				ORDER BY COUNT DESC,
					NAME ASC
				LIMIT 10
			`, timeframeDateString)
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	default:
		return resultHeaders, resultRows, errors.New("invalid topic provided")
	}

	rows, err := query(sqlString, params...)
	if err != nil {
		return resultHeaders, resultRows, err
	}
	defer rows.Close()

	rowValuePointers := make([]any, len(resultHeaders))
	for i := range rowValuePointers {
		rowValuePointers[i] = new(string)
	}

	for rows.Next() {
		if err := rows.Scan(rowValuePointers...); err != nil {
			log.Println(err)
			return resultHeaders, resultRows, errors.New("failed to scan row in query results")
		}

		row := make([]string, len(resultHeaders))
		for i, v := range rowValuePointers {
			row[i] = reflect.ValueOf(v).Elem().String()
		}
		resultRows = append(resultRows, row)
	}

	return resultHeaders, resultRows, nil
}

func GetStatsUser(userId uuid.UUID) (StatUser, error) {
	var result StatUser

	sqlString := `
		SELECT
			U.NAME AS USER_NAME,
			(
				SELECT
					COUNT(DISTINCT LOBBY_ID)
				FROM LOG_RESPONSE_CARD
				WHERE PLAYER_USER_ID = U.ID
			) AS GAME_PLAY_COUNT,
			(
				SELECT
					COUNT(DISTINCT LOBBY_ID) AS WIN_COUNT
				FROM V_GAME_WINNER
				WHERE USER_ID = U.ID
			) AS GAME_WIN_COUNT,
			(
				SELECT
					COUNT(DISTINCT ROUND_ID)
				FROM LOG_RESPONSE_CARD
				WHERE PLAYER_USER_ID = U.ID
			) AS ROUND_PLAY_COUNT,
			(
				SELECT
					COUNT(DISTINCT ROUND_ID)
				FROM V_ROUND_WINNER
				WHERE USER_ID = U.ID
			) AS ROUND_WIN_COUNT,
			(SELECT COUNT(*) FROM LOG_RESPONSE_CARD WHERE PLAYER_USER_ID = U.ID) AS RESPONSE_CARD_PLAY_COUNT,
			(SELECT COUNT(*) FROM LOG_DISCARD WHERE USER_ID = U.ID) AS RESPONSE_CARD_DISCARD_COUNT,
			(
				SELECT
					COUNT(DISTINCT ROUND_ID)
				FROM LOG_RESPONSE_CARD
				WHERE JUDGE_USER_ID = U.ID
			) AS PROMPT_CARD_PLAY_COUNT,
			(SELECT COUNT(*) FROM LOG_SKIP WHERE USER_ID = U.ID) AS PROMPT_CARD_SKIP_COUNT,
			COALESCE(
				(
					SELECT
						SUM(AMOUNT)
					FROM LOG_CREDITS_SPENT
					WHERE USER_ID = U.ID
						AND AMOUNT > 0
				),
				0
			) AS CREDITS_SPENT_COUNT,
			COALESCE(
				(
					SELECT
						SUM(AMOUNT) * -1
					FROM LOG_CREDITS_SPENT
					WHERE USER_ID = U.ID
						AND AMOUNT < 0
				),
				0
			) AS CREDITS_EARNED_COUNT,
			(SELECT COUNT(*) FROM LOG_KICK WHERE USER_ID = U.ID) AS LOBBY_KICK_COUNT,
			(SELECT COUNT(*) FROM LOG_FLIP_TABLE WHERE USER_ID = U.ID) AS FLIP_TABLE_COUNT
		FROM USER AS U
		WHERE U.ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&result.UserName,
			&result.GamePlayCount,
			&result.GameWinCount,
			&result.RoundPlayCount,
			&result.RoundWinCount,
			&result.ResponseCardPlayCount,
			&result.ResponseCardDiscardCount,
			&result.PromptCardPlayCount,
			&result.PromptCardSkipCount,
			&result.CreditsSpentCount,
			&result.CreditsEarnedCount,
			&result.LobbyKickCount,
			&result.FlipTableCount,
		); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
	}

	return result, nil
}

func GetAchievementsUser(userId uuid.UUID) ([]Achievement, error) {
	var result []Achievement

	sqlString := `
		WITH CREDITS_SPENT AS (
				SELECT
					CATEGORY
				FROM LOG_CREDITS_SPENT
				WHERE USER_ID = ?
			)
		SELECT
			'Games Won',
			(SELECT COUNT(DISTINCT LOBBY_ID) FROM V_GAME_WINNER WHERE USER_ID = ?)
		UNION
		SELECT
			'Games Played',
			(SELECT COUNT(DISTINCT LOBBY_ID) FROM LOG_RESPONSE_CARD WHERE PLAYER_USER_ID = ?)
		UNION
		SELECT
			'Winning Streaks',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'WINNING-STREAK')
		UNION
		SELECT
			'Losing Streaks',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'LOSING-STREAK')
		UNION
		SELECT
			'Skipped Being Judge',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'SKIP-JUDGE')
		UNION
		SELECT
			'Lobby Alerts',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'ALERT')
		UNION
		SELECT
			'Gambles Made',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'GAMBLE')
		UNION
		SELECT
			'Gambles Won',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'GAMBLE-WIN')
		UNION
		SELECT
			'Bets Placed',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'BET')
		UNION
		SELECT
			'Bets Won',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'BET-WIN')
		UNION
		SELECT
			'Extra Responses',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'EXTRA-RESPONSE')
		UNION
		SELECT
			'Blocked Responses',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'BLOCK-RESPONSE')
		UNION
		SELECT
			'Stolen Cards Played',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'STEAL')
		UNION
		SELECT
			'Surprise Cards Played',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'SURPRISE')
		UNION
		SELECT
			'Find Cards Played',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'FIND')
		UNION
		SELECT
			'Wild Cards Played',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'WILD')
		UNION
		SELECT
			'Perks Purchased',
			(SELECT COUNT(*) FROM CREDITS_SPENT WHERE CATEGORY = 'PERK')
		UNION
		SELECT
			'Cards Discarded',
			(SELECT COUNT(*) FROM LOG_DISCARD WHERE USER_ID = ?)
		UNION
		SELECT
			'Prompts Skipped',
			(SELECT COUNT(*) FROM LOG_SKIP WHERE USER_ID = ?)
		UNION
		SELECT
			'Kicked From Lobby',
			(SELECT COUNT(*) FROM LOG_KICK WHERE USER_ID = ?)
		UNION
		SELECT
			'Flipped Tables',
			(SELECT COUNT(*) FROM LOG_FLIP_TABLE WHERE USER_ID = ?)
	`
	rows, err := query(sqlString, userId, userId, userId, userId, userId, userId, userId)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var achievement Achievement
		if err := rows.Scan(
			&achievement.Name,
			&achievement.Progress,
		); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}

		if achievement.Progress >= 100 {
			achievement.Progress = 100
		}

		result = append(result, achievement)
	}

	return result, nil
}

func GetStatsCard(cardId uuid.UUID) (StatCard, error) {
	var result StatCard

	sqlString := `
		SELECT
			D.NAME AS DECK_NAME,
			C.CATEGORY,
			C.TEXT,
			(
				SELECT
					COUNT(DISTINCT ID)
				FROM LOG_RESPONSE_CARD
				WHERE JUDGE_CARD_ID = C.ID
					OR PLAYER_CARD_ID = C.ID
			) AS PLAY_COUNT,
			(
				SELECT
					COUNT(DISTINCT LW.ID)
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
				WHERE LRC.PLAYER_CARD_ID = C.ID
			) AS WIN_COUNT,
			(SELECT COUNT(DISTINCT ID) FROM LOG_DISCARD WHERE CARD_ID = C.ID) AS DISCARD_COUNT,
			(SELECT COUNT(DISTINCT ID) FROM LOG_SKIP WHERE CARD_ID = C.ID) AS SKIP_COUNT
		FROM CARD AS C
			INNER JOIN DECK AS D ON D.ID = C.DECK_ID
		WHERE C.ID = ?
	`
	rows, err := query(sqlString, cardId)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&result.DeckName,
			&result.Category,
			&result.Text,
			&result.PlayCount,
			&result.WinCount,
			&result.DiscardCount,
			&result.SkipCount,
		); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
	}

	return result, nil
}
