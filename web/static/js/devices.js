ready(() =>
  ((el, valoo) => {
    "use strict";

    // Returns single device component.
    const deviceComp = ({ tag, id }) =>
      el(
        "li",
        {},
        el("span", {}, el("b", {}, tag)),
        el(
          "span",
          {},
          el("a", {
            onClick: () => deleteDevice(id),
            "class": "rm",
          }, "remove"),
        ),
      );

    const privMode = valoo(false);

    const privModeCheckbox = ({ store, onClick }) => {
      const checkbox = el("input", { "type": "checkbox" });

      const text = el("div", { "id": "priv-mode-text" }, "");

      // set checked value
      checkbox.checked = store();
      text.textContent = store() ? " Enabled" : " Disabled";

      // assign new checkbox checked value to given store
      store((checked) => {
        checkbox.checked = checked;
        text.textContent = checked ? " Enabled" : " Disabled";
      });

      checkbox.onclick = onClick;

      return el(
        "p",
        {},
        el(
          "label",
          { "id": "priv-mode-label" },
          checkbox,
          text,
        ),
      );
    };

    // Returns array with devices components constructed from
    // given aray with devices objects.
    const devicesComp = (devices) => {
      return devices.map(deviceComp);
    };

    // Default device data.
    const emptyDevice = { tag: "", mac: "", id: 0 };

    // Global data storages.
    const errorMessage = valoo("");
    const devices = valoo([]);
    const currentDevice = valoo(emptyDevice);

    // Toggle error message when is not empty.
    errorMessage((msg) => {
      const elements = document.querySelectorAll(".err-msg");
      if (msg !== "") {
        Array.prototype.forEach.call(elements, (el, i) => {
          el.classList.remove("hidden");
          el.innerText = msg;
        });
      } else {
        Array.prototype.forEach.call(elements, (el, i) => {
          el.classList.add("hidden");
        });
      }
    });

    const handleErrors = (error) => {
      switch (error.status) {
        case 400:
          errorMessage(serverError("invalid device data"));
          break;
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

    const empty = (target) => {
      while (target.firstChild) {
        target.removeChild(target.firstChild);
      }
    };

    const renderDevices = (data) => {
      const node = document.querySelectorAll(".devices")[0];

      // Clear current rendered devices.
      empty(node);

      // Render new devices.
      devicesComp(data).forEach((device, _) => {
        node.append(device);
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

    const userData = () =>
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
          return fetch("/api/v1/users/" + data.id, {
            method: "GET",
            headers: {
              "Content-Type": "application/json",
            },
            credentials: "include",
          });
        })
        .then(checkResponse)
        .then(responseJSON);

    const renderPrivMode = (store) => {
      userData()
        .then((data) => {
          store(data.priv);
          store(togglePrivateMode);

          const checkbox = privModeCheckbox({
            store: store,
            onClick: () => {
              store(!store());
            },
          });

          document.getElementById("private-mode").append(checkbox);
        })
        .catch(handleErrors);
    };

    const togglePrivateMode = (mode) => {
      userData()
        .then((data) => {
          return fetch("/api/v1/users/" + data.id, {
            method: "PATCH",
            headers: {
              "Content-Type": "application/json",
            },
            credentials: "include",
            body: JSON.stringify({ priv: mode }),
          });
        })
        .then(checkResponse)
        .then(responseJSON)
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

    document.getElementById("tag-form").addEventListener("input", (e) => {
      currentDevice({
        ...currentDevice(),
        tag: e.currentTarget.value,
      });
    });

    document.getElementById("mac-form").addEventListener("input", (e) => {
      currentDevice({
        ...currentDevice(),
        mac: e.currentTarget.value,
      });
    });

    document.getElementById("device-form").addEventListener("submit", (e) => {
      e.preventDefault();

      // Post current device to API
      postDevice(currentDevice());

      // Clear form
      document.querySelectorAll("#mac-form, #tag-form").forEach((el, i) => {
        el.value = "";
      });

      // Empty current device
      currentDevice(emptyDevice);
    });

    // Initial fetch devices.
    fetchDevices();

    // Render private mode checkbox
    renderPrivMode(privMode);
  })(el, valoo)
);
