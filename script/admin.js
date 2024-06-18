var adminBtns = document.querySelectorAll('.admin');
if (adminBtns != null) {
    adminBtns.forEach(adminBtn => {
        adminBtn.addEventListener('click', function(e) {
            e.preventDefault();
            const userId = this.getAttribute('data-user-id');
            fetch('/switchAdminStatus', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    userID: parseInt(userId),
                })
            }).then(response => {
                if (!response.ok) {
                    console.error('Error:', response.statusText);
                } else {
                    location.reload();
                }
            }).catch(error => {
                console.error('Error:', error);
            });
        });
    });
}

var modoBtns = document.querySelectorAll('.moderator');
if (modoBtns != null) {
    modoBtns.forEach(modoBtn => {
        modoBtn.addEventListener('click', function(e) {
            e.preventDefault();
            const userId = this.getAttribute('data-user-id');
            fetch('/switchModoStatus', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    userID: parseInt(userId),
                })
            }).then(response => {
                if (!response.ok) {
                    console.error('Error:', response.statusText);
                } else {
                    location.reload();
                }
            }).catch(error => {
                console.error('Error:', error);
            });
        });
    });
}

var banBtns = document.querySelectorAll('.ban');
if (banBtns != null) {
    banBtns.forEach(banBtn => {
        banBtn.addEventListener('click', function(e) {
            e.preventDefault();
            const userId = this.getAttribute('data-user-id');
            fetch('/switchBanStatus', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    userID: parseInt(userId),
                })
            }).then(response => {
                if (!response.ok) {
                    console.error('Error:', response.statusText);
                } else {
                    location.reload();
                }
            }).catch(error => {
                console.error('Error:', error);
            });
        });
    });
}

var resetBtns = document.querySelectorAll('.reset-button');
if (resetBtns != null) {
    resetBtns.forEach(resetBtn => {
        resetBtn.addEventListener('click', function(e) {
            e.preventDefault();
            const userId = this.getAttribute('data-user-id');
            fetch('/resetPP', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    userID: parseInt(userId),
                })
            }).then(response => {
                if (!response.ok) {
                    console.error('Error:', response.statusText);
                } else {
                    location.reload();
                }
            }).catch(error => {
                console.error('Error:', error);
            });
        });
    });
}

function displayTable() {
    var select = document.getElementById("tableSelect");
    var adminTable = document.getElementById("adminTable");
    var reportTable = document.getElementById("reportTable");

    // Hide all tables
    adminTable.style.display = "none";
    reportTable.style.display = "none";

    // Show the selected table
    if (select.value === "admin") {
        adminTable.style.display = "block";
    } else if (select.value === "report") {
        reportTable.style.display = "block";
    }
}

window.onload = displayTable;

var binReportButtons = document.querySelectorAll('.bin-button-report');
if (binReportButtons != null) {
    binReportButtons.forEach(binButton => {
        binButton.addEventListener('click', function(e) {
            e.preventDefault();
            const postId = this.getAttribute('data-post-id');
            fetch('/deleteReport', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    postID: parseInt(postId)
                })
            }).then(response => {
                if (!response.ok) {
                    console.error('Error:', response.statusText);
                } else {
                    location.reload();
                }
            }).catch(error => {
                console.error('Error:', error);
            });
        });
    });
}