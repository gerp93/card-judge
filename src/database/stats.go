package database

import (
	"errors"
	"log"
	"reflect"

	"github.com/google/uuid"
)

type StatPersonal struct {
	GamePlayCount      int
	GameWinCount       int
	RoundPlayCount     int
	RoundWinCount      int
	CardPlayCount      int
	CardDrawCount      int
	CardDiscardCount   int
	CardSkipCount      int
	CreditsSpentCount  int
	CreditsEarnedCount int
	LobbyKickCount     int
}

func GetPersonalStats(userId uuid.UUID) (StatPersonal, error) {
	var result StatPersonal

	sqlString := `
		SELECT
			(
				SELECT COUNT(DISTINCT LOBBY_ID)
				FROM LOG_RESPONSE_CARD
				WHERE PLAYER_USER_ID = U.ID
			) AS GAME_PLAY_COUNT,
			(
				SELECT
					COUNT(*)
				FROM (
						SELECT
							LRC.LOBBY_ID,
							LRC.PLAYER_USER_ID,
							COUNT(LW.ID) AS ROUND_WIN_COUNT,
							RANK() OVER (PARTITION BY LRC.LOBBY_ID ORDER BY ROUND_WIN_COUNT DESC) AS RANKING
						FROM LOG_RESPONSE_CARD AS LRC
							INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
						GROUP BY LRC.LOBBY_ID, LRC.PLAYER_USER_ID
					) AS ROUND_WINS
				WHERE PLAYER_USER_ID = U.ID
					AND RANKING = 1
			) AS GAME_WIN_COUNT,
			(
				SELECT COUNT(DISTINCT ROUND_ID)
				FROM LOG_RESPONSE_CARD
				WHERE PLAYER_USER_ID = U.ID
			) AS ROUND_PLAY_COUNT,
			(
				SELECT COUNT(DISTINCT LW.ID)
				FROM LOG_RESPONSE_CARD AS LRC
						INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
				WHERE LRC.PLAYER_USER_ID = U.ID
			) AS ROUND_WIN_COUNT,
			(
				SELECT COUNT(*)
				FROM LOG_RESPONSE_CARD
				WHERE PLAYER_USER_ID = U.ID
			) AS CARD_PLAY_COUNT,
			(SELECT COUNT(*) FROM LOG_DRAW WHERE USER_ID = U.ID) AS CARD_DRAW_COUNT,
			(SELECT COUNT(*) FROM LOG_DISCARD WHERE USER_ID = U.ID) AS CARD_DISCARD_COUNT,
			(SELECT COUNT(*) FROM LOG_SKIP WHERE USER_ID = U.ID) AS CARD_SKIP_COUNT,
			COALESCE(
				(
					SELECT SUM(AMOUNT)
					FROM LOG_CREDITS_SPENT
					WHERE USER_ID = U.ID
						AND AMOUNT > 0
				), 0
			) AS CREDITS_SPENT_COUNT,
			COALESCE(
				(
					SELECT SUM(AMOUNT) * -1
					FROM LOG_CREDITS_SPENT
					WHERE USER_ID = U.ID
						AND AMOUNT < 0
				), 0
			) AS CREDITS_EARNED_COUNT,
			(SELECT COUNT(*) FROM LOG_KICK WHERE USER_ID = U.ID) AS LOBBY_KICK_COUNT
		FROM USER AS U
		WHERE U.ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&result.GamePlayCount,
			&result.GameWinCount,
			&result.RoundPlayCount,
			&result.RoundWinCount,
			&result.CardPlayCount,
			&result.CardDrawCount,
			&result.CardDiscardCount,
			&result.CardSkipCount,
			&result.CreditsSpentCount,
			&result.CreditsEarnedCount,
			&result.LobbyKickCount); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
	}

	return result, nil
}

func GetLeaderboardStats(userId uuid.UUID, topic string, subject string) ([]string, [][]string, error) {
	resultHeaders := make([]string, 0)
	resultRows := make([][]string, 0)
	params := make([]any, 0)

	var sqlString string
	switch topic {
	case "game-win-ratio":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Games Played")
			resultHeaders = append(resultHeaders, "Games Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					GP.GAME_PLAY_COUNT AS PLAY_COUNT,
					COALESCE(GW.GAME_WIN_COUNT, 0) AS WIN_COUNT,
					COALESCE((GW.GAME_WIN_COUNT * 1.0) / (GP.GAME_PLAY_COUNT * 1.0), 0.0) AS WIN_RATIO,
					U.NAME AS NAME
				FROM USER AS U
						INNER JOIN (
							SELECT
								PLAYER_USER_ID,
								COUNT(DISTINCT LOBBY_ID) AS GAME_PLAY_COUNT
							FROM LOG_RESPONSE_CARD
							GROUP BY PLAYER_USER_ID
						) AS GP ON GP.PLAYER_USER_ID = U.ID
						LEFT JOIN (
							SELECT
								PLAYER_USER_ID,
								COUNT(*) AS GAME_WIN_COUNT
							FROM (
								SELECT
									LRC.PLAYER_USER_ID,
									COUNT(LRC.ID) AS ROUND_WIN_COUNT,
									RANK() OVER (PARTITION BY LRC.LOBBY_ID ORDER BY ROUND_WIN_COUNT DESC) AS RANKING
								FROM LOG_RESPONSE_CARD AS LRC
									INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
								GROUP BY LRC.LOBBY_ID, LRC.PLAYER_USER_ID
							) AS GAME_RANK
							WHERE RANKING = 1
							GROUP BY PLAYER_USER_ID
						) AS GW ON GW.PLAYER_USER_ID = U.ID
				ORDER BY
					WIN_RATIO DESC,
					PLAY_COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "game-win":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Games Won")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(*) AS COUNT,
					U.NAME AS NAME
				FROM (
						SELECT
							LRC.LOBBY_ID,
							LRC.PLAYER_USER_ID,
							COUNT(LW.ID) AS ROUND_WIN_COUNT,
							RANK() OVER (PARTITION BY LRC.LOBBY_ID ORDER BY ROUND_WIN_COUNT DESC) AS RANKING
						FROM LOG_RESPONSE_CARD AS LRC
							INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
						GROUP BY LRC.LOBBY_ID, LRC.PLAYER_USER_ID
					) AS RW
					INNER JOIN USER AS U ON U.ID = RW.PLAYER_USER_ID
				WHERE RW.RANKING = 1
				GROUP BY RW.PLAYER_USER_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "game-play":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Games Played")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.LOBBY_ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN USER AS U ON U.ID = LRC.PLAYER_USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Games Played")
			resultHeaders = append(resultHeaders, "Card")
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.LOBBY_ID) AS COUNT,
					COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Games Played")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.LOBBY_ID) AS COUNT,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
				GROUP BY LRC.SPECIAL_CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
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
			sqlString = `
				SELECT
					PLAY_COUNT,
					WIN_COUNT,
					COALESCE((WIN_COUNT * 1.0) / (PLAY_COUNT * 1.0), 0.0) AS WIN_RATIO,
					NAME
				FROM (
						SELECT
							COUNT(DISTINCT LRC.ROUND_ID) AS PLAY_COUNT,
							COUNT(DISTINCT LW.ID)        AS WIN_COUNT,
							U.NAME                       AS NAME
						FROM LOG_RESPONSE_CARD AS LRC
							LEFT JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
							INNER JOIN USER AS U ON U.ID = LRC.PLAYER_USER_ID
						GROUP BY U.ID
					) AS T
				ORDER BY
					WIN_RATIO DESC,
					PLAY_COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					PLAY_COUNT,
					WIN_COUNT,
					COALESCE((WIN_COUNT * 1.0) / (PLAY_COUNT * 1.0), 0.0) AS WIN_RATIO,
					NAME
				FROM (
						SELECT
							COUNT(DISTINCT LRC.ROUND_ID) AS PLAY_COUNT,
							COUNT(DISTINCT LW.ID)        AS WIN_COUNT,
							COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
						FROM LOG_RESPONSE_CARD AS LRC
							LEFT JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
							LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
						WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
						GROUP BY C.ID
					) AS T
				ORDER BY
					WIN_RATIO DESC,
					PLAY_COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = `
				SELECT
					PLAY_COUNT,
					WIN_COUNT,
					COALESCE((WIN_COUNT * 1.0) / (PLAY_COUNT * 1.0), 0.0) AS WIN_RATIO,
					NAME
				FROM (
						SELECT
							COUNT(DISTINCT LRC.ROUND_ID) AS PLAY_COUNT,
							COUNT(DISTINCT LW.ID)        AS WIN_COUNT,
							COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
						FROM LOG_RESPONSE_CARD AS LRC
							LEFT JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
						GROUP BY LRC.SPECIAL_CATEGORY
					) AS T
				ORDER BY
					WIN_RATIO DESC,
					PLAY_COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "round-win":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS U ON U.ID = LRC.PLAYER_USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
				GROUP BY LRC.SPECIAL_CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "round-play":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN USER AS U ON U.ID = LRC.PLAYER_USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.ROUND_ID) AS COUNT,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
				GROUP BY LRC.SPECIAL_CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "round-judge-time":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Avg. Seconds")
			resultHeaders = append(resultHeaders, "Judge")
			sqlString = `
				SELECT
					AVG(TIMESTAMPDIFF(SECOND, LAST_PLAY.CREATED_ON_DATE, ROUND_WIN.CREATED_ON_DATE)) AS AVG_SECONDS,
					U.NAME
				FROM (SELECT ROUND_ID,
							CREATED_ON_DATE,
							RANK() OVER (PARTITION BY ROUND_ID ORDER BY CREATED_ON_DATE DESC) AS REV_PLAY_ORDER
					FROM LOG_RESPONSE_CARD) AS LAST_PLAY
						INNER JOIN (SELECT LRC.ROUND_ID,
											LRC.JUDGE_USER_ID,
											LW.CREATED_ON_DATE
									FROM LOG_WIN AS LW
											INNER JOIN LOG_RESPONSE_CARD AS LRC ON LRC.RESPONSE_ID = LW.RESPONSE_ID
									GROUP BY LRC.RESPONSE_ID) AS ROUND_WIN ON ROUND_WIN.ROUND_ID = LAST_PLAY.ROUND_ID
						INNER JOIN USER AS U ON U.ID = ROUND_WIN.JUDGE_USER_ID
				WHERE LAST_PLAY.REV_PLAY_ORDER = 1
				GROUP BY ROUND_WIN.JUDGE_USER_ID
				ORDER BY AVG_SECONDS, NAME;
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "round-play-time":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Avg. Seconds")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					AVG(TIMESTAMPDIFF(SECOND, PREV_WIN.CREATED_ON_DATE, LAST_PLAY.CREATED_ON_DATE)) AS AVG_SECONDS,
					U.NAME
				FROM (SELECT ROUND_ID,
							PLAYER_USER_ID,
							CREATED_ON_DATE,
							RANK() OVER (PARTITION BY ROUND_ID, PLAYER_USER_ID ORDER BY CREATED_ON_DATE DESC) AS REV_PLAY_ORDER
					FROM LOG_RESPONSE_CARD) AS LAST_PLAY
						INNER JOIN (SELECT ROUND_ID,
											LAG(ROUND_ID) OVER (PARTITION BY LOBBY_ID ORDER BY CREATED_ON_DATE) AS PREV_ROUND_ID
									FROM LOG_RESPONSE_CARD
									GROUP BY ROUND_ID) AS ROUND_AND_PREV ON ROUND_AND_PREV.ROUND_ID = LAST_PLAY.ROUND_ID
						INNER JOIN (SELECT LRC.ROUND_ID,
											LW.CREATED_ON_DATE
									FROM LOG_WIN AS LW
											INNER JOIN LOG_RESPONSE_CARD AS LRC ON LRC.RESPONSE_ID = LW.RESPONSE_ID
									GROUP BY LRC.RESPONSE_ID) AS PREV_WIN ON PREV_WIN.ROUND_ID = ROUND_AND_PREV.PREV_ROUND_ID
						INNER JOIN USER AS U ON U.ID = LAST_PLAY.PLAYER_USER_ID
				WHERE LAST_PLAY.REV_PLAY_ORDER = 1
				GROUP BY LAST_PLAY.PLAYER_USER_ID
				ORDER BY AVG_SECONDS, NAME;
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "card-play":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN USER AS U ON U.ID = LRC.PLAYER_USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.ID) AS COUNT,
					COALESCE(C.TEXT, LRC.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
					LEFT JOIN CARD AS C ON C.ID = LRC.PLAYER_CARD_ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = `
				SELECT
					COUNT(DISTINCT LRC.ID) AS COUNT,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_RESPONSE_CARD AS LRC
				GROUP BY LRC.SPECIAL_CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "card-draw":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Cards Drawn")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(DISTINCT LD.ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_DRAW AS LD
					INNER JOIN USER AS U ON U.ID = LD.USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Cards Drawn")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LD.ID) AS COUNT,
					COALESCE(C.TEXT, 'Unknown') AS NAME
				FROM LOG_DRAW AS LD
					LEFT JOIN CARD AS C ON C.ID = LD.CARD_ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "card-discard":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Cards Discarded")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(DISTINCT LD.ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_DISCARD AS LD
					INNER JOIN USER AS U ON U.ID = LD.USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Cards Discarded")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LD.ID) AS COUNT,
					COALESCE(C.TEXT, 'Unknown') AS NAME
				FROM LOG_DISCARD AS LD
					LEFT JOIN CARD AS C ON C.ID = LD.CARD_ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "card-skip":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Cards Skipped")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(DISTINCT LS.ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_SKIP AS LS
					INNER JOIN USER AS U ON U.ID = LS.USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Cards Skipped")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LS.ID) AS COUNT,
					COALESCE(C.TEXT, 'Unknown') AS NAME
				FROM LOG_SKIP AS LS
					LEFT JOIN CARD AS C ON C.ID = LS.CARD_ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
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
			sqlString = `
				SELECT
					UJ.NAME AS JUDGE_NAME,
					UP.NAME AS NAME,
					COUNT(LW.ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UJ ON UJ.ID = LRC.JUDGE_USER_ID
					INNER JOIN USER AS UP ON UP.ID = LRC.PLAYER_USER_ID
				WHERE UJ.ID = ?
				GROUP BY LRC.JUDGE_USER_ID, LRC.PLAYER_USER_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Judge Picking")
			resultHeaders = append(resultHeaders, "Card")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = `
				SELECT
					UJ.NAME AS JUDGE_NAME,
					CP.TEXT AS NAME,
					COUNT(LW.ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UJ ON UJ.ID = LRC.JUDGE_USER_ID
					INNER JOIN CARD AS CP ON CP.ID = LRC.PLAYER_CARD_ID
				WHERE UJ.ID = ?
				GROUP BY LRC.JUDGE_USER_ID, LRC.PLAYER_CARD_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Judge Picking")
			resultHeaders = append(resultHeaders, "Special Category")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = `
				SELECT
					UJ.NAME AS JUDGE_NAME,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME,
					COUNT(LW.ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UJ ON UJ.ID = LRC.JUDGE_USER_ID
				WHERE UJ.ID = ?
				GROUP BY LRC.JUDGE_USER_ID, LRC.SPECIAL_CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
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
			sqlString = `
				SELECT
					UP.NAME AS PLAYER_NAME,
					UJ.NAME AS NAME,
					COUNT(LW.ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UJ ON UJ.ID = LRC.JUDGE_USER_ID
					INNER JOIN USER AS UP ON UP.ID = LRC.PLAYER_USER_ID
				WHERE UP.ID = ?
				GROUP BY LRC.JUDGE_USER_ID, LRC.PLAYER_USER_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "card":
			resultHeaders = append(resultHeaders, "Winner")
			resultHeaders = append(resultHeaders, "Card Played")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = `
				SELECT
					UP.NAME AS PLAYER_NAME,
					CJ.TEXT AS NAME,
					COUNT(LW.ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN CARD AS CJ ON CJ.ID = LRC.PLAYER_CARD_ID
					INNER JOIN USER AS UP ON UP.ID = LRC.PLAYER_USER_ID
				WHERE UP.ID = ?
				GROUP BY LRC.PLAYER_CARD_ID, LRC.PLAYER_USER_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Winner")
			resultHeaders = append(resultHeaders, "Special Category Played")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = `
				SELECT
					UP.NAME AS PLAYER_NAME,
					COALESCE(LRC.SPECIAL_CATEGORY, 'NONE') AS NAME,
					COUNT(LW.ID) AS COUNT
				FROM LOG_RESPONSE_CARD AS LRC
					INNER JOIN LOG_WIN AS LW ON LW.RESPONSE_ID = LRC.RESPONSE_ID
					INNER JOIN USER AS UP ON UP.ID = LRC.PLAYER_USER_ID
				WHERE UP.ID = ?
				GROUP BY LRC.SPECIAL_CATEGORY, LRC.PLAYER_USER_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-spent":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Spent")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					SUM(LCS.AMOUNT) AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.AMOUNT > 0
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-earned":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Earned")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					SUM(LCS.AMOUNT) * -1 AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.AMOUNT < 0
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-spent-category":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Spent")
			resultHeaders = append(resultHeaders, "Category")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					SUM(LCS.AMOUNT) AS COUNT,
					LCS.CATEGORY AS CATEGORY,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.AMOUNT > 0
				GROUP BY U.ID, LCS.CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC,
					CATEGORY ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-earned-category":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Earned")
			resultHeaders = append(resultHeaders, "Category")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					SUM(LCS.AMOUNT) * -1 AS COUNT,
					LCS.CATEGORY AS CATEGORY,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.AMOUNT < 0
				GROUP BY U.ID, LCS.CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC,
					CATEGORY ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-spent-game":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Spent in a Game")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					SUM(LCS.AMOUNT) AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.AMOUNT > 0
				GROUP BY U.ID, LCS.LOBBY_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "credits-earned-game":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Credits Earned in a Game")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					SUM(LCS.AMOUNT) * -1 AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.AMOUNT < 0
				GROUP BY U.ID, LCS.LOBBY_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "gamble":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Gamble")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					MAX(LCS.AMOUNT) AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CATEGORY = 'GAMBLE'
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "gamble-win":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Gamble Win")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					MIN(LCS.AMOUNT) * -1 AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CATEGORY = 'GAMBLE-WIN'
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "bet":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Bet")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					MAX(LCS.AMOUNT) AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CATEGORY = 'BET'
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "bet-win":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Bet Win")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					MIN(LCS.AMOUNT) * -1 AS COUNT,
					U.NAME AS NAME
				FROM LOG_CREDITS_SPENT AS LCS
					INNER JOIN USER AS U ON U.ID = LCS.USER_ID
				WHERE LCS.CATEGORY = 'BET-WIN'
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
		default:
			return resultHeaders, resultRows, errors.New("invalid subject provided")
		}
	case "kick":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Kicked")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(*) AS COUNT,
					U.NAME AS NAME
				FROM LOG_KICK AS LK
					INNER JOIN USER AS U ON U.ID = LK.USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 10
			`
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
