// delete comment
var binButtonComments = document.querySelectorAll('.bin-button-comment');
if (binButtonComments != null) {
    binButtonComments.forEach(binButtonComment => {
        binButtonComment.addEventListener('click', function(e) {
            e.preventDefault();
            const commentId = this.getAttribute('data-comment-id');
            fetch('/deleteComment', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    commentID: parseInt(commentId),
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