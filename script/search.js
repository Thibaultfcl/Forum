// search
if (document.getElementById('search') != null) {
  document.getElementById('search').addEventListener('input', function() {
    const suggestions = document.getElementById('suggestions');
    suggestions.innerHTML = '';

    const search = this.value;
    if (search == '') {
      return;
    }

    fetch('/search', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        search: search
      })
    }).then(response => {
      if (!response.ok) {
        console.error('Error:', response.statusText);
      } else {
        response.json().then(data => {
          if (data == null) {
            return;
          }
          data.forEach(suggestion => {
            const option = document.createElement('option');
            option.value = suggestion.name;
            option.dataset.type = suggestion.type;
            if (suggestion.type == 'user') {
              option.dataset.id = suggestion.id;
            }
            suggestions.appendChild(option);
          });
        });
      }
    }).catch(error => {
      console.error('Error:', error);
    });
  });
}

// search redirect
form = document.getElementById('formSearch');
if (form != null) {
  form.addEventListener('submit', function(e) {
    e.preventDefault();
    let search = document.getElementById('search').value;
    const datalist = document.getElementById('suggestions');
    let url = '/searchError/' + search;

    for (let i = 0; i < datalist.options.length; i++) {
      let option = datalist.options[i];
      if (option.value === search) {
        if (option.dataset.type == 'user') {
          search = option.dataset.id;
        }
        url = '/' + option.dataset.type + '/' + search;
        break;
      }
    }
    window.location.href = url;
  });
}