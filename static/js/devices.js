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

    const errorMessage = valoo("");
    const devices = valoo([]);
    const currentDevice = valoo(emptyDevice);

    // Toggle error message when is not empty.
    errorMessage((msg) => {
      if (msg) {
        u(".err-msg").removeClass("hidden");
        u(".err-msg").text(msg);
      } else {
        u(".err-msg").addClass("hidden");
      }
    });

    const handleErrors = (error) => {
      switch (error.status) {
        case 401:
          errorMessage(serverError("invalid user data, please login in"));
          break;
        case 409:
          errorMessage(serverError("there is already resource with given tag"));
          break;
        default:
          errorMessage(serverError("internal server error, please try again"));
          break;
      }
    };

    // Clear error message whenever someone enters data for
    // new device.
    currentDevice(() => errorMessage(""));

    const serverError = (msg) => "server error: " + msg;

    const renderDevices = (data) => {
      // Render html with given devices data .devices class
      u(".devices").html(devicesTempl({ devices: data }));

      // Add event handler to every "rm" a href
      u(".device-rm").on("click", (e) => {
        deleteDevice(e.currentTarget.dataset.id);
      });
    };

    const checkResponse = (response) => {
      if (!response.ok) {
        return Promise.reject(response);
      }
      return response;
    };

    const responseJSON = (response) => response.json();

    const fetchDevices = () => {
      fetch("/who", {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
      })
        .then(checkResponse)
        .then(responseJSON)
        .then((data) => {
          return fetch("/api/v1/users/" + data.id + "/devices", {
            method: "GET",
            headers: {
              "Content-Type": "application/json",
            },
            credentials: "include",
          });
        })
        .then(checkResponse)
        .then(responseJSON)
        .then((data) => devices(data))
        .catch(handleErrors);
    };

    const addDevice = ({ tag, id }) => {
      // Add given device to devices state
      devices(
        devices().concat({
          tag: tag,
          id: id,
        }),
      );
    };

    const postDevice = ({ tag, mac }) => {
      fetch("/who", {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
      })
        .then(checkResponse)
        .then(responseJSON)
        .then((data) => {
          return fetch("/api/v1/users/" + data.id + "/devices", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            credentials: "include",
            body: JSON.stringify({ tag: tag, mac: mac }),
          });
        })
        .then(checkResponse)
        .then(responseJSON)
        .then(addDevice)
        .catch(handleErrors);
    };

    // removeDevice removes device with given device id
    // from device state manager.
    const removeDevice = (deviceID) => {
      devices(
        devices().filter((item) => item.id != deviceID),
      );
    };

    // deleteDevice sends delete request to API
    // to remove device with given ID from
    // user collection.
    //
    // After successful request removes device
    // with given id from devices storage.
    const deleteDevice = (deviceID) => {
      fetch("/who", {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
      })
        .then(checkResponse)
        .then(responseJSON)
        .then((data) => {
          return fetch("/api/v1/users/" + data.id + "/devices/" + deviceID, {
            method: "DELETE",
            headers: {
              "Content-Type": "application/json",
            },
            credentials: "include",
          });
        })
        .then(checkResponse)
        .then(() => {
          removeDevice(deviceID);
        })
        .catch(handleErrors);
    };

    // Listen for changes at devices and render
    // new devices every time new device is added
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
      // Post current device to API
      postDevice(currentDevice());

      // Clear form
      u("#mac-form, #tag-form").each((node, i) => {
        node.value = "";
      });

      // Empty current device
      currentDevice(emptyDevice);
    });

    // Initial fetch devices.
    fetchDevices();
  })(u, Handlebars, valoo)
);
