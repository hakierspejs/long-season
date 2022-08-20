import { el, valoo } from "/static/js/utils.js";

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

const unknownStatus = (unknownDevicesCount) => {
  switch (unknownDevicesCount) {
    case 0:
      return el("p", { style: "display:none;" }, "");
    case 1:
      return el("p", null, UNKNOWN_STATE.FOREVER_ALONE);
    default:
      return el("p", null, UNKNOWN_STATE.PARTY(unknownDevicesCount));
  }
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

const homeComp = (data) =>
  el(
    "div",
    { id: "app" },
    onlineStatus(data.users.length),
    unknownStatus(data.unknownDevices),
    onlineTitle(data.users.length),
    usersComp(data.users),
  );

const HACKER_STATE = {
  CLOSED: "Hackerspace is closed.",
  FOREVER_ALONE: "There is one person in the hackerspace.",
  PARTY: (num) => "There are " + num + " people in the hackerspace.",
};

const UNKNOWN_STATE = {
  FOREVER_ALONE: "There is one unknown device in the hackerspace.",
  PARTY: (num) => "There are " + num + " unknown devices in the hackerspace.",
};

const homeStorage = valoo({
  users: [],
  onlineUsers: 0,
  unknownDevices: 0,
});

const replace = (toReplace, replecament) => {
  if (toReplace !== null) {
    const parentNode = toReplace.parentNode;
    parentNode.replaceChild(replecament, toReplace);
  }
};

homeStorage((data) => {
  replace(document.getElementById("app"), homeComp(data));
});

const clearApp = () => el("div", { "id": "app" }, "");

const fetchData = () => {
  const info = document.getElementById("info");

  // clear info text
  info.innerText = "";

  fetch("/api/v1/users?online=true")
    .then((response) => response.json())
    .then((users) =>
      homeStorage({
        ...homeStorage(),
        users: users,
      })
    )
    .catch(() => {
      info.innerText = "Failed to load users data.";
      replace(document.getElementById("app"), clearApp());
    });

  fetch("/api/v1/status")
    .then((response) => response.json())
    .then((data) =>
      homeStorage({
        ...homeStorage(),
        onlineUsers: data.online,
        unknownDevices: data.unknown,
      })
    )
    .catch(() => {
      info.innerText = "Failed to load users data.";
      replace(document.getElementById("app"), clearApp());
    });
};

fetchData();
window.setInterval(fetchData, 1000 * 60 * 2);
