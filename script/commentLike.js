// comment like
var commentCheckboxes = document.querySelectorAll('.checkboxComment');
if (commentCheckboxes != null) {
  for (var i = 0; i < commentCheckboxes.length; i++) {
    commentCheckboxes[i].addEventListener('change', function() {
        const userId = this.getAttribute('data-user-id');
        const commentId = this.getAttribute('data-comment-id');
        const url = this.checked ? '/add-liked-comment' : '/remove-liked-comment';
        fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                userID: parseInt(userId),
                commentID: parseInt(commentId)
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