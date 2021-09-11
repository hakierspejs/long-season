import { el } from "/static/js/utils.js";

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

const nav = document.querySelector("nav");

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
    empty(nav);
    Array.prototype.forEach.call(navbar(data), (el, i) => {
      nav.append(el);
    });
  })
  .catch((error) => {
    empty(nav);
    Array.prototype.forEach.call(navbar({}), (el, i) => {
      nav.append(el);
    });
  });
