const usersTemplate = Handlebars.compile(`
<ul>
  {{#each users}}
    <li>{{this.nickname}}</li>
  {{/each}}
</ul>
`);

const hackerState = {
  closed:       "Hackerspace is closed.",
  foreverAlone: "There is one person in the hackerspace.",
  party:        (num) => "There are " + num + " people in the hackerspace."
};

const downloadUsers = () => {
  u("#users").text("Loading...");
  fetch('/api/v1/users?online=true')
    .then(response => response.json())
    .then(data => {
      u("#users").html(usersTemplate({ users: data }));
      switch(data.length) {
        case 0:
          u("#online").text(hackerState.closed);
          break;
        case 1:
          u("#online").text(hackerState.foreverAlone);
          break;
        default:
          u("#online").text(hackerState.party(data.length));
          break;
      }
    })
    .catch(() => {
      u("#users").text("Failed to load users data.");
    });
};

downloadUsers();
window.setInterval(downloadUsers, 1000 * 60 * 2);
