const newPost = document.getElementById("newPost");
const postForm = document.getElementById("post-form");
let postActive = false;

// display the post form
if (newPost != null) {
  newPost.addEventListener("click", function () {
    if (!postActive) {
      postForm.style.display = "block";
      overlay.style.display = "block";
      postActive = true;
    }
  });
}

// remove the post form by clicking outside of it
overlay.addEventListener("click", function () {
  if (postActive) {
    postForm.style.display = "none";
    overlay.style.display = "none";
    postActive = false;
  }
});