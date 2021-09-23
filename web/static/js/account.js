import { el, main, render, withErr } from "/static/js/utils.js";
import * as api from "/static/js/api.js";
import * as otp from "/static/js/otp.js";

const PasswordInput = (label, props) => {
  props.type = "password";

  return el(
    "p",
    null,
    el("label", { "for": props.for }, label),
    el("br", null, null),
    el("input", props),
  );
};

const SubmitButton = (props) => {
  props.type = "submit";
  return el(
    "p",
    null,
    el("button", props, "Change Password"),
  );
};

const ChangePassword = () => {
  let state = {
    oldPass: "",
    newPass: "",
    newPassRepeat: "",
  };

  const errMsg = el("p", null, "");

  // define input fields with labels for
  // changing password
  const oldPasswordInput = PasswordInput("Old password", {
    "name": "old-password",
    onInput: (e) => {
      errMsg.textContent = "";
      state.oldPass = e.currentTarget.value;
    },
  });
  const passwordInput = PasswordInput("New password", {
    "name": "new-password",
    onInput: (e) => {
      errMsg.textContent = "";
      state.newPass = e.currentTarget.value;
    },
  });
  const repeatPasswordInput = PasswordInput("Repeat new password", {
    "name": "new-password-repeat",
    onInput: (e) => {
      errMsg.textContent = "";
      state.newPassRepeat = e.currentTarget.value;
    },
  });

  const clearInput = () => {
    [oldPasswordInput, passwordInput, repeatPasswordInput].forEach(
      (node, _) => {
        node.querySelector("input").value = "";
      },
    );
  };

  const submitChange = async (event) => {
    event.preventDefault();
    errMsg.textContent = "";

    if (state.newPass === "" || state.newPassRepeat === "") {
      errMsg.textContent = "Passwords can't be empty.";
      return;
    }

    if (state.newPass !== state.newPassRepeat) {
      errMsg.textContent = "Password does not match.";
      return;
    }
    let [user, whoErr] = await api.who();
    if (whoErr) {
      errMsg.textContent = err.message;
      return;
    }

    let [res, passErr] = await api.updatePassword(user.id, {
      oldPass: state.oldPass,
      newPass: state.newPass,
    });
    if (passErr) {
      errMsg.textContent = passErr.message;
      return;
    }
    if (!res.ok) {
      let [json, jsonErr] = await withErr(res.json());
      if (jsonErr) {
        errMsg.textContent = jsonErr.message;
      }
      errMsg.textContent = json.error.message;
      return;
    }

    clearInput();
    errMsg.textContent = "Succesfully changed password!";
  };

  return [
    el("strong", null, errMsg),
    oldPasswordInput,
    passwordInput,
    repeatPasswordInput,
    SubmitButton({
      name: "submit-update-password",
      onClick: submitChange,
    }),
  ];
};

// Returns single device component.
const TwoFactorMethod = ({ name, type, onRemove }) =>
  el(
    "li",
    {},
    el("span", {}, el("b", {}, `${name} (${type})`)),
    el(
      "span",
      {},
      el("a", {
        onClick: onRemove,
        "class": "rm",
      }, "remove"),
    ),
  );

const TwoFactorMethods = (userID, methods) => {
  // There are no methods, so there is no need
  // to render anything.
  if (methods && methods.length === 0) {
    return "";
  }

  const errContainer = el("strong", null, "");

  return [
    methods ? el("h3", null, "Current methods") : "",
    ...(methods
      ? methods.map((method) =>
        TwoFactorMethod({
          name: method.name,
          type: method.type,
          onRemove: async (e) => {
            let err = await api.removeTwoFactorMethod(userID, method.id);
            if (err) {
              errContainer = "Failed to remove second authentication method.";
              return;
            }
            renderTwoFactorMethods();
          },
        })
      )
      : []),
  ];
};

const twoFactorMethods = document.getElementById("two-factor-methods");

async function renderTwoFactorMethods() {
  let [user, errWho] = await api.who();
  if (errWho) {
    return;
  }

  let [methods, errMethods] = await api.twoFactorMethods(user.id);
  if (errMethods) {
    return;
  }

  render(twoFactorMethods, TwoFactorMethods(user.id, methods.active));
}

main(() => {
  // Render form for changing password.
  render(document.getElementById("update-password"), ChangePassword());

  // Initial rendering of two factor methods.
  renderTwoFactorMethods();

  // Mount OTP codes.
  otp.mount({
    // Render two factor methods every time user
    // submits successfully new OTP code.
    onAdd: renderTwoFactorMethods,
  });
});
