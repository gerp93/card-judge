package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
)

type Lobby struct {
	Id            uuid.UUID
	CreatedOnDate time.Time

	Name         string
	GameType     string
	Message      sql.NullString
	PasswordHash sql.NullString

	DrawPriority        string
	HandSize            int
	FreeCredits         int
	WinStreakThreshold  int
	LoseStreakThreshold int
}

type LobbyDetails struct {
	Lobby
	UserCount int
}

type LobbyGameInfo struct {
	LobbyName string

	DrawPilePromptCount   int
	DrawPileResponseCount int
	DrawPileDeckNames     string

	JudgeName sql.NullString
}

type PlayerHandData struct {
	LobbyId uuid.UUID

	PlayerId      uuid.UUID
	PlayerIsJudge bool
	PlayerIsReady bool
	PlayerHand    []Card
}

type PlayerSpecialsData struct {
	LobbyId                  uuid.UUID
	LobbyFreeCredits         int
	LobbyWinStreakThreshold  int
	LobbyLoseStreakThreshold int

	BoardHasAnySpecial  bool
	BoardHasAnyRevealed bool
	BoardResponses      []boardResponse

	Opponents []opponentData

	PlayerId               uuid.UUID
	PlayerIsJudge          bool
	PlayerIsWinning        bool
	PlayerIsReady          bool
	PlayerWinningStreak    int
	PlayerLosingStreak     int
	PlayerCreditsSpent     int
	PlayerBetOnWin         int
	PlayerExtraResponses   int
	PlayerCreditsRemaining int
}

type LobbyGameBoardData struct {
	LobbyId uuid.UUID

	JudgeCardText      sql.NullString
	JudgeCardYouTube   sql.NullString
	JudgeCardImage     sql.NullString
	JudgeBlankCount    int
	JudgeResponseCount int

	BoardIsReady        bool
	BoardHasAnySpecial  bool
	BoardHasAnyRevealed bool
	BoardIsAllRevealed  bool
	BoardIsAllRuledOut  bool
	BoardResponses      []boardResponse

	PlayerId        uuid.UUID
	PlayerIsJudge   bool
	PlayerResponses []boardResponse
}

type LobbyGameStatsData struct {
	LobbyId uuid.UUID

	PlayerId uuid.UUID

	Wins           []nameCountRow
	UpcomingJudges []string
	KickVotes      []kickVote
}

type opponentData struct {
	PlayerId uuid.UUID
	UserName string
}

type boardResponse struct {
	ResponseId     uuid.UUID
	IsRevealed     bool
	IsRuledOut     bool
	PlayerId       uuid.UUID
	PlayerUserName string
	ResponseCards  []boardResponseCard
}

type boardResponseCard struct {
	ResponseCardId uuid.UUID
	Card
	SpecialCategory sql.NullString
}

type nameCountRow struct {
	Name  string
	Count int
}

type kickVote struct {
	PlayerId uuid.UUID
	UserName string
	Voted    bool
}

func SearchLobbies(name string, page int) ([]LobbyDetails, error) {
	name = "%" + name + "%"

	if page < 1 {
		page = 1
	}

	sqlString := `
		SELECT
			L.ID,
			L.CREATED_ON_DATE,
			L.NAME,
			L.PASSWORD_HASH,
			L.DRAW_PRIORITY,
			L.HAND_SIZE,
			L.FREE_CREDITS,
			L.WIN_STREAK_THRESHOLD,
			L.LOSE_STREAK_THRESHOLD,
			COUNT(P.ID) AS USER_COUNT
		FROM LOBBY AS L
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = L.ID
			AND P.IS_ACTIVE = 1
		WHERE L.NAME LIKE ?
			AND (L.GAME_TYPE IS NULL OR L.GAME_TYPE != 'chronology')
		GROUP BY L.ID
		ORDER BY L.NAME
		LIMIT 10 OFFSET ?
	`
	rows, err := query(sqlString, name, (page-1)*10)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]LobbyDetails, 0)
	for rows.Next() {
		var ld LobbyDetails
		if err := rows.Scan(
			&ld.Id,
			&ld.CreatedOnDate,
			&ld.Name,
			&ld.PasswordHash,
			&ld.DrawPriority,
			&ld.HandSize,
			&ld.FreeCredits,
			&ld.WinStreakThreshold,
			&ld.LoseStreakThreshold,
			&ld.UserCount,
		); err != nil {
			log.Println(err)
			return nil, errors.New("failed to scan row in query results")
		}
		result = append(result, ld)
	}
	return result, nil
}

func CountLobbies(name string) (int, error) {
	name = "%" + name + "%"

	sqlString := `
		SELECT
			COUNT(*)
		FROM LOBBY
		WHERE NAME LIKE ?
			AND (GAME_TYPE IS NULL OR GAME_TYPE != 'chronology')
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Println(err)
			return 0, errors.New("failed to scan row in query results")
		}
	}

	return count, nil
}

func GetLobby(id uuid.UUID) (Lobby, error) {
	var lobby Lobby

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			NAME,
			COALESCE(GAME_TYPE, 'cah') AS GAME_TYPE,
			MESSAGE,
			PASSWORD_HASH,
			DRAW_PRIORITY,
			HAND_SIZE,
			FREE_CREDITS,
			WIN_STREAK_THRESHOLD,
			LOSE_STREAK_THRESHOLD
		FROM LOBBY
		WHERE ID = ?
	`
	rows, err := query(sqlString, id)
	if err != nil {
		return lobby, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&lobby.Id,
			&lobby.CreatedOnDate,
			&lobby.Name,
			&lobby.GameType,
			&lobby.Message,
			&lobby.PasswordHash,
			&lobby.DrawPriority,
			&lobby.HandSize,
			&lobby.FreeCredits,
			&lobby.WinStreakThreshold,
			&lobby.LoseStreakThreshold); err != nil {
			log.Println(err)
			return lobby, errors.New("failed to scan row in query results")
		}
	}

	return lobby, nil
}

func GetLobbyPasswordHash(id uuid.UUID) (sql.NullString, error) {
	var passwordHash sql.NullString

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM LOBBY
		WHERE ID = ?
	`
	rows, err := query(sqlString, id)
	if err != nil {
		return passwordHash, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&passwordHash); err != nil {
			log.Println(err)
			return passwordHash, errors.New("failed to scan row in query results")
		}
	}

	return passwordHash, nil
}

func CreateLobby(name string, message string, password string, drawPriority string, handSize int, freeCredits int, winStreakThreshold int, loseStreakThreshold int) (uuid.UUID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to generate new id")
	}

	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to hash password")
	}

	sqlString := `
		INSERT INTO LOBBY(
			ID,
			NAME,
			MESSAGE,
			PASSWORD_HASH,
			DRAW_PRIORITY,
			HAND_SIZE,
			FREE_CREDITS,
			WIN_STREAK_THRESHOLD,
			LOSE_STREAK_THRESHOLD
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	if message == "" {
		if password == "" {
			return id, execute(sqlString, id, name, nil, nil, drawPriority, handSize, freeCredits, winStreakThreshold, loseStreakThreshold)
		} else {
			return id, execute(sqlString, id, name, nil, passwordHash, drawPriority, handSize, freeCredits, winStreakThreshold, loseStreakThreshold)
		}
	} else {
		if password == "" {
			return id, execute(sqlString, id, name, message, nil, drawPriority, handSize, freeCredits, winStreakThreshold, loseStreakThreshold)
		} else {
			return id, execute(sqlString, id, name, message, passwordHash, drawPriority, handSize, freeCredits, winStreakThreshold, loseStreakThreshold)
		}
	}
}

func SyncDecksInLobby(lobbyId uuid.UUID, deckIds []uuid.UUID) error {
	if len(deckIds) == 0 {
		return errors.New("cannot sync decks in lobby, no deck ids provided")
	}

	var err error

	err = addDecksToLobby(lobbyId, deckIds)
	if err != nil {
		return err
	}

	err = removeDecksFromLobby(lobbyId, deckIds)
	if err != nil {
		return err
	}

	return nil
}

func addDecksToLobby(lobbyId uuid.UUID, deckIds []uuid.UUID) error {
	sqlString := fmt.Sprintf(`
		INSERT INTO DRAW_PILE (LOBBY_ID, CARD_ID)
		SELECT
			? AS LOBBY_ID,
			C.ID AS CARD_ID
		FROM CARD AS C
			LEFT JOIN (
					SELECT
						DISTINCT
						E_C.DECK_ID
					FROM DRAW_PILE AS E_DP
						INNER JOIN CARD AS E_C ON E_C.ID = E_DP.CARD_ID
					WHERE E_DP.LOBBY_ID = ?
				) AS E ON E.DECK_ID = C.DECK_ID
		WHERE C.DECK_ID IN (%s)
			AND E.DECK_ID IS NULL
	`, strings.Repeat("?,", len(deckIds)-1)+"?")

	args := make([]any, len(deckIds)+2)
	args[0] = lobbyId
	args[1] = lobbyId
	for i, deckId := range deckIds {
		args[i+2] = deckId
	}

	err := execute(sqlString, args...)
	if err != nil {
		return err
	}

	return nil
}

func removeDecksFromLobby(lobbyId uuid.UUID, deckIds []uuid.UUID) error {
	sqlString := fmt.Sprintf(`
		DELETE DP
		FROM DRAW_PILE AS DP
			INNER JOIN CARD AS C ON C.ID = DP.CARD_ID
		WHERE DP.LOBBY_ID = ?
			AND C.DECK_ID NOT IN (%s)
	`, strings.Repeat("?,", len(deckIds)-1)+"?")

	args := make([]any, len(deckIds)+1)
	args[0] = lobbyId
	for i, deckId := range deckIds {
		args[i+1] = deckId
	}

	err := execute(sqlString, args...)
	if err != nil {
		return err
	}

	return nil
}

func AddUserToLobby(lobbyId uuid.UUID, userId uuid.UUID) (uuid.UUID, error) {
	player, err := GetLobbyUserPlayer(lobbyId, userId)
	if err != nil {
		log.Println(err)
		return player.Id, errors.New("failed to get player")
	}

	if player.Id == uuid.Nil {
		player.Id, err = uuid.NewUUID()
		if err != nil {
			log.Println(err)
			return player.Id, errors.New("failed to generate new player id")
		}
	}

	sqlString := "CALL SP_SET_PLAYER_ACTIVE (?, ?, ?)"
	err = execute(sqlString, player.Id, lobbyId, userId)
	return player.Id, err
}

func SetPlayerInactive(lobbyId uuid.UUID, userId uuid.UUID) error {
	sqlString := "CALL SP_SET_PLAYER_INACTIVE (?, ?)"
	return execute(sqlString, lobbyId, userId)
}

func GetLobbyId(name string) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM LOBBY
		WHERE NAME = ?
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return id, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			log.Println(err)
			return id, errors.New("failed to scan row in query results")
		}
	}

	return id, nil
}

func GetPlayerLobbyId(playerId uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			LOBBY_ID
		FROM PLAYER
		WHERE ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return id, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			log.Println(err)
			return id, errors.New("failed to scan row in query results")
		}
	}

	return id, nil
}

func SetLobbyName(id uuid.UUID, name string) error {
	sqlString := `
		UPDATE LOBBY
		SET NAME = ?
		WHERE ID = ?
	`
	return execute(sqlString, name, id)
}

func SetLobbyMessage(id uuid.UUID, message string) error {
	sqlString := `
		UPDATE LOBBY
		SET MESSAGE = ?
		WHERE ID = ?
	`
	if message == "" {
		return execute(sqlString, nil, id)
	} else {
		return execute(sqlString, message, id)
	}
}

func SetLobbyDrawPriority(id uuid.UUID, drawPriority string) error {
	sqlString := `
		UPDATE LOBBY
		SET DRAW_PRIORITY = ?
		WHERE ID = ?
	`
	return execute(sqlString, drawPriority, id)
}

func SetLobbyHandSize(id uuid.UUID, handSize int) error {
	sqlString := `
		UPDATE LOBBY
		SET HAND_SIZE = ?
		WHERE ID = ?
	`
	return execute(sqlString, handSize, id)
}

func SetLobbyFreeCredits(id uuid.UUID, freeCredits int) error {
	sqlString := `
		UPDATE LOBBY
		SET FREE_CREDITS = ?
		WHERE ID = ?
	`
	return execute(sqlString, freeCredits, id)
}

func SetLobbyWinStreakThreshold(id uuid.UUID, winStreakThreshold int) error {
	sqlString := `
		UPDATE LOBBY
		SET WIN_STREAK_THRESHOLD = ?
		WHERE ID = ?
	`
	return execute(sqlString, winStreakThreshold, id)
}

func SetLobbyLoseStreakThreshold(id uuid.UUID, loseStreakThreshold int) error {
	sqlString := `
		UPDATE LOBBY
		SET LOSE_STREAK_THRESHOLD = ?
		WHERE ID = ?
	`
	return execute(sqlString, loseStreakThreshold, id)
}

func DeleteLobby(lobbyId uuid.UUID) error {
	sqlString := `
		DELETE
		FROM LOBBY
		WHERE ID = ?
	`
	return execute(sqlString, lobbyId)
}

func GetLobbyGameInfo(lobbyId uuid.UUID) (LobbyGameInfo, error) {
	var data LobbyGameInfo

	sqlString := `
		SELECT
			L.NAME AS LOBBY_NAME,
			(
				SELECT
					COUNT(*)
				FROM DRAW_PILE AS DP
					INNER JOIN CARD AS DPC ON DPC.ID = DP.CARD_ID
				WHERE DP.LOBBY_ID = L.ID
					AND DPC.CATEGORY = 'PROMPT'
			) AS DRAW_PILE_PROMPT_COUNT,
			(
				SELECT
					COUNT(*)
				FROM DRAW_PILE AS DP
					INNER JOIN CARD AS DPC ON DPC.ID = DP.CARD_ID
				WHERE DP.LOBBY_ID = L.ID
					AND DPC.CATEGORY = 'RESPONSE'
			) AS DRAW_PILE_RESPONSE_COUNT,
			(
				SELECT
					JU.NAME
				FROM USER AS JU
					INNER JOIN PLAYER AS JP ON JP.USER_ID = JU.ID
				WHERE JP.ID = J.PLAYER_ID
			) AS JUDGE_NAME
		FROM LOBBY AS L
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = L.ID
		WHERE L.ID = ?
	`
	rows, err := query(sqlString, lobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&data.LobbyName,
			&data.DrawPilePromptCount,
			&data.DrawPileResponseCount,
			&data.JudgeName,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
	}

	sqlString = `
		SELECT DISTINCT
			DPD.NAME
		FROM DRAW_PILE AS DP
			INNER JOIN CARD AS DPC ON DPC.ID = DP.CARD_ID
			INNER JOIN DECK AS DPD ON DPD.ID = DPC.DECK_ID
		WHERE DP.LOBBY_ID = ?
		ORDER BY DPD.NAME
	`
	rows, err = query(sqlString, lobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var deckName string
		if err := rows.Scan(&deckName); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.DrawPileDeckNames += deckName + "&#010;"
	}

	return data, nil
}

func GetPlayerHandData(playerId uuid.UUID) (PlayerHandData, error) {
	var data PlayerHandData

	sqlString := `
		SELECT
			L.ID AS LOBBY_ID,
			P.ID AS PLAYER_ID,
			IF(FN_GET_LOBBY_JUDGE_PLAYER_ID(L.ID) = P.ID, 1, 0) AS PLAYER_IS_JUDGE
		FROM PLAYER AS P
			INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
		WHERE P.ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&data.LobbyId,
			&data.PlayerId,
			&data.PlayerIsJudge,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
	}

	sqlString = `
		SELECT
			IF(
				R.ID IS NULL -- CANNOT PLAY
				OR COUNT(RC.ID) = -- CARDS PLAYED
				(
					J.BLANK_COUNT * -- CARDS PER RESPONSE
					(J.RESPONSE_COUNT + P.EXTRA_RESPONSES) -- PLAYER RESPONSES
				),
				1,
				0
			) AS PLAYER_IS_READY
		FROM PLAYER AS P
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = P.LOBBY_ID
			LEFT JOIN RESPONSE AS R ON R.PLAYER_ID = P.ID
			LEFT JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
		WHERE P.ID = ?
		GROUP BY P.ID
	`
	rows, err = query(sqlString, data.PlayerId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var isReady bool
		if err := rows.Scan(&isReady); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}

		data.PlayerIsReady = isReady
	}

	sqlString = `
		SELECT
			C.ID,
			C.TEXT,
			C.YOUTUBE,
			C.IMAGE
		FROM HAND AS H
			INNER JOIN CARD AS C ON C.ID = H.CARD_ID
		WHERE H.PLAYER_ID = ?
		ORDER BY C.TEXT
	`
	rows, err = query(sqlString, data.PlayerId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var card Card
		var imageBytes []byte
		if err := rows.Scan(
			&card.Id,
			&card.Text,
			&card.YouTube,
			&imageBytes,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}

		card.Image.Valid = imageBytes != nil
		if card.Image.Valid {
			card.Image.String = base64.StdEncoding.EncodeToString(imageBytes)
		}

		data.PlayerHand = append(data.PlayerHand, card)
	}

	return data, nil
}

func GetPlayerSpecialsData(playerId uuid.UUID) (PlayerSpecialsData, error) {
	var data PlayerSpecialsData

	sqlString := `
		SELECT
			L.ID AS LOBBY_ID,
			L.FREE_CREDITS AS LOBBY_FREE_CREDITS,
			L.WIN_STREAK_THRESHOLD AS LOBBY_WIN_STREAK_THRESHOLD,
			L.LOSE_STREAK_THRESHOLD AS LOBBY_LOSE_STREAK_THRESHOLD,
			P.ID AS PLAYER_ID,
			IF(FN_GET_LOBBY_JUDGE_PLAYER_ID(L.ID) = P.ID, 1, 0) AS PLAYER_IS_JUDGE,
			FN_GET_PLAYER_IS_WINNING(P.ID) AS PLAYER_IS_WINNING,
			P.WINNING_STREAK AS PLAYER_WINNING_STREAK,
			P.LOSING_STREAK AS PLAYER_LOSING_STREAK,
			P.CREDITS_SPENT AS PLAYER_CREDITS_SPENT,
			P.BET_ON_WIN AS PLAYER_BET_ON_WIN,
			P.EXTRA_RESPONSES AS PLAYER_EXTRA_RESPONSES
		FROM PLAYER AS P
			INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = L.ID
		WHERE P.ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&data.LobbyId,
			&data.LobbyFreeCredits,
			&data.LobbyWinStreakThreshold,
			&data.LobbyLoseStreakThreshold,
			&data.PlayerId,
			&data.PlayerIsJudge,
			&data.PlayerIsWinning,
			&data.PlayerWinningStreak,
			&data.PlayerLosingStreak,
			&data.PlayerCreditsSpent,
			&data.PlayerBetOnWin,
			&data.PlayerExtraResponses,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
	}

	data.PlayerCreditsRemaining = data.LobbyFreeCredits - data.PlayerCreditsSpent

	sqlString = `
		SELECT
			R.ID
		FROM PLAYER AS P
			INNER JOIN RESPONSE AS R ON R.PLAYER_ID = P.ID
		WHERE P.LOBBY_ID = ?
			AND R.IS_REVEALED = 1
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	data.BoardHasAnyRevealed = rows.Next()

	sqlString = `
		SELECT
			P.BET_ON_WIN,
			P.EXTRA_RESPONSES,
			RC.SPECIAL_CATEGORY
		FROM PLAYER AS P
			LEFT JOIN RESPONSE AS R ON R.PLAYER_ID = P.ID
			LEFT JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
		WHERE P.LOBBY_ID = ?
			AND P.IS_ACTIVE = 1
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	data.BoardHasAnySpecial = false
	for rows.Next() {
		var betOnWin int
		var extraResponse int
		var specialCategory sql.NullString
		if err := rows.Scan(
			&betOnWin,
			&extraResponse,
			&specialCategory,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		if betOnWin > 0 || extraResponse > 0 || specialCategory.Valid {
			data.BoardHasAnySpecial = true
			break
		}
	}

	sqlString = `
		SELECT
			IF(
				R.ID IS NULL -- CANNOT PLAY
				OR COUNT(RC.ID) = -- CARDS PLAYED
				(
					J.BLANK_COUNT * -- CARDS PER RESPONSE
					(J.RESPONSE_COUNT + P.EXTRA_RESPONSES) -- PLAYER RESPONSES
				),
				1,
				0
			) AS PLAYER_IS_READY
		FROM PLAYER AS P
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = P.LOBBY_ID
			LEFT JOIN RESPONSE AS R ON R.PLAYER_ID = P.ID
			LEFT JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
		WHERE P.ID = ?
		GROUP BY P.ID
	`
	rows, err = query(sqlString, data.PlayerId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var isReady bool
		if err := rows.Scan(&isReady); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}

		data.PlayerIsReady = isReady
	}

	sqlString = `
		SELECT
			P.ID AS PLAYER_ID,
			U.NAME AS USER_NAME
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
			INNER JOIN RESPONSE AS R ON R.PLAYER_ID = P.ID
		WHERE P.ID <> ?
			AND P.LOBBY_ID = ?
		ORDER BY U.NAME
	`
	rows, err = query(sqlString, playerId, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var row opponentData
		if err := rows.Scan(
			&row.PlayerId,
			&row.UserName,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.Opponents = append(data.Opponents, row)
	}

	return data, nil
}

func GetLobbyGameBoardData(playerId uuid.UUID) (LobbyGameBoardData, error) {
	var data LobbyGameBoardData

	sqlString := `
		SELECT
			L.ID AS LOBBY_ID,
			(SELECT TEXT FROM CARD WHERE ID = J.CARD_ID) AS JUDGE_CARD_TEXT,
			(SELECT YOUTUBE FROM CARD WHERE ID = J.CARD_ID) AS JUDGE_CARD_YOUTUBE,
			(SELECT IMAGE FROM CARD WHERE ID = J.CARD_ID) AS JUDGE_CARD_IMAGE,
			J.BLANK_COUNT AS JUDGE_BLANK_COUNT,
			J.RESPONSE_COUNT AS JUDGE_RESPONSE_COUNT,
			P.ID AS PLAYER_ID,
			IF(FN_GET_LOBBY_JUDGE_PLAYER_ID(L.ID) = P.ID, 1, 0) AS PLAYER_IS_JUDGE
		FROM PLAYER AS P
			INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = L.ID
		WHERE P.ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var imageBytes []byte
		if err := rows.Scan(
			&data.LobbyId,
			&data.JudgeCardText,
			&data.JudgeCardYouTube,
			&imageBytes,
			&data.JudgeBlankCount,
			&data.JudgeResponseCount,
			&data.PlayerId,
			&data.PlayerIsJudge,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}

		data.JudgeCardImage.Valid = imageBytes != nil
		if data.JudgeCardImage.Valid {
			data.JudgeCardImage.String = base64.StdEncoding.EncodeToString(imageBytes)
		}
	}

	sqlString = `
		SELECT
			R.ID AS RESPONSE_ID,
			R.IS_REVEALED AS IS_REVEALED,
			R.IS_RULEDOUT AS IS_RULEDOUT,
			P.ID AS PLAYER_ID,
			U.NAME AS PLAYER_USER_NAME
		FROM LOBBY AS L
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = L.ID
			INNER JOIN USER AS U ON U.ID = P.USER_ID
			INNER JOIN RESPONSE AS R ON R.PLAYER_ID = P.ID
			LEFT JOIN JUDGE AS J ON J.PLAYER_ID = P.ID
		WHERE L.ID = ?
			AND P.IS_ACTIVE = 1
			AND J.ID IS NULL
		ORDER BY U.NAME,
			R.CREATED_ON_DATE
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var br boardResponse
		if err := rows.Scan(
			&br.ResponseId,
			&br.IsRevealed,
			&br.IsRuledOut,
			&br.PlayerId,
			&br.PlayerUserName); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.BoardResponses = append(data.BoardResponses, br)
	}

	data.BoardHasAnyRevealed = false
	data.BoardIsAllRevealed = true
	data.BoardIsAllRuledOut = true
	totalCardsPlayedCount := 0
	for i, br := range data.BoardResponses {
		if br.IsRevealed {
			data.BoardHasAnyRevealed = true
		}

		if !br.IsRevealed {
			data.BoardIsAllRevealed = false
		}

		if !br.IsRuledOut {
			data.BoardIsAllRuledOut = false
		}

		sqlString = `
			SELECT
				RC.ID AS RESPONSE_CARD_ID,
				C.ID AS CARD_ID,
				C.TEXT AS CARD_TEXT,
				C.YOUTUBE AS CARD_YOUTUBE,
				C.IMAGE AS CARD_IMAGE,
				RC.SPECIAL_CATEGORY
			FROM RESPONSE AS R
				INNER JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
				INNER JOIN CARD AS C ON C.ID = RC.CARD_ID
			WHERE R.ID = ?
			ORDER BY RC.CREATED_ON_DATE
		`
		rows, err = query(sqlString, br.ResponseId)
		if err != nil {
			return data, err
		}
		defer rows.Close()

		for rows.Next() {
			var responseCard boardResponseCard
			var imageBytes []byte
			if err := rows.Scan(
				&responseCard.ResponseCardId,
				&responseCard.Id,
				&responseCard.Text,
				&responseCard.YouTube,
				&imageBytes,
				&responseCard.SpecialCategory,
			); err != nil {
				log.Println(err)
				return data, errors.New("failed to scan row in query results")
			}

			responseCard.Image.Valid = imageBytes != nil
			if responseCard.Image.Valid {
				responseCard.Image.String = base64.StdEncoding.EncodeToString(imageBytes)
			}

			data.BoardResponses[i].ResponseCards = append(data.BoardResponses[i].ResponseCards, responseCard)

			totalCardsPlayedCount += 1
		}

		if br.PlayerId == data.PlayerId {
			data.PlayerResponses = append(data.PlayerResponses, data.BoardResponses[i])
		}
	}

	sqlString = `
		SELECT
			P.BET_ON_WIN,
			P.EXTRA_RESPONSES,
			RC.SPECIAL_CATEGORY
		FROM PLAYER AS P
			LEFT JOIN RESPONSE AS R ON R.PLAYER_ID = P.ID
			LEFT JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
		WHERE P.LOBBY_ID = ?
			AND P.IS_ACTIVE = 1
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	data.BoardHasAnySpecial = false
	for rows.Next() {
		var betOnWin int
		var extraResponse int
		var specialCategory sql.NullString
		if err := rows.Scan(
			&betOnWin,
			&extraResponse,
			&specialCategory,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		if betOnWin > 0 || extraResponse > 0 || specialCategory.Valid {
			data.BoardHasAnySpecial = true
			break
		}
	}

	data.BoardIsReady = totalCardsPlayedCount > 0

	sqlString = `
		SELECT
			P.ID AS PLAYER_ID,
			IF(
				COUNT(RC.ID) = -- CARDS PLAYED
				(
					J.BLANK_COUNT * -- CARDS PER RESPONSE
					(J.RESPONSE_COUNT + P.EXTRA_RESPONSES) -- PLAYER RESPONSES
				),
				1,
				0
			) AS PLAYER_IS_READY
		FROM RESPONSE AS R
			LEFT JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
			INNER JOIN PLAYER AS P ON P.ID = R.PLAYER_ID
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = P.LOBBY_ID
		WHERE J.LOBBY_ID = ?
		GROUP BY P.ID
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var pId uuid.UUID
		var isReady bool
		if err := rows.Scan(&pId, &isReady); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}

		if !isReady {
			data.BoardIsReady = false
		}
	}

	if data.BoardIsReady {
		sort.Slice(data.BoardResponses, func(i, j int) bool {
			if len(data.BoardResponses[i].ResponseCards) == 0 {
				return true
			}
			if len(data.BoardResponses[j].ResponseCards) == 0 {
				return false
			}
			return data.BoardResponses[i].ResponseCards[0].Text < data.BoardResponses[j].ResponseCards[0].Text
		})
	}

	return data, nil
}

func GetLobbyGameStatsData(playerId uuid.UUID) (LobbyGameStatsData, error) {
	var data LobbyGameStatsData

	data.PlayerId = playerId
	lobbyId, err := GetPlayerLobbyId(playerId)
	if err != nil {
		return data, err
	}
	data.LobbyId = lobbyId

	sqlString := `
		SELECT
			U.NAME AS USER_NAME,
			COUNT(W.ID) AS WINS
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
			LEFT JOIN WIN AS W ON W.PLAYER_ID = P.ID
		WHERE P.LOBBY_ID = ?
			AND P.IS_ACTIVE = 1
		GROUP BY P.USER_ID
		ORDER BY COUNT(W.ID) DESC,
			U.NAME ASC
	`
	rows, err := query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var row nameCountRow
		if err := rows.Scan(&row.Name, &row.Count); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.Wins = append(data.Wins, row)
	}

	sqlString = `
		SELECT
			U.NAME
		FROM LOBBY AS L
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = L.ID
			INNER JOIN (
				SELECT
					LOBBY_ID,
					USER_ID,
					RANK() OVER (PARTITION BY LOBBY_ID ORDER BY CREATED_ON_DATE) AS JOIN_ORDER
				FROM PLAYER
				WHERE IS_ACTIVE = 1
			) AS T ON T.LOBBY_ID = L.ID
			INNER JOIN USER AS U ON U.ID = T.USER_ID
		WHERE L.ID = ?
		ORDER BY T.JOIN_ORDER <= J.POSITION,
			T.JOIN_ORDER
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var judgeName string
		if err := rows.Scan(&judgeName); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.UpcomingJudges = append(data.UpcomingJudges, judgeName)
	}

	sqlString = `
		SELECT
			P.ID AS PLAYER_ID,
			U.NAME AS USER_NAME,
			IF(
				EXISTS(
					SELECT
						ID
					FROM KICK
					WHERE VOTER_PLAYER_ID = ?
						AND SUBJECT_PLAYER_ID = P.ID
				),
				1,
				0
			) AS VOTED
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
			LEFT JOIN KICK AS K ON K.SUBJECT_PLAYER_ID = P.ID
		WHERE P.IS_ACTIVE = 1
			AND P.ID <> ?
			AND P.LOBBY_ID = ?
		ORDER BY U.NAME
	`
	rows, err = query(sqlString, data.PlayerId, data.PlayerId, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var row kickVote
		if err := rows.Scan(
			&row.PlayerId,
			&row.UserName,
			&row.Voted); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.KickVotes = append(data.KickVotes, row)
	}

	return data, nil
}

func PlayCard(playerId uuid.UUID, cardId uuid.UUID) error {
	sqlString := "CALL SP_RESPOND_WITH_CARD (?, ?, NULL)"
	return execute(sqlString, playerId, cardId)
}

func PurchaseCredits(playerId uuid.UUID) error {
	sqlString := "CALL SP_PURCHASE_CREDITS (?)"
	return execute(sqlString, playerId)
}

func SkipJudge(playerId uuid.UUID) error {
	sqlString := "CALL SP_SKIP_JUDGE (?)"
	return execute(sqlString, playerId)
}

func ResetResponses(playerId uuid.UUID) error {
	sqlString := "CALL SP_RESET_RESPONSES (?)"
	return execute(sqlString, playerId)
}

func AlertLobby(playerId uuid.UUID, credits int) error {
	sqlString := "CALL SP_ALERT_LOBBY (?, ?)"
	return execute(sqlString, playerId, credits)
}

func GambleCredits(playerId uuid.UUID, credits int) (bool, error) {
	var gambleWon bool
	sqlString := "CALL SP_GAMBLE_CREDITS (?, ?)"
	rows, err := query(sqlString, playerId, credits)
	if err != nil {
		return gambleWon, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&gambleWon); err != nil {
			log.Println(err)
			return gambleWon, errors.New("failed to scan row in query results")
		}
	}

	return gambleWon, nil
}

func BetOnWin(playerId uuid.UUID, credits int) error {
	sqlString := "CALL SP_BET_ON_WIN (?, ?)"
	return execute(sqlString, playerId, credits)
}

func BetOnWinUndo(playerId uuid.UUID) error {
	sqlString := "CALL SP_BET_ON_WIN_UNDO (?)"
	return execute(sqlString, playerId)
}

func AddExtraResponse(playerId uuid.UUID) error {
	sqlString := "CALL SP_ADD_EXTRA_RESPONSE (?)"
	return execute(sqlString, playerId)
}

func AddExtraResponseUndo(playerId uuid.UUID) error {
	sqlString := "CALL SP_ADD_EXTRA_RESPONSE_UNDO (?)"
	return execute(sqlString, playerId)
}

func BlockResponse(playerId uuid.UUID, targetPlayerId uuid.UUID) error {
	sqlString := "CALL SP_BLOCK_RESPONSE (?, ?)"
	return execute(sqlString, playerId, targetPlayerId)
}

func PlaySurpriseCard(playerId uuid.UUID) error {
	sqlString := "CALL SP_RESPOND_WITH_SURPRISE_CARD (?)"
	return execute(sqlString, playerId)
}

func PlayStealCard(playerId uuid.UUID) error {
	sqlString := "CALL SP_RESPOND_WITH_STEAL_CARD (?)"
	return execute(sqlString, playerId)
}

func PlayFindCard(playerId uuid.UUID, cardId uuid.UUID) error {
	sqlString := "CALL SP_RESPOND_WITH_FIND_CARD (?, ?)"
	return execute(sqlString, playerId, cardId)
}

func PlayWildCard(playerId uuid.UUID, text string) error {
	sqlString := "CALL SP_RESPOND_WITH_WILD_CARD (?, ?)"
	return execute(sqlString, playerId, text)
}

func WithdrawCard(responseCardId uuid.UUID) error {
	sqlString := "CALL SP_WITHDRAW_CARD (?)"
	return execute(sqlString, responseCardId)
}

func DiscardCard(playerId uuid.UUID, cardId uuid.UUID) error {
	sqlString := "CALL SP_DISCARD_CARD (?, ?)"
	return execute(sqlString, playerId, cardId)
}

func VoteToKick(voterPlayerId uuid.UUID, subjectPlayerId uuid.UUID) (bool, error) {
	var isKicked bool
	sqlString := "CALL SP_VOTE_TO_KICK (?, ?)"
	rows, err := query(sqlString, voterPlayerId, subjectPlayerId)
	if err != nil {
		return isKicked, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&isKicked); err != nil {
			log.Println(err)
			return isKicked, errors.New("failed to scan row in query results")
		}
	}

	return isKicked, nil
}

func VoteToKickUndo(voterPlayerId uuid.UUID, subjectPlayerId uuid.UUID) error {
	sqlString := "CALL SP_VOTE_TO_KICK_UNDO (?, ?)"
	return execute(sqlString, voterPlayerId, subjectPlayerId)
}

func RevealResponse(responseId uuid.UUID) error {
	sqlString := `
		UPDATE RESPONSE
		SET IS_REVEALED = 1
		WHERE ID = ?
	`
	return execute(sqlString, responseId)
}

func ToggleRuleOutResponse(responseId uuid.UUID) error {
	sqlString := `
		UPDATE RESPONSE
		SET IS_RULEDOUT = !IS_RULEDOUT
		WHERE ID = ?
	`
	return execute(sqlString, responseId)
}

func PickWinner(responseId uuid.UUID) (string, error) {
	var playerName string
	sqlString := "CALL SP_PICK_WINNER (?)"
	rows, err := query(sqlString, responseId)
	if err != nil {
		return playerName, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&playerName); err != nil {
			log.Println(err)
			return playerName, errors.New("failed to scan row in query results")
		}
	}

	return playerName, nil
}

func PickRandomWinner(lobbyId uuid.UUID) (string, error) {
	var playerName string
	sqlString := "CALL SP_PICK_RANDOM_WINNER (?)"
	rows, err := query(sqlString, lobbyId)
	if err != nil {
		return playerName, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&playerName); err != nil {
			log.Println(err)
			return playerName, errors.New("failed to scan row in query results")
		}
	}

	return playerName, nil
}

func SkipPrompt(lobbyId uuid.UUID) error {
	sqlString := "CALL SP_SKIP_PROMPT (?)"
	return execute(sqlString, lobbyId)
}

func SetJudgeResponseCount(lobbyId uuid.UUID, responseCount int) error {
	sqlString := "CALL SP_SET_RESPONSE_COUNT (?, ?)"
	return execute(sqlString, lobbyId, responseCount)
}
