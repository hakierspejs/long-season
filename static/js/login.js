ready(() =>
  ((u, kel) => {
    "use strict";

    const store = kel({ login: "", password: "", error: "" });

    const EVENTS = {
      UPDATE: "UPDATE",
      SUBMIT: "SUBMIT",
      ERROR: "ERROR",
    };

    const ERROR_MSGS = {
      401: "password does not match",
      404: "there is no user with given nickname",
      default: "invalid server response, please try again later",
    };

    store.on(EVENTS.UPDATE, id);

    store.on(EVENTS.ERROR, ({ login, password, error }) => {
      u("#err-msg").text(error);
    });

    const err = (msg) => {
      store.emit(EVENTS.ERROR, (store) => ({
        ...store,
        error: msg,
      }));
    };

    store.on(EVENTS.SUBMIT, ({ login, password }) => {
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

    u("#login-form").handle("submit", (e) => {
      store.emit(EVENTS.SUBMIT, id);
    });
  })(u, Kel)
);
