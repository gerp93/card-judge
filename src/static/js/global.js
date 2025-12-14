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

document.addEventListener("htmx:confirm", function (event) {
    const message = event.target.getAttribute("hx-confirm");
    if (message) {
        // override default browser confirm with custom dialog
        event.preventDefault();
        confirmationDialogCreate(event, message);
    }
});

let confirmationEvent = null;

function confirmationDialogCreate(event, message) {
    confirmationEvent = event;
    document.getElementById("confirmation-dialog-message").textContent = message;
    document.getElementById("confirmation-dialog").showModal();
}

function confirmationDialogDelete() {
    document.getElementById("confirmation-dialog").close();
    document.getElementById("confirmation-dialog-message").textContent = "";
    confirmationEvent = null;
}

function confirmationAnswerYes() {
    if (confirmationEvent && confirmationEvent.detail) {
        confirmationEvent.detail.issueRequest(true);
    }
    confirmationDialogDelete();
}

function confirmationAnswerNo() {
    confirmationDialogDelete();
}

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

function goToPreviousTablePage() {
    const pageNumberElement = document.getElementById("pageNumber");
    if (pageNumberElement) {
        const currentPage = parseInt(pageNumberElement.value) || 1;
        goToTablePage(Math.max(currentPage - 1, 1));
    }
}

function goToNextTablePage() {
    const pageNumberElement = document.getElementById("pageNumber");
    if (pageNumberElement) {
        const currentPage = parseInt(pageNumberElement.value) || 1;
        const totalPages = parseInt(pageNumberElement.max) || 1;
        goToTablePage(Math.min(currentPage + 1, totalPages));
    }
}

function goToTablePage(pageNumber) {
    const pageNumberElement = document.getElementById("pageNumber");
    if (pageNumberElement) {
        if (pageNumberElement.value != pageNumber) {
            pageNumberElement.value = pageNumber;
            submitTableFilterForm();
        }
    }
}

function submitTableFilterForm() {
    const formElement = document.getElementById("table-filter-form");
    if (formElement) {
        formElement.submit();
    }
}