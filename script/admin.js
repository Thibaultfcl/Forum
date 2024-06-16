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