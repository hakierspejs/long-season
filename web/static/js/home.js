ready(() =>
  ((u, el) => {
    "use strict";

    const onlineStatus = (usersCount) => {
      let text = "";
      switch (usersCount) {
        case 0:
          text = HACKER_STATE.CLOSED;
          break;
        case 1:
          text = HACKER_STATE.FOREVER_ALONE;
          break;
        default:
          text = HACKER_STATE.PARTY(usersCount);
          break;
      }
      return el("p", null, text);
    };

    const onlineTitle = (length) =>
      el(
        "h3",
        length > 0 ? null : { style: "display:none;" },
        "Who is online?",
      );

    const usersComp = (users) =>
      el(
        "ul",
        null,
        ...(
          users.map((user) => el("li", null, user.nickname))
        ),
      );

    const homeComp = (users) =>
      el(
        "div",
        { id: "app" },
        onlineStatus(users.length),
        onlineTitle(users.length),
        usersComp(users),
      );

    const HACKER_STATE = {
      CLOSED: "Hackerspace is closed.",
      FOREVER_ALONE: "There is one person in the hackerspace.",
      PARTY: (num) => "There are " + num + " people in the hackerspace.",
    };

    const users = valoo([]);

    users((data) => {
      u("#app").replace(homeComp(data));
    });

    const clearApp = () => el("div", null, "");

    const downloadUsers = () => {
      fetch("/api/v1/users?online=true")
        .then((response) => response.json())
        .then((data) => users(data))
        .catch(() => {
          u("#info").text("Failed to load users data.");
          u("#app").replace(clearApp());
        });
    };

    downloadUsers();
    window.setInterval(downloadUsers, 1000 * 60 * 2);
  })(u, el)
);
