window.onload = () => {
    const lobbyMessageDialog = document.getElementById("lobby-message-dialog");
    if (lobbyMessageDialog) lobbyMessageDialog.showModal();

    let wsProtocol = "wss://";
    if (document.location.protocol === "http:") {
        wsProtocol = "ws://";
    }

    const conn = new WebSocket(wsProtocol + document.location.host + "/ws" + document.location.pathname);

    if (!conn) {
        alert("Failed to make connection.");
        document.location.href = "/lobbies";
    }

    conn.onclose = () => {
        alert("Connection Lost");
        document.location.href = "/lobbies";
    };

    const lobbyChatForm = document.getElementById("lobby-chat-form");
    const lobbyChatMessages = document.getElementById("lobby-chat-messages");
    const lobbyChatInput = document.getElementById("lobby-chat-input");

    lobbyChatForm.onsubmit = (event) => {
        event.preventDefault();
        if (!lobbyChatInput.value) return;
        conn.send(lobbyChatInput.value);
        lobbyChatInput.value = "";
    };

    conn.onmessage = (event) => {
        let messageText = event.data;

        switch (messageText) {
            case "refresh":
                confirmationDialogDelete();
                htmx.ajax("GET", "/api" + document.location.pathname + "/html/game-interface", {
                    source: "#lobby-grid-interface",
                    target: "#lobby-grid-interface"
                });
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

            case "kick":
                document.location.href = "/lobbies";
                return;
        }

        if (messageText.startsWith("alert")) {
            const alertData = messageText.split(";;");
            if (alertData.length === 4) {
                const alertHeader = document.getElementById("lobby-alert-dialog-header");
                if (alertHeader) alertHeader.innerText = alertData[2];
                const alertBody = document.getElementById("lobby-alert-dialog-body");
                if (alertBody) alertBody.innerText = alertData[3];
                const alertDialog = document.getElementById("lobby-alert-dialog");
                if (alertDialog) {
                    alertDialog.showModal();
                    setTimeout(() => alertDialog.close(), alertData[1] * 2000);
                }
            }
            return;
        }

        messageText = messageText.replaceAll("<red>", '<span class="lobby-chat-message-red">');
        messageText = messageText.replaceAll("<green>", '<span class="lobby-chat-message-green">');
        messageText = messageText.replaceAll("<blue>", '<span class="lobby-chat-message-blue">');
        messageText = messageText.replaceAll("</>", "</span>");

        const now = new Date();
        messageText = now.getHours().toString().padStart(2, "0") + ":" + now.getMinutes().toString().padStart(2, "0") + " " + messageText;

        const message = document.createElement("div");
        message.innerHTML = messageText;
        lobbyChatMessages.appendChild(message);

        while (lobbyChatMessages.childNodes.length > 100) {
            lobbyChatMessages.removeChild(lobbyChatMessages.childNodes[0]);
        }

        lobbyChatMessages.scrollTop = lobbyChatMessages.scrollHeight - lobbyChatMessages.clientHeight;
    };
};

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
