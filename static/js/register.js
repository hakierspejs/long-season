ready(() =>
  ((u, valoo) => {
    "use strict";

    const errorState = valoo("");
    const registerData = valoo({
      login: "",
      password: "",
      confirmPassword: "",
    });

    errorState((msg) => {
      u(".err-msg").text(msg);
    });

    registerData(({ password, confirmPassword }) => {
      // Clear error message.
      errorState("");

      // Check if passwords are the same.
      let repeatPassword = u("#r-password").first();
      if (password != confirmPassword) {
        repeatPassword.setCustomValidity("Passwords don't match.");
      } else {
        repeatPassword.setCustomValidity("");
      }
    });

    const ERROR_MSGS = {
      404: "invalid input data",
      409: "given username is already registered",
      default: "invalid server response, please try again later",
    };

    const err = (msg) => {
      errorState("server error: " + msg);
    };

    const submitData = (store) => {
      let data = {
        nickname: store.login,
        password: store.password,
      };

      fetch("/api/v1/users", {
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
            case 409:
              err(ERROR_MSGS[response.status]);
              break;
            case 200:
              window.location.href = "/";
              break;
            default:
              err(ERROR_MSGS.default);
              break;
          }
        })
        .catch(() => {
          err(ERROR_MSGS.default);
        });
    };

    u("#nickname").on("input", (e) => {
      registerData({
        ...registerData(),
        login: e.currentTarget.value,
      });
    });

    u("#password").on("input", (e) => {
      registerData({
        ...registerData(),
        password: e.currentTarget.value,
      });
    });

    u("#r-password").on("input", (e) => {
      registerData({
        ...registerData(),
        confirmPassword: e.currentTarget.value,
      });
    });

    u("#register-form").handle("submit", (e) => {
      submitData(registerData());
    });
  })(u, valoo)
);
