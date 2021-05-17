ready(() =>
  ((valoo) => {
    "use strict";

    const errorState = valoo("");
    const registerData = valoo({
      login: "",
      password: "",
      confirmPassword: "",
    });

    errorState((msg) => {
      const elements = document.querySelectorAll(".err-msg");
      Array.prototype.forEach.call(elements, (el, i) => {
        el.innerText = msg;
      });
    });

    registerData(({ password, confirmPassword }) => {
      // Clear error message.
      errorState("");

      // Check if passwords are the same.
      const repeatPassword = document.getElementById("r-password");
      if (password != confirmPassword) {
        repeatPassword.setCustomValidity("Passwords don't match.");
      } else {
        repeatPassword.setCustomValidity("");
      }
    });

    const ERROR_MSGS = {
      400: "invalid input data",
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
            case 400:
              response.json()
                .then((data) => {
                  err(data.message);
                })
                .catch(() => {
                  err(ERROR_MSGS.default);
                });
              break;
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

    document.getElementById("nickname").addEventListener("input", (e) => {
      registerData({
        ...registerData(),
        login: e.currentTarget.value,
      });
    });

    document.getElementById("password").addEventListener("input", (e) => {
      registerData({
        ...registerData(),
        password: e.currentTarget.value,
      });
    });

    document.getElementById("r-password").addEventListener("input", (e) => {
      registerData({
        ...registerData(),
        confirmPassword: e.currentTarget.value,
      });
    });

    document.getElementById("register-form").addEventListener("submit", (e) => {
      e.preventDefault();
      submitData(registerData());
    });
  })(valoo)
);
