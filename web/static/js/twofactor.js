import { el, empty, main, render } from "/static/js/utils.js";
import * as api from "/static/js/api.js";

const twoFactorCodesButton = document.getElementById("two-factor-codes");
const twoFactorMethodSection = document.getElementById("two-factor-method");

const Codes = () => {
  let code = "";

  let errorContainer = el("strong", null, "");

  return [
    el("h3", null, "Codes"),
    el(
      "p",
      null,
      "Enter one of your codes in below text input form in order to confirm your identity.",
    ),
    el(
      "form",
      null,
      el("p", null, errorContainer),
      el("input", {
        type: "text",
        onInput: (e) => {
          code = e.currentTarget.value;
          errorContainer.innerText = "";
        },
      }, null),
      el(
        "p",
        null,
        el("button", {
          onClick: async (e) => {
            e.preventDefault();
            let err = await api.authWithCodes(code);
            if (err) {
              errorContainer.innerText =
                "Failed to authenticate with given code.";
              return;
            }
            window.location.href = "/";
          },
        }, "Submit"),
        el("button", {
          onClick: () => {
            empty(twoFactorMethodSection);
          },
        }, "Cancel"),
      ),
    ),
  ];
};

main(() => {
  twoFactorCodesButton.addEventListener("click", () => {
    render(twoFactorMethodSection, Codes());
  });
});
