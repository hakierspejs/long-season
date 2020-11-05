loginErrorMessages = {
  401: "password does not match",
  404: "there is no user with given nickname",
  "default": "invalid server response, please try again later"
};

function setErrorMessage(msg) {
  u("#err-msg").text(msg);
}

u("#login-form").handle('submit', (e) => {
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
    .then(response => {
      switch(response.status) {
        case 401:
          setErrorMessage(loginErrorMessages[response.status]);
          break;
        case 404:
          setErrorMessage(loginErrorMessages[response.status]);
          break;
        case 200:
          window.location.href = "/";
          break
        default:
          setErrorMessage(loginErrorMessages["default"]);
          break;
      }
    })
    .catch(() => {
      setErrorMessage(loginErrorMessages["default"]);
    });

});
