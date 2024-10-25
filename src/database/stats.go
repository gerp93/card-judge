package database

import (
	"errors"
	"log"
	"reflect"

	"github.com/google/uuid"
)

type StatPersonal struct {
	RoundWinRatio    float64
	RoundWinCount    int
	CardPlayCount    int
	CardDrawCount    int
	CardDiscardCount int
	CardSkipCount    int
}

func GetPersonalStats(userId uuid.UUID) (StatPersonal, error) {
	var result StatPersonal

	sqlString := `
		SELECT
			(
				SELECT
					COALESCE((COUNT(DISTINCT LW.ID)*1.0) / (COUNT(DISTINCT LP.ID)*1.0),0.0)
				FROM LOG_WIN AS LW
					INNER JOIN LOG_PLAY AS LP ON LP.PLAYER_USER_ID = U.ID
				WHERE LW.PLAYER_USER_ID = U.ID
					AND LP.PLAYER_USER_ID = U.ID
				GROUP BY LP.PLAYER_USER_ID
			) AS ROUND_WIN_RATIO,
			(SELECT COUNT(*) FROM LOG_WIN WHERE PLAYER_USER_ID = U.ID) AS ROUND_WIN_COUNT,
			(SELECT COUNT(*) FROM LOG_PLAY WHERE PLAYER_USER_ID = U.ID) AS CARD_PLAY_COUNT,
			(SELECT COUNT(*) FROM LOG_DRAW WHERE PLAYER_USER_ID = U.ID) AS CARD_DRAW_COUNT,
			(SELECT COUNT(*) FROM LOG_DISCARD WHERE PLAYER_USER_ID = U.ID) AS CARD_DISCARD_COUNT,
			(SELECT COUNT(*) FROM LOG_SKIP WHERE PLAYER_USER_ID = U.ID) AS CARD_SKIP_COUNT
		FROM USER AS U
		WHERE U.ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&result.RoundWinRatio,
			&result.RoundWinCount,
			&result.CardPlayCount,
			&result.CardDrawCount,
			&result.CardDiscardCount,
			&result.CardSkipCount); err != nil {
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
	case "round-win-ratio":
		switch subject {
		case "player":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Player")
			sqlString = `
				SELECT
					COUNT(DISTINCT LP.ID) AS PLAY_COUNT,
					COUNT(DISTINCT LW.ID) AS WIN_COUNT,
					COALESCE((COUNT(DISTINCT LW.ID)*1.0) / (COUNT(DISTINCT LP.ID)*1.0),0.0) AS WIN_RATIO,
					U.NAME AS NAME
				FROM LOG_WIN AS LW
					INNER JOIN USER AS U ON U.ID = LW.PLAYER_USER_ID
					INNER JOIN LOG_PLAY AS LP ON LP.PLAYER_USER_ID = U.ID
				GROUP BY U.ID
				ORDER BY
					WIN_RATIO DESC,
					NAME ASC
				LIMIT 5
			`
		case "card":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LP.ID) AS PLAY_COUNT,
					COUNT(DISTINCT LW.ID) AS WIN_COUNT,
					COALESCE((COUNT(DISTINCT LW.ID)*1.0) / (COUNT(DISTINCT LP.ID)*1.0),0.0) AS WIN_RATIO,
					COALESCE(C.TEXT, LW.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_WIN AS LW
					LEFT JOIN CARD AS C ON C.ID = LW.CARD_ID
					LEFT JOIN LOG_PLAY AS LP ON LP.CARD_ID = C.ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					WIN_RATIO DESC,
					NAME ASC
				LIMIT 5
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Rounds Played")
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Win Ratio")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = `
				SELECT
					COUNT(DISTINCT LP.ID) AS PLAY_COUNT,
					COUNT(DISTINCT LW.ID) AS WIN_COUNT,
					COALESCE((COUNT(DISTINCT LW.ID)*1.0) / (COUNT(DISTINCT LP.ID)*1.0),0.0) AS WIN_RATIO,
					COALESCE(LP.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_PLAY AS LP
					INNER JOIN LOG_WIN AS LW ON (LW.SPECIAL_CATEGORY = LP.SPECIAL_CATEGORY) OR
												(LW.SPECIAL_CATEGORY IS NULL AND LP.SPECIAL_CATEGORY IS NULL)
				GROUP BY LP.SPECIAL_CATEGORY
				ORDER BY
					WIN_RATIO DESC,
					NAME ASC
				LIMIT 5
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
					COUNT(DISTINCT LW.ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_WIN AS LW
					INNER JOIN USER AS U ON U.ID = LW.PLAYER_USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
			`
		case "card":
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LW.ID) AS COUNT,
					COALESCE(C.TEXT, LW.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_WIN AS LW
					LEFT JOIN CARD AS C ON C.ID = LW.CARD_ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Rounds Won")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = `
				SELECT
					COUNT(DISTINCT LW.ID) AS COUNT,
					COALESCE(LW.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_WIN AS LW
				GROUP BY LW.SPECIAL_CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
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
					COUNT(DISTINCT LP.ID) AS COUNT,
					U.NAME AS NAME
				FROM LOG_PLAY AS LP
					INNER JOIN USER AS U ON U.ID = LP.PLAYER_USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
			`
		case "card":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LP.ID) AS COUNT,
					COALESCE(C.TEXT, LP.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_PLAY AS LP
					LEFT JOIN CARD AS C ON C.ID = LP.CARD_ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Cards Played")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = `
				SELECT
					COUNT(DISTINCT LP.ID) AS COUNT,
					COALESCE(LP.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_PLAY AS LP
				GROUP BY LP.SPECIAL_CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
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
					INNER JOIN USER AS U ON U.ID = LD.PLAYER_USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
			`
		case "card":
			resultHeaders = append(resultHeaders, "Cards Drawn")
			resultHeaders = append(resultHeaders, "Card")
			params = append(params, userId)
			sqlString = `
				SELECT
					COUNT(DISTINCT LD.ID) AS COUNT,
					COALESCE(C.TEXT, LD.SPECIAL_CATEGORY, 'Unknown') AS NAME
				FROM LOG_DRAW AS LD
					LEFT JOIN CARD AS C ON C.ID = LD.CARD_ID
				WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
				GROUP BY C.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Cards Drawn")
			resultHeaders = append(resultHeaders, "Special Category")
			sqlString = `
				SELECT
					COUNT(DISTINCT LD.ID) AS COUNT,
					COALESCE(LD.SPECIAL_CATEGORY, 'NONE') AS NAME
				FROM LOG_DRAW AS LD
				GROUP BY LD.SPECIAL_CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
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
					INNER JOIN USER AS U ON U.ID = LD.PLAYER_USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
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
				LIMIT 5
			`
		case "special-category":
			return resultHeaders, resultRows, errors.New("cannot combine discard with special category")
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
					INNER JOIN USER AS U ON U.ID = LS.PLAYER_USER_ID
				GROUP BY U.ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
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
				LIMIT 5
			`
		case "special-category":
			return resultHeaders, resultRows, errors.New("cannot combine skip with special category")
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
				FROM LOG_WIN AS LW
					INNER JOIN USER AS UJ ON UJ.ID = LW.JUDGE_USER_ID
					INNER JOIN USER AS UP ON UP.ID = LW.PLAYER_USER_ID
				WHERE UJ.ID = ?
				GROUP BY LW.JUDGE_USER_ID, LW.PLAYER_USER_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
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
				FROM LOG_WIN AS LW
					INNER JOIN USER AS UJ ON UJ.ID = LW.JUDGE_USER_ID
					INNER JOIN CARD AS CP ON CP.ID = LW.CARD_ID
				WHERE UJ.ID = ?
				GROUP BY LW.JUDGE_USER_ID, LW.CARD_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Judge Picking")
			resultHeaders = append(resultHeaders, "Special Category")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = `
				SELECT
					UJ.NAME AS JUDGE_NAME,
					COALESCE(LW.SPECIAL_CATEGORY, 'NONE') AS NAME,
					COUNT(LW.ID) AS COUNT
				FROM LOG_WIN AS LW
					INNER JOIN USER AS UJ ON UJ.ID = LW.JUDGE_USER_ID
				WHERE UJ.ID = ?
				GROUP BY LW.JUDGE_USER_ID, LW.SPECIAL_CATEGORY
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
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
				FROM LOG_WIN AS LW
					INNER JOIN USER AS UJ ON UJ.ID = LW.JUDGE_USER_ID
					INNER JOIN USER AS UP ON UP.ID = LW.PLAYER_USER_ID
				WHERE UP.ID = ?
				GROUP BY LW.JUDGE_USER_ID, LW.PLAYER_USER_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
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
				FROM LOG_WIN AS LW
					INNER JOIN CARD AS CJ ON CJ.ID = LW.CARD_ID
					INNER JOIN USER AS UP ON UP.ID = LW.PLAYER_USER_ID
				WHERE UP.ID = ?
				GROUP BY LW.CARD_ID, LW.PLAYER_USER_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
			`
		case "special-category":
			resultHeaders = append(resultHeaders, "Winner")
			resultHeaders = append(resultHeaders, "Special Category Played")
			resultHeaders = append(resultHeaders, "Count")
			params = append(params, userId)
			sqlString = `
				SELECT
					UP.NAME AS PLAYER_NAME,
					COALESCE(LW.SPECIAL_CATEGORY, 'NONE') AS NAME,
					COUNT(LW.ID) AS COUNT
				FROM LOG_WIN AS LW
					INNER JOIN USER AS UP ON UP.ID = LW.PLAYER_USER_ID
				WHERE UP.ID = ?
				GROUP BY LW.SPECIAL_CATEGORY, LW.PLAYER_USER_ID
				ORDER BY
					COUNT DESC,
					NAME ASC
				LIMIT 5
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
