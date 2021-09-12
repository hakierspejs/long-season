import { el, withErr } from "/static/js/utils.js";
import * as api from "/static/js/api.js";

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

const render = (parentNode, target) => {
  while (parentNode.firstChild) {
    parentNode.removeChild(parentNode.firstChild);
  }

  if (Array.isArray(target)) {
    target.forEach((node, _) => {
      parentNode.append(node);
    });
  } else {
    parentNode.append(target);
  }
};

render(document.getElementById("update-password"), ChangePassword());
