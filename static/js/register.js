ready(() =>
  ((u, kel) => {
    "use strict";

    const store = kel({
      login: "",
      password: "",
      confirmPassword: "",
      error: "",
    });

    const EVENTS = {
      UPDATE: "UPDATE",
      SUBMIT: "SUBMIT",
      ERROR: "ERROR",
    };

    const ERROR_MSGS = {
      404: "invalid input data",
      409: "given username is already registered",
      default: "invalid server response, please try again later",
    };

    store.on(EVENTS.UPDATE, ({ password, confirmPassword }) => {
      // Clear error message.
      u(".err-msg").text("");

      // Check if passwords are the same.
      let repeatPassword = u("#r-password").first();
      if (password != confirmPassword) {
        repeatPassword.setCustomValidity("Passwords don't match.");
      } else {
        repeatPassword.setCustomValidity("");
      }
    });

    store.on(EVENTS.ERROR, ({ login, password, error }) => {
      u(".err-msg").text("server error: " + error);
    });

    const err = (msg) => {
      store.emit(EVENTS.ERROR, (store) => ({
        ...store,
        error: msg,
      }));
    };

    store.on(EVENTS.SUBMIT, (store) => {
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
    });

    u("#nickname").on("input", (e) => {
      store.emit(EVENTS.UPDATE, (store) => {
        return {
          ...store,
          login: e.currentTarget.value,
        };
      });
    });

    u("#password").on("input", (e) => {
      store.emit(EVENTS.UPDATE, (store) => {
        return {
          ...store,
          password: e.currentTarget.value,
        };
      });
    });

    u("#r-password").on("input", (e) => {
      store.emit(EVENTS.UPDATE, (store) => {
        return {
          ...store,
          confirmPassword: e.currentTarget.value,
        };
      });
    });

    u("#register-form").handle("submit", (e) => {
      store.emit(EVENTS.SUBMIT, id);
    });
  })(u, Kel)
);
