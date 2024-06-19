const RecentRadio = document.getElementById("recent");
const PopularRadio = document.getElementById("like");

const RecentPosts = document.getElementById("recentPost");
const PopularPosts = document.getElementById("mostLikedPost");

RecentRadio.addEventListener("change", function() {
    if (RecentRadio.checked) {
        RecentPosts.style.display = "block";
        PopularPosts.style.display = "none";
    }
});

PopularRadio.addEventListener("change", function() {
    if (PopularRadio.checked) {
        RecentPosts.style.display = "none";
        PopularPosts.style.display = "block";
    }
});