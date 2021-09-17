import { valoo } from "/static/js/utils.js";

const errorState = valoo("");
const loginData = valoo({ login: "", password: "" });

errorState((msg) => {
  const elements = document.querySelectorAll(".err-msg");
  Array.prototype.forEach.call(elements, (el, i) => {
    el.innerText = msg;
  });
});

loginData(() => {
  errorState("");
});

const ERROR_MSGS = {
  401: "password does not match",
  404: "there is no user with given nickname",
  default: "invalid server response, please try again later",
};

const err = (msg) => {
  errorState("server error: " + msg);
};

const submitData = ({ login, password }) => {
  let data = {
    nickname: login,
    password: password,
  };

  fetch("/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    redirect: "follow",
    credentials: "include",
    body: JSON.stringify(data),
  })
    .then((response) => {
      switch (response.status) {
        case 401:
          err(ERROR_MSGS[response.status]);
          break;
        case 404:
          err(ERROR_MSGS[response.status]);
          break;
        case 200:
          window.location.href = "/";
          break;
        default:
          err(ERROR_MSGS["default"]);
          break;
      }
    })
    .catch(() => {
      err(ERROR_MSGS["default"]);
    });
};

document.getElementById("nickname").addEventListener("input", (e) => {
  loginData({
    ...loginData(),
    login: e.currentTarget.value,
  });
});

document.getElementById("password").addEventListener("input", (e) => {
  loginData({
    ...loginData(),
    password: e.currentTarget.value,
  });
});

document.getElementById("login-form").addEventListener("submit", (e) => {
  e.preventDefault();
  submitData(loginData());
});
