document.addEventListener("htmx:afterSwap", function (event) {
    if (event.target.id === "deck-card-export-content") {
        downloadCSV("card-judge-cards.csv", event.target.innerHTML);
    }
});

function downloadCSV(fileName, content) {
    const element = document.createElement("a");
    element.setAttribute("href", "data:text/csv;charset=utf-8," + encodeURIComponent(content));
    element.setAttribute("download", fileName);
    element.style.display = "none";
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
}
