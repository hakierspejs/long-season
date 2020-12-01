ready(() =>
  ((u, el) => {
    "use strict";

    const elem = (...children) =>
      el(
        "div",
        { "class": "elem" },
        ...children,
      );

    const a = (link, ...children) =>
      el(
        "a",
        { href: link },
        ...children,
      );

    const navbar = ({ nickname }) => {
      return [
        el("div", { "class": "at" }, "hackerspace@lodz:~$"),
        elem(a("/", "home")),
        ...(nickname
          ? [
            elem(a("/devices", `${nickname}@hsldz`)),
            elem(a("/logout", "logout")),
          ]
          : [
            elem(a("/register", "register")),
            elem(a("/login", "login")),
          ]),
      ];
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
        let nav = u("nav");
        nav.empty();
        nav.append(navbar(data));
      })
      .catch((error) => {
        let nav = u("nav");
        nav.empty();
        nav.append(navbar({}));
      });
  })(u, el)
);
