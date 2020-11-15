ready(() =>
  ((u, handlebars, valoo) => {
    "use strict";

    const devicesTempl = handlebars.compile(`
      {{#each devices}}
        <div class="device">
          <div class="device-name">{{this.tag}}</div>
          <a class="device-rm" data-id="{{this.id}}">rm</a>
        </div>
      {{/each}}
    `);

    const emptyDevice = { tag: "", mac: "", id: 0 };
    const ids = valoo(0);

    const devices = valoo([]);
    const currentDevice = valoo(emptyDevice);

    const renderDevices = (data) => {
      u(".devices").html(devicesTempl({ devices: data }));

      // TODO(dudekb) Add function for removing devices with given id
      u(".device-rm").on("click", (e) => {
        devices(
          devices().filter((device) => device.id != e.currentTarget.dataset.id),
        );
      });
    };

    // TODO(thinkofher) Add function for fetching devices

    devices(renderDevices);

    u("#tag-form").on("input", (e) => {
      currentDevice({
        ...currentDevice(),
        tag: e.currentTarget.value,
      });
    });

    u("#mac-form").on("input", (e) => {
      currentDevice({
        ...currentDevice(),
        mac: e.currentTarget.value,
      });
    });

    u("#device-form").handle("submit", (e) => {
      // TODO: Add ajax for adding new device.

      // Add current device to devices state
      devices(
        devices().concat({
          ...currentDevice(),
          id: ids(),
        }),
      );

      u("#mac-form, #tag-form").each((node, i) => {
        // u(node).first().value = "";
        node.value = "";
      });

      // Empty current device
      currentDevice(emptyDevice);

      // Add one to current id
      ids(ids() + 1);
    });
  })(u, Handlebars, valoo)
);
