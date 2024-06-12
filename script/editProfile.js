const editProfile = document.getElementById("editProfile");
const profileForm = document.getElementById("profile-form");
let profileActive = false;

// display the profile form
if (editProfile != null) {
  editProfile.addEventListener("click", function () {
    if (!profileActive) {
      profileForm.style.display = "block";
      overlay.style.display = "block";
      profileActive = true;
    }
  });
}

// remove the profile form by clicking outside of it
overlay.addEventListener("click", function () {
  if (profileActive) {
    profileForm.style.display = "none";
    overlay.style.display = "none";
    profileActive = false;
  }
});