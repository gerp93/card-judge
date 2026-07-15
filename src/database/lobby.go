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
	Message      sql.NullString
	PasswordHash sql.NullString

	DrawPriority        string
	HandSize            int
	RoundTimer          int
	FreeCredits         int
	FreeSpecialCards    bool
	WinStreakThreshold  int
	LoseStreakThreshold int
}

type LobbyDetails struct {
	Lobby
	UserCount int
}

type LobbyGameInfo struct {
	LobbyName string

	PlayerIsLobbyOwner bool

	JudgeName sql.NullString

	RoundTimer int

	DrawPilePromptCount   int
	DrawPileResponseCount int
	DrawPileDeckNames     string
}

type PlayerHandData struct {
	LobbyId uuid.UUID

	PlayerId               uuid.UUID
	PlayerIsJudge          bool
	PlayerDiscardAdvantage bool
	PlayerIsReady          bool
	PlayerHand             []Card
}

type PlayerSpecialsData struct {
	LobbyId                  uuid.UUID
	LobbyFreeCredits         int
	LobbyFreeSpecialCards    bool
	LobbyWinStreakThreshold  int
	LobbyLoseStreakThreshold int

	BoardHasAnySpecial  bool
	BoardHasAnyRevealed bool
	BoardResponses      []boardResponse

	Opponents []opponentData

	PlayerId                uuid.UUID
	PlayerIsJudge           bool
	PlayerIsReady           bool
	PlayerHandicap          int
	PlayerWinningStreak     int
	PlayerLosingStreak      int
	PlayerCreditsSpent      int
	PlayerBetOnWin          int
	PlayerExtraResponses    int
	PlayerDiscardAdvantage  bool
	PlayerHandicapAdvantage bool
	PlayerSpyAdvantage      bool
	PlayerCreditsRemaining  int

	CreditHistory []creditHistoryData

	SpecialCostSkipBeingJudge int
	SpecialCostExtraResponse  int
	SpecialCostBlockResponse  int
	SpecialCostSurpriseCard   int
	SpecialCostStealCard      int
	SpecialCostFindCard       int
	SpecialCostWildCard       int
	SpecialCostPerk           int

	CannotAffordAlert          bool
	CannotAffordGamble         bool
	CannotAffordBet            bool
	CannotAffordSkipBeingJudge bool
	CannotAffordExtraResponse  bool
	CannotAffordBlockResponse  bool
	CannotAffordSurpriseCard   bool
	CannotAffordStealCard      bool
	CannotAffordFindCard       bool
	CannotAffordWildCard       bool
	CannotAffordPerk           bool
}

type LobbyGameBoardData struct {
	LobbyId uuid.UUID

	JudgeCardText      sql.NullString
	JudgeCardYouTube   sql.NullString
	JudgeCardImage     sql.NullString
	JudgeBlankCount    int
	JudgeResponseCount int

	RoundTimer int

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

	PlayerId           uuid.UUID
	PlayerSpyAdvantage bool

	Wins           []nameCountRow
	Credits        []nameCountRow
	UpcomingJudges []string
	KickVotes      []kickVote
}

type opponentData struct {
	PlayerId uuid.UUID
	UserName string
}

type creditHistoryData struct {
	Category      string
	BalanceChange int
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
			CJLS.DRAW_PRIORITY,
			CJLS.HAND_SIZE,
			CJLS.ROUND_TIMER,
			CJLS.FREE_CREDITS,
			CJLS.FREE_SPECIAL_CARDS,
			CJLS.WIN_STREAK_THRESHOLD,
			CJLS.LOSE_STREAK_THRESHOLD,
			COUNT(P.ID) AS USER_COUNT
		FROM LOBBY AS L
			INNER JOIN CJ_LOBBY_SETTINGS AS CJLS ON CJLS.LOBBY_ID = L.ID
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = L.ID
			AND P.IS_ACTIVE = 1
		WHERE L.NAME LIKE ?
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
			&ld.RoundTimer,
			&ld.FreeCredits,
			&ld.FreeSpecialCards,
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
			L.ID,
			L.CREATED_ON_DATE,
			L.NAME,
			L.MESSAGE,
			L.PASSWORD_HASH,
			CJLS.DRAW_PRIORITY,
			CJLS.HAND_SIZE,
			CJLS.ROUND_TIMER,
			CJLS.FREE_CREDITS,
			CJLS.FREE_SPECIAL_CARDS,
			CJLS.WIN_STREAK_THRESHOLD,
			CJLS.LOSE_STREAK_THRESHOLD
		FROM LOBBY AS L
			INNER JOIN CJ_LOBBY_SETTINGS AS CJLS ON CJLS.LOBBY_ID = L.ID
		WHERE L.ID = ?
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
			&lobby.Message,
			&lobby.PasswordHash,
			&lobby.DrawPriority,
			&lobby.HandSize,
			&lobby.RoundTimer,
			&lobby.FreeCredits,
			&lobby.FreeSpecialCards,
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

func CreateLobby(name string, message string, password string, drawPriority string, handSize int, roundTimer int, freeCredits int, freeSpecialCards bool, winStreakThreshold int, loseStreakThreshold int) (uuid.UUID, error) {
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
			PASSWORD_HASH
		)
		VALUES (?, ?, ?, ?)
	`
	if message == "" {
		if password == "" {
			err = execute(sqlString, id, name, nil, nil)
		} else {
			err = execute(sqlString, id, name, nil, passwordHash)
		}
	} else {
		if password == "" {
			err = execute(sqlString, id, name, message, nil)
		} else {
			err = execute(sqlString, id, name, message, passwordHash)
		}
	}
	if err != nil {
		return id, err
	}

	sqlString = `
		UPDATE CJ_LOBBY_SETTINGS
		SET DRAW_PRIORITY = ?,
			HAND_SIZE = ?,
			ROUND_TIMER = ?,
			FREE_CREDITS = ?,
			FREE_SPECIAL_CARDS = ?,
			WIN_STREAK_THRESHOLD = ?,
			LOSE_STREAK_THRESHOLD = ?
		WHERE LOBBY_ID = ?
	`
	return id, execute(sqlString, drawPriority, handSize, roundTimer, freeCredits, freeSpecialCards, winStreakThreshold, loseStreakThreshold, id)
}

func SyncDecksInLobby(lobbyId uuid.UUID, deckIdsPrompt []uuid.UUID, deckIdsResponse []uuid.UUID) error {
	if len(deckIdsPrompt) == 0 {
		return errors.New("cannot sync decks in lobby, no prompt deck ids provided")
	}

	if len(deckIdsResponse) == 0 {
		return errors.New("cannot sync decks in lobby, no response deck ids provided")
	}

	var err error

	err = addDecksToLobby(lobbyId, deckIdsPrompt, "PROMPT")
	if err != nil {
		return err
	}

	err = addDecksToLobby(lobbyId, deckIdsResponse, "RESPONSE")
	if err != nil {
		return err
	}

	err = removeDecksFromLobby(lobbyId, deckIdsPrompt, "PROMPT")
	if err != nil {
		return err
	}

	err = removeDecksFromLobby(lobbyId, deckIdsResponse, "RESPONSE")
	if err != nil {
		return err
	}

	return nil
}

func addDecksToLobby(lobbyId uuid.UUID, deckIds []uuid.UUID, cardCategory string) error {
	sqlString := fmt.Sprintf(`
		INSERT INTO DRAW_PILE(LOBBY_ID, CARD_ID)
		SELECT
			? AS LOBBY_ID,
			C.ID AS CARD_ID
		FROM CARD AS C
			LEFT JOIN (
				SELECT DISTINCT
					E_C.DECK_ID
				FROM DRAW_PILE AS E_DP
					INNER JOIN CARD AS E_C ON E_C.ID = E_DP.CARD_ID
				WHERE E_DP.LOBBY_ID = ?
					AND E_C.CATEGORY = ?
			) AS E ON E.DECK_ID = C.DECK_ID
		WHERE C.CATEGORY = ?
			AND C.DECK_ID IN (%s)
			AND E.DECK_ID IS NULL
	`, strings.Repeat("?,", len(deckIds)-1)+"?")

	args := make([]any, len(deckIds)+4)
	args[0] = lobbyId
	args[1] = lobbyId
	args[2] = cardCategory
	args[3] = cardCategory
	for i, deckId := range deckIds {
		args[i+4] = deckId
	}

	err := execute(sqlString, args...)
	if err != nil {
		return err
	}

	return nil
}

func removeDecksFromLobby(lobbyId uuid.UUID, deckIds []uuid.UUID, cardCategory string) error {
	sqlString := fmt.Sprintf(`
		DELETE DP
		FROM DRAW_PILE AS DP
			INNER JOIN CARD AS C ON C.ID = DP.CARD_ID
		WHERE DP.LOBBY_ID = ?
			AND C.CATEGORY = ?
			AND C.DECK_ID NOT IN (%s)
	`, strings.Repeat("?,", len(deckIds)-1)+"?")

	args := make([]any, len(deckIds)+2)
	args[0] = lobbyId
	args[1] = cardCategory
	for i, deckId := range deckIds {
		args[i+2] = deckId
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
		UPDATE CJ_LOBBY_SETTINGS
		SET DRAW_PRIORITY = ?
		WHERE LOBBY_ID = ?
	`
	return execute(sqlString, drawPriority, id)
}

func SetLobbyHandSize(id uuid.UUID, handSize int) error {
	sqlString := `
		UPDATE CJ_LOBBY_SETTINGS
		SET HAND_SIZE = ?
		WHERE LOBBY_ID = ?
	`
	return execute(sqlString, handSize, id)
}

func SetLobbyRoundTimer(id uuid.UUID, roundTimer int) error {
	sqlString := `
		UPDATE CJ_LOBBY_SETTINGS
		SET ROUND_TIMER = ?
		WHERE LOBBY_ID = ?
	`
	return execute(sqlString, roundTimer, id)
}

func SetLobbyFreeCredits(id uuid.UUID, freeCredits int) error {
	sqlString := `
		UPDATE CJ_LOBBY_SETTINGS
		SET FREE_CREDITS = ?
		WHERE LOBBY_ID = ?
	`
	return execute(sqlString, freeCredits, id)
}

func SetLobbyFreeSpecialCards(id uuid.UUID, freeSpecialCards bool) error {
	sqlString := `
		UPDATE CJ_LOBBY_SETTINGS
		SET FREE_SPECIAL_CARDS = ?
		WHERE LOBBY_ID = ?
	`
	return execute(sqlString, freeSpecialCards, id)
}

func SetLobbyWinStreakThreshold(id uuid.UUID, winStreakThreshold int) error {
	sqlString := `
		UPDATE CJ_LOBBY_SETTINGS
		SET WIN_STREAK_THRESHOLD = ?
		WHERE LOBBY_ID = ?
	`
	return execute(sqlString, winStreakThreshold, id)
}

func SetLobbyLoseStreakThreshold(id uuid.UUID, loseStreakThreshold int) error {
	sqlString := `
		UPDATE CJ_LOBBY_SETTINGS
		SET LOSE_STREAK_THRESHOLD = ?
		WHERE LOBBY_ID = ?
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

func GetLobbyGameInfo(lobbyId uuid.UUID, playerId uuid.UUID) (LobbyGameInfo, error) {
	var data LobbyGameInfo

	sqlString := `
		SELECT
			L.NAME AS LOBBY_NAME,
			(
				SELECT
					P.ID
				FROM PLAYER AS P
				WHERE P.LOBBY_ID = L.ID
					AND P.IS_ACTIVE = 1
				ORDER BY P.JOIN_ORDER ASC
				LIMIT 1
			) AS LOBBY_OWNER_ID,
			(
				SELECT
					U.NAME
				FROM USER AS U
					INNER JOIN PLAYER AS P ON P.USER_ID = U.ID
					INNER JOIN JUDGE AS J ON J.PLAYER_ID = P.ID
				WHERE J.LOBBY_ID = L.ID
				LIMIT 1
			) AS JUDGE_NAME,
			(
				SELECT
					ROUND_TIMER
				FROM CJ_LOBBY_SETTINGS
				WHERE LOBBY_ID = L.ID
			) AS ROUND_TIMER,
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
			) AS DRAW_PILE_RESPONSE_COUNT
		FROM LOBBY AS L
		WHERE L.ID = ?
	`
	rows, err := query(sqlString, lobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var lobbyOwnerId uuid.UUID
		if err := rows.Scan(
			&data.LobbyName,
			&lobbyOwnerId,
			&data.JudgeName,
			&data.RoundTimer,
			&data.DrawPilePromptCount,
			&data.DrawPileResponseCount,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}

		data.PlayerIsLobbyOwner = playerId == lobbyOwnerId
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
			IF(FN_GET_LOBBY_JUDGE_PLAYER_ID(L.ID) = P.ID, 1, 0) AS PLAYER_IS_JUDGE,
			(
				SELECT
					DISCARD_ADVANTAGE
				FROM CJ_PLAYER_STATE
				WHERE PLAYER_ID = P.ID
			) AS PLAYER_DISCARD_ADVANTAGE
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
			&data.PlayerDiscardAdvantage,
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
					(J.RESPONSE_COUNT + CJPS.EXTRA_RESPONSES) -- PLAYER RESPONSES
				),
				1,
				0
			) AS PLAYER_IS_READY
		FROM PLAYER AS P
			INNER JOIN CJ_PLAYER_STATE AS CJPS ON CJPS.PLAYER_ID = P.ID
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
			CJLS.FREE_CREDITS AS LOBBY_FREE_CREDITS,
			CJLS.FREE_SPECIAL_CARDS AS LOBBY_FREE_SPECIAL_CARDS,
			CJLS.WIN_STREAK_THRESHOLD AS LOBBY_WIN_STREAK_THRESHOLD,
			CJLS.LOSE_STREAK_THRESHOLD AS LOBBY_LOSE_STREAK_THRESHOLD,
			P.ID AS PLAYER_ID,
			IF(FN_GET_LOBBY_JUDGE_PLAYER_ID(L.ID) = P.ID, 1, 0) AS PLAYER_IS_JUDGE,
			FN_GET_PLAYER_HANDICAP(P.ID) AS PLAYER_HANDICAP,
			CJPS.WINNING_STREAK AS PLAYER_WINNING_STREAK,
			CJPS.LOSING_STREAK AS PLAYER_LOSING_STREAK,
			CJPS.CREDITS_SPENT AS PLAYER_CREDITS_SPENT,
			CJPS.BET_ON_WIN AS PLAYER_BET_ON_WIN,
			CJPS.EXTRA_RESPONSES AS PLAYER_EXTRA_RESPONSES,
			CJPS.DISCARD_ADVANTAGE AS PLAYER_DISCARD_ADVANTAGE,
			CJPS.HANDICAP_ADVANTAGE AS PLAYER_HANDICAP_ADVANTAGE,
			CJPS.SPY_ADVANTAGE AS PLAYER_SPY_ADVANTAGE
		FROM PLAYER AS P
			INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
			INNER JOIN CJ_LOBBY_SETTINGS AS CJLS ON CJLS.LOBBY_ID = L.ID
			INNER JOIN CJ_PLAYER_STATE AS CJPS ON CJPS.PLAYER_ID = P.ID
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
			&data.LobbyFreeSpecialCards,
			&data.LobbyWinStreakThreshold,
			&data.LobbyLoseStreakThreshold,
			&data.PlayerId,
			&data.PlayerIsJudge,
			&data.PlayerHandicap,
			&data.PlayerWinningStreak,
			&data.PlayerLosingStreak,
			&data.PlayerCreditsSpent,
			&data.PlayerBetOnWin,
			&data.PlayerExtraResponses,
			&data.PlayerDiscardAdvantage,
			&data.PlayerHandicapAdvantage,
			&data.PlayerSpyAdvantage,
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
			1
		FROM PLAYER AS P
			INNER JOIN CREDITS_SPENT AS C ON C.PLAYER_ID = P.ID
		WHERE P.LOBBY_ID = ?
			AND P.IS_ACTIVE = 1
			AND C.CATEGORY IN (
				'BET',
				'EXTRA-RESPONSE',
				'BLOCK-RESPONSE',
				'STEAL',
				'SURPRISE',
				'FIND',
				'WILD'
			)
		LIMIT 1
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()
	data.BoardHasAnySpecial = rows.Next()

	sqlString = `
		SELECT
			IF(
				R.ID IS NULL -- CANNOT PLAY
				OR COUNT(RC.ID) = -- CARDS PLAYED
				(
					J.BLANK_COUNT * -- CARDS PER RESPONSE
					(J.RESPONSE_COUNT + CJPS.EXTRA_RESPONSES) -- PLAYER RESPONSES
				),
				1,
				0
			) AS PLAYER_IS_READY
		FROM PLAYER AS P
			INNER JOIN CJ_PLAYER_STATE AS CJPS ON CJPS.PLAYER_ID = P.ID
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

	sqlString = `
		SELECT
			LCS.CATEGORY,
			LCS.AMOUNT * -1 AS BALANCE_CHANGE
		FROM LOG_CREDITS_SPENT AS LCS
			INNER JOIN PLAYER AS P ON P.USER_ID = LCS.USER_ID
		WHERE P.ID = ?
			AND LCS.CREATED_ON_DATE >= DATE_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 HOUR)
		ORDER BY LCS.CREATED_ON_DATE DESC
		LIMIT 10
	`
	rows, err = query(sqlString, playerId)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var row creditHistoryData
		if err := rows.Scan(
			&row.Category,
			&row.BalanceChange,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.CreditHistory = append(data.CreditHistory, row)
	}

	sqlString = `
		SELECT
			FN_GET_SPECIAL_COST('SKIP-JUDGE'),
			FN_GET_SPECIAL_COST('EXTRA-RESPONSE'),
			FN_GET_SPECIAL_COST('BLOCK-RESPONSE'),
			FN_GET_SPECIAL_COST('SURPRISE'),
			FN_GET_SPECIAL_COST('STEAL'),
			FN_GET_SPECIAL_COST('FIND'),
			FN_GET_SPECIAL_COST('WILD'),
			FN_GET_SPECIAL_COST('PERK')
	`
	rows, err = query(sqlString)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&data.SpecialCostSkipBeingJudge,
			&data.SpecialCostExtraResponse,
			&data.SpecialCostBlockResponse,
			&data.SpecialCostSurpriseCard,
			&data.SpecialCostStealCard,
			&data.SpecialCostFindCard,
			&data.SpecialCostWildCard,
			&data.SpecialCostPerk,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
	}

	data.SpecialCostSkipBeingJudge += data.PlayerHandicap
	data.SpecialCostExtraResponse += data.PlayerHandicap
	data.SpecialCostBlockResponse += data.PlayerHandicap
	data.SpecialCostSurpriseCard += data.PlayerHandicap
	data.SpecialCostStealCard += data.PlayerHandicap
	data.SpecialCostFindCard += data.PlayerHandicap
	data.SpecialCostWildCard += data.PlayerHandicap
	data.SpecialCostPerk += data.PlayerHandicap

	data.CannotAffordAlert = data.PlayerCreditsRemaining < 1
	data.CannotAffordGamble = data.PlayerCreditsRemaining < 1
	data.CannotAffordBet = data.PlayerCreditsRemaining < 1
	data.CannotAffordSkipBeingJudge = data.PlayerCreditsRemaining < data.SpecialCostSkipBeingJudge
	data.CannotAffordExtraResponse = data.PlayerCreditsRemaining < data.SpecialCostExtraResponse
	data.CannotAffordBlockResponse = data.PlayerCreditsRemaining < data.SpecialCostBlockResponse
	data.CannotAffordSurpriseCard = data.PlayerCreditsRemaining < data.SpecialCostSurpriseCard
	data.CannotAffordStealCard = data.PlayerCreditsRemaining < data.SpecialCostStealCard
	data.CannotAffordFindCard = data.PlayerCreditsRemaining < data.SpecialCostFindCard
	data.CannotAffordWildCard = data.PlayerCreditsRemaining < data.SpecialCostWildCard
	data.CannotAffordPerk = data.PlayerCreditsRemaining < data.SpecialCostPerk

	if data.LobbyFreeSpecialCards {
		data.SpecialCostSurpriseCard = 0
		data.SpecialCostStealCard = 0
		data.SpecialCostFindCard = 0
		data.SpecialCostWildCard = 0

		data.CannotAffordSurpriseCard = false
		data.CannotAffordStealCard = false
		data.CannotAffordFindCard = false
		data.CannotAffordWildCard = false
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
			(
				SELECT
					ROUND_TIMER
				FROM CJ_LOBBY_SETTINGS
				WHERE LOBBY_ID = L.ID
			) AS ROUND_TIMER,
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
			&data.RoundTimer,
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
			1
		FROM PLAYER AS P
			INNER JOIN CREDITS_SPENT AS C ON C.PLAYER_ID = P.ID
		WHERE P.LOBBY_ID = ?
			AND P.IS_ACTIVE = 1
			AND C.CATEGORY IN (
				'BET',
				'EXTRA-RESPONSE',
				'BLOCK-RESPONSE',
				'STEAL',
				'SURPRISE',
				'FIND',
				'WILD'
			)
		LIMIT 1
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}
	defer rows.Close()
	data.BoardHasAnySpecial = rows.Next()

	data.BoardIsReady = totalCardsPlayedCount > 0

	sqlString = `
		SELECT
			P.ID AS PLAYER_ID,
			IF(
				COUNT(RC.ID) = -- CARDS PLAYED
				(
					J.BLANK_COUNT * -- CARDS PER RESPONSE
					(J.RESPONSE_COUNT + CJPS.EXTRA_RESPONSES) -- PLAYER RESPONSES
				),
				1,
				0
			) AS PLAYER_IS_READY
		FROM RESPONSE AS R
			LEFT JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
			INNER JOIN PLAYER AS P ON P.ID = R.PLAYER_ID
			INNER JOIN CJ_PLAYER_STATE AS CJPS ON CJPS.PLAYER_ID = P.ID
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

	sqlString := `
		SELECT
			P.LOBBY_ID AS LOBBY_ID,
			P.ID AS PLAYER_ID,
			(
				SELECT
					SPY_ADVANTAGE
				FROM CJ_PLAYER_STATE
				WHERE PLAYER_ID = P.ID
			) AS PLAYER_SPY_ADVANTAGE
		FROM PLAYER AS P
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
			&data.PlayerSpyAdvantage,
		); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
	}

	sqlString = `
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
	rows, err = query(sqlString, data.LobbyId)
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

	if data.PlayerSpyAdvantage {
		sqlString = `
			SELECT
				U.NAME AS USER_NAME,
				CJLS.FREE_CREDITS - CJPS.CREDITS_SPENT AS CREDITS
			FROM PLAYER AS P
				INNER JOIN USER AS U ON U.ID = P.USER_ID
				INNER JOIN CJ_LOBBY_SETTINGS AS CJLS ON CJLS.LOBBY_ID = P.LOBBY_ID
				INNER JOIN CJ_PLAYER_STATE AS CJPS ON CJPS.PLAYER_ID = P.ID
			WHERE P.LOBBY_ID = ?
				AND P.IS_ACTIVE = 1
			ORDER BY CREDITS DESC,
				U.NAME ASC
		`
		rows, err = query(sqlString, data.LobbyId)
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
			data.Credits = append(data.Credits, row)
		}
	}

	sqlString = `
		SELECT
			U.NAME
		FROM PLAYER AS P
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = P.LOBBY_ID
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE P.LOBBY_ID = ?
			AND P.IS_ACTIVE = 1
			AND P.ID <> J.PLAYER_ID
		ORDER BY P.JOIN_ORDER <= J.POSITION,
			P.JOIN_ORDER
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

func PlayForceCard(playerId uuid.UUID) (bool, error) {
	var cardWasPlayed bool
	sqlString := "CALL SP_RESPOND_WITH_FORCE_CARD (?)"
	rows, err := query(sqlString, playerId)
	if err != nil {
		return cardWasPlayed, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&cardWasPlayed); err != nil {
			log.Println(err)
			return cardWasPlayed, errors.New("failed to scan row in query results")
		}
	}

	return cardWasPlayed, nil
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

func PerkHandSizeAdvantage(playerId uuid.UUID) error {
	sqlString := "CALL SP_PERK_HAND_SIZE_ADVANTAGE (?)"
	return execute(sqlString, playerId)
}

func PerkDiscardAdvantage(playerId uuid.UUID) error {
	sqlString := "CALL SP_PERK_DISCARD_ADVANTAGE (?)"
	return execute(sqlString, playerId)
}

func PerkHandicapAdvantage(playerId uuid.UUID) error {
	sqlString := "CALL SP_PERK_HANDICAP_ADVANTAGE (?)"
	return execute(sqlString, playerId)
}

func PerkSpyAdvantage(playerId uuid.UUID) error {
	sqlString := "CALL SP_PERK_SPY_ADVANTAGE (?)"
	return execute(sqlString, playerId)
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

func FlipTable(playerId uuid.UUID) error {
	sqlString := "CALL SP_FLIP_TABLE (?)"
	return execute(sqlString, playerId)
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
