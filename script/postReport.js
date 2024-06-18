// report post
var postReport = document.querySelectorAll('.report');
if (postReport != null) {
    for (var i = 0; i < postReport.length; i++) {
        postReport[i].addEventListener('change', function() {
            const userId = this.getAttribute('data-user-id');
            const postId = this.getAttribute('data-post-id');
            const url = this.checked ? '/report-post' : '/unreport-post';
            fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    userID: parseInt(userId),
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
    }
}