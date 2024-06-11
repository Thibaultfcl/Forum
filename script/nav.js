const getHubergerIcon = document.getElementById("hamburger-menu");
const getHubergerCrossIcon = document.getElementById("hamburger-cross");
const getMobileMenu = document.getElementById("mobile-menu");
const mobileNavItems = document.getElementById("mobile-nav-items");

getHubergerIcon.addEventListener("click", function () {
  getMobileMenu.style.display = "flex";
  setTimeout(function () {
    getMobileMenu.style.transform = "translateX(0%)";
  }, 50);
});

getHubergerCrossIcon.addEventListener("click", function () {
  getMobileMenu.style.transform = "translateX(-100%)";
  setTimeout(function () {
    getMobileMenu.style.display = "none";
  }, 300);
});

mobileNavItems.addEventListener("click", function () {
  getMobileMenu.style.transform = "translateX(-100%)";
  setTimeout(function () {
    getMobileMenu.style.display = "none";
  }, 300);
});

window.addEventListener("resize", function () {
  if (window.innerWidth > 770) {
    getMobileMenu.style.display = "none";
  }
});