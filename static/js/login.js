u("#form").handle('submit', (e) => {
  let nickname = u("#nickname").first().value;
  let password = u("#password").first().value;

  let data = {
    'nickname': nickname,
    'password': password
  };

  fetch('/api/v1/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    redirect: 'follow',
    credentials: 'include',
    body: JSON.stringify(data)
  })
    .then(response => response.json())
    .then(data => console.log(data));
});

u("#test").on('click', () => {
  fetch('/secret', {
    method: 'GET',
    credentials: 'same-origin'
  })
    .then(response => {
      console.log(response);
      return response.json();
    })
    .then(data => console.log(data));
});
