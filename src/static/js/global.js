document.addEventListener("htmx:beforeOnLoad", function (event) {
    // always swap htmx response even if event.detail.xhr.status != 200
    event.detail.shouldSwap = true;
    event.detail.isError = false;
});

document.addEventListener("htmx:afterSwap", function (event) {
    if (event.detail.target.classList.contains("htmx-result")) {
        if (event.detail.xhr.status >= 200 && event.detail.xhr.status < 300) {
            addClassToTarget("htmx-result-good", event.detail.target);
            removeClassFromTarget("htmx-result-bad", event.detail.target);
        } else {
            addClassToTarget("htmx-result-bad", event.detail.target);
            removeClassFromTarget("htmx-result-good", event.detail.target);
        }
        // remove status message after 10 seconds
        setTimeout(() => {
            event.detail.target.innerHTML = "";
        }, 10000);
    }
});

function addClassToTarget(className, targetElement) {
    if (!targetElement.classList.contains(className)) {
        targetElement.classList.add(className);
    }
}

function removeClassFromTarget(className, targetElement) {
    if (targetElement.classList.contains(className)) {
        targetElement.classList.remove(className);
    }
}

function toggleTopBarMenu() {
    const topBarMenu = document.getElementById("top-bar-menu");
    if (topBarMenu.style.display === "block") {
        topBarMenu.style.display = "none";
    } else {
        topBarMenu.style.display = "block";
    }
}
