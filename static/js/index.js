const usersTemplate = Handlebars.compile(`
<ul>
  {{#each users}}
    <li>{{this.nickname}} is <span>
      {{#if this.online}}
        online
      {{else}}
        offline
      {{/if}}
    </span>
    </li>
  {{/each}}
</ul>
`);

const downloadUsers = () => {
  u("#users").text("Loading...");
  fetch('/users')
    .then(response => response.json())
    .then(data => {
      u("#users").html(usersTemplate({ users: data }));
    })
    .catch(() => {
      u("#users").text("Failed to load users data.");
    });
};

downloadUsers();
window.setInterval(downloadUsers, 1000 * 60 * 2);
