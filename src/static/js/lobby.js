let wsReconnectAttempts = 0;

window.onload = () => {
    const lobbyMessageDialog = document.getElementById("lobby-message-dialog");
    if (lobbyMessageDialog) lobbyMessageDialog.showModal();

    websocketConnect();
};

function websocketConnect() {
    let wsProtocol = "wss://";
    if (document.location.protocol === "http:") {
        wsProtocol = "ws://";
    }

    const ws = new WebSocket(wsProtocol + document.location.host + "/ws" + document.location.pathname);

    if (!ws) {
        alert("Failed to make connection.");
        document.location.href = "/lobbies";
    }

    ws.onopen = () => {
        if (wsReconnectAttempts > 0) {
            displayLobbyAlert("Connection Restored", `Restored on attempt ${wsReconnectAttempts}`, 3);
        }
        wsReconnectAttempts = 0;
    };

    ws.onclose = () => {
        if (wsReconnectAttempts < 3) {
            wsReconnectAttempts++;
            displayLobbyAlert("Connection Lost", `Attempting to reconnect (${wsReconnectAttempts}/3)...`, 5);
            setTimeout(() => { websocketConnect() }, 5000);
        } else {
            alert("Connection Lost");
            document.location.href = "/lobbies";
        }
    };

    const lobbyChatForm = document.getElementById("lobby-chat-form");
    const lobbyChatMessages = document.getElementById("lobby-chat-messages");
    const lobbyChatInput = document.getElementById("lobby-chat-input");

    gsChat.wireForm(lobbyChatForm, lobbyChatInput, ws);

    ws.onmessage = (event) => {
        let messageText = event.data;

        switch (messageText) {
            case "refresh":
                confirmationDialogDelete();
                htmx.ajax("GET", "/api" + document.location.pathname + "/html/game-interface", {
                    source: "#lobby-grid-interface",
                    target: "#lobby-grid-interface"
                });
                resetRoundTimerInterval();
                return;

            case "refresh-lobby-game-info":
                confirmationDialogDelete();
                htmx.ajax("GET", "/api" + document.location.pathname + "/html/lobby-game-info", {
                    source: "#lobby-game-info",
                    target: "#lobby-game-info"
                });
                return;

            case "refresh-player-hand":
                confirmationDialogDelete();
                htmx.ajax("GET", "/api" + document.location.pathname + "/html/player-hand", {
                    source: "#player-hand",
                    target: "#player-hand"
                });
                return;

            case "refresh-player-specials":
                confirmationDialogDelete();
                htmx.ajax("GET", "/api" + document.location.pathname + "/html/player-specials", {
                    source: "#player-specials",
                    target: "#player-specials"
                });
                return;

            case "refresh-lobby-game-board":
                confirmationDialogDelete();
                htmx.ajax("GET", "/api" + document.location.pathname + "/html/lobby-game-board", {
                    source: "#lobby-game-board",
                    target: "#lobby-game-board"
                });
                return;

            case "refresh-lobby-game-stats":
                confirmationDialogDelete();
                htmx.ajax("GET", "/api" + document.location.pathname + "/html/lobby-game-stats", {
                    source: "#lobby-game-stats",
                    target: "#lobby-game-stats"
                });
                return;

            case "table-flipped":
            case "player-kicked":
                confirmationDialogDelete();
                const gifDialog = document.getElementById(`${messageText}-dialog`);
                if (gifDialog) {
                    gifDialog.showModal();
                    setTimeout(() => gifDialog.close(), 2000);
                }
                return;

            case "exit":
                document.location.href = "/lobbies";
                return;
        }

        if (messageText.startsWith("timer")) {
            const timerData = messageText.split(";;");
            if (timerData.length === 2) {
                startRoundTimerInterval(timerData[1]);
            }
            return;
        }

        if (messageText.startsWith("alert")) {
            const alertData = messageText.split(";;");
            if (alertData.length === 4) {
                displayLobbyAlert(alertData[2], alertData[3], alertData[1]);
            }
            return;
        }

        // Shared renderer: color tokens + timestamp + history trim
        // (see gameshell-framework /gs/js/chat.js).
        gsChat.append(lobbyChatMessages, messageText);
    };
}

function displayLobbyAlert(header, body, seconds) {
    const alertHeader = document.getElementById("lobby-alert-dialog-header");
    if (!alertHeader) return;
    const alertBody = document.getElementById("lobby-alert-dialog-body");
    if (!alertBody) return;
    const alertDialog = document.getElementById("lobby-alert-dialog");
    if (!alertDialog) return;

    alertHeader.innerText = header;
    alertBody.innerText = body;
    alertDialog.showModal();
    setTimeout(() => alertDialog.close(), seconds * 1000);
}

let roundTimerInterval = null;

resetRoundTimerInterval();

function resetRoundTimerInterval(seconds = 0) {
    if (roundTimerInterval) clearInterval(roundTimerInterval);

    const roundTimerSecondsElement = document.getElementById("round-timer-seconds");
    if (!roundTimerSecondsElement) return;
    roundTimerSecondsElement.innerText = seconds;
}

function startRoundTimerInterval(seconds) {
    resetRoundTimerInterval(seconds);
    roundTimerInterval = setInterval(() => {
        const roundTimerElement = document.getElementById("round-timer");
        if (!roundTimerElement) return;

        const roundTimerSecondsElement = document.getElementById("round-timer-seconds");
        if (!roundTimerSecondsElement) return;

        let secondsRemaining = parseInt(roundTimerSecondsElement.innerText) || 0;
        secondsRemaining -= 1;

        roundTimerSecondsElement.innerText = secondsRemaining < 0 ? 0 : secondsRemaining;

        switch (true) {
            case (secondsRemaining >= 6 && secondsRemaining < 10):
                roundTimerElement.className = "red-text";
                break;
            case (secondsRemaining >= 0 && secondsRemaining < 6):
                roundTimerElement.className = "red-text pulse-fast";
                break;
            default:
                roundTimerElement.className = "";
                break;
        }

        if (secondsRemaining === 0) {
            fetch("/api" + document.location.pathname + "/card/force/play", { method: "POST" });
        }
    }, 1000);
}

let lobbyPlayerDataScrollTop = 0;
let lobbyGameBoardScrollTop = 0;
let lobbyGameStatsScrollTop = 0;

document.addEventListener("htmx:beforeSwap", function () {
    const lobbyPlayerData = document.getElementById("lobby-player-data");
    if (lobbyPlayerData) lobbyPlayerDataScrollTop = lobbyPlayerData.scrollTop;
    const lobbyGameBoard = document.getElementById("lobby-game-board");
    if (lobbyGameBoard) lobbyGameBoardScrollTop = lobbyGameBoard.scrollTop;
    const lobbyGameStats = document.getElementById("lobby-game-stats");
    if (lobbyGameStats) lobbyGameStatsScrollTop = lobbyGameStats.scrollTop;
});

document.addEventListener("htmx:afterSwap", function () {
    const lobbyPlayerData = document.getElementById("lobby-player-data");
    if (lobbyPlayerData) lobbyPlayerData.scrollTop = lobbyPlayerDataScrollTop;
    const lobbyGameBoard = document.getElementById("lobby-game-board");
    if (lobbyGameBoard) lobbyGameBoard.scrollTop = lobbyGameBoardScrollTop;
    const lobbyGameStats = document.getElementById("lobby-game-stats");
    if (lobbyGameStats) lobbyGameStats.scrollTop = lobbyGameStatsScrollTop;
});
