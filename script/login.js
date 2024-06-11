const loginForm = document.getElementById("login-form");
const loginButton = document.getElementById("navButton");
const loginButtonMobile = document.getElementById("navButtonMobile");
const login = document.getElementById("login");
const loginMobile = document.getElementById("loginMobile");
const signUpButton = document.getElementById("signUp");
const signInButton = document.getElementById("signIn");
const container = document.getElementById("container");
const overlay = document.getElementById("overlay");
let loginActive = false;

// display the login form for mobile
loginButtonMobile.addEventListener("click", function () {
  getMobileMenu.style.transform = "translateX(-100%)";
  setTimeout(function () {
    getMobileMenu.style.display = "none";
  }, 300);
  if (!loginActive) {
    loginForm.style.display = "block";
    document.getElementById("overlay").style.display = "block";
    loginActive = true;
  }
});

// display the login form
if (loginButton != null) {
  loginButton.addEventListener("click", function () {
    if (!loginActive) {
      loginForm.style.display = "block";
      document.getElementById("overlay").style.display = "block";
      loginActive = true;
    }
  });
}

// switch between the 2 panels
signUpButton.addEventListener("click", () => {
  container.classList.add("right-panel-active");
});
signInButton.addEventListener("click", () => {
  container.classList.remove("right-panel-active");
});

// remove the login form by clicking outside of it
overlay.addEventListener("click", function () {
  if (loginActive) {
    loginForm.style.display = "none";
    overlay.style.display = "none";
    loginActive = false;
  }
});