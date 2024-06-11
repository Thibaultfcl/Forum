// get user page
var pps = document.querySelectorAll('.ppUser');
if (pps != null) {
  pps.forEach(pp => {
    pp.addEventListener('click', function(e) {
      e.preventDefault();
      const userId = this.getAttribute('data-user-id');
      window.location.href = '/user/' + userId;
    });
  });
}