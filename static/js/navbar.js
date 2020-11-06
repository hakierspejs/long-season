ready(() =>
  ((u, handlebars) => {
    "use strict";

    const navbar = handlebars.compile(`
    <div class="at">hackerspace@lodz:~$</div>
    <div class="elem"><a href="/">home</a></div>
    {{#if nickname}}
      <div class="elem"><a href="#">{{ nickname }}@hsldz</a></div>
      <div class="elem"><a href="/logout">logout</a></div>
    {{else}}
      <div class="elem"><a href="/register">register</a></div>
      <div class="elem"><a href="/login">login</a></div>
    {{/if}}
  `);

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
        u("nav").html(navbar(data));
      })
      .catch((error) => {
        u("nav").html(navbar({}));
      });
  })(u, Handlebars)
);
