ready(() =>
  ((u, valoo) => {
    "use strict";

    const errorState = valoo("");
    const loginData = valoo({ login: "", password: "" });

    errorState((msg) => {
      u(".err-msg").text(msg);
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

      fetch("/api/v1/login", {
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

    u("#nickname").on("input", (e) => {
      loginData({
        ...loginData(),
        login: e.currentTarget.value,
      });
    });

    u("#password").on("input", (e) => {
      loginData({
        ...loginData(),
        password: e.currentTarget.value,
      });
    });

    u("#login-form").handle("submit", (e) => {
      submitData(loginData());
    });
  })(u, valoo)
);
