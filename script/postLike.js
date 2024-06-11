// like button
if (document.getElementById('like_btn') != null) {
  document.getElementById('like_btn').addEventListener('change', function() {
      const userId = this.getAttribute('data-user-id');
      const postId = this.getAttribute('data-post-id');
      const url = this.checked ? '/add-liked-post' : '/remove-liked-post';
      
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