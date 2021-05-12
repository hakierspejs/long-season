ready(() =>
  ((el) => {
    "use strict";

    const a = (link, ...children) =>
      el(
        "a",
        { href: link },
        ...children,
      );

    const navbar = ({ nickname }) => {
      return [
        a("/", "Home"),
        ...(nickname
          ? [
            a("/devices", `${nickname}@lodz`),
            a("/logout", "Logout"),
          ]
          : [
            a("/register", "Register"),
            a("/login", "Login"),
          ]),
      ];
    };

    const empty = (target) => {
      while (target.firstChild) {
        target.removeChild(target.firstChild);
      }
    };

    fetch("/who", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((response) => {
        if (response.ok) {
          return response.json();
        }
        return Promise.reject(response);
      })
      .then((data) => {
        let nav = document.getElementsByTagName("nav")[0];
        empty(nav);
        Array.prototype.forEach.call(navbar(data), (el, i) => {
          nav.append(el);
        });
      })
      .catch((error) => {
        let nav = document.getElementsByTagName("nav")[0];
        empty(nav);
        Array.prototype.forEach.call(navbar({}), (el, i) => {
          nav.append(el);
        });
      });
  })(el)
);
