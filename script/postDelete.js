// delete post
var binButtons = document.querySelectorAll('.bin-button');
if (binButtons != null) {
    binButtons.forEach(binButton => {
        binButton.addEventListener('click', function(e) {
            e.preventDefault();
            const postId = this.getAttribute('data-post-id');
            const categoryId = this.getAttribute('data-category-id');
            fetch('/deletePost', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    postID: parseInt(postId),
                    categoryId: parseInt(categoryId)
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