// category like
if (document.getElementById('likeCheckbox') != null) {
  document.getElementById('likeCheckbox').addEventListener('change', function() {
      const userId = this.getAttribute('data-user-id');
      const categoryId = this.getAttribute('data-category-id');
      const url = this.checked ? '/add-liked-category' : '/remove-liked-category';
      
      fetch(url, {
          method: 'POST',
          headers: {
              'Content-Type': 'application/json'
          },
          body: JSON.stringify({
              userID: parseInt(userId),
              categoryID: parseInt(categoryId)
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