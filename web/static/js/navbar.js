ready(() =>
  ((u, el) => {
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
