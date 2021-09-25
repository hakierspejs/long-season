import { el, main, render, empty, withErr } from "/static/js/utils.js";
import * as api from "/static/js/api.js";

function genCode(length) {
  let result = "";
  let characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
  let len = characters.length;
  for (var i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(
      Math.random() *
        len,
    ));
  }
  return result;
}

const FormInput = ({ label, forTag, props }) =>
  el(
    "p",
    null,
    el("label", { "for": forTag }, label),
    el("br", null, null),
    el("input", props),
  );

const FormButton = (desc, props) => el("button", props, desc);

const Quote = (str) => el("pre", {}, el("code", {}, str));

const Recovery = ({ targetForm, targetButton, onAdd }) => {
  let name = "";

  const errContainer = el("strong", null, "");
  const clearErr = () => {
    errContainer.innerText = "";
  };

  const nameForm = FormInput({
    label: "Enter name for generated recovery codes.",
    forTag: "recovery-name",
    props: {
      type: "text",
      name: "recovery-name",
      onInput: (e) => {
        name = e.currentTarget.value;
        clearErr();
      },
    },
  });

  let generatedCodes = Array(10).fill(null).map(() => genCode(20));

  return [
    el("h4", null, "Recovery Codes"),
    el("p", null, errContainer),
    nameForm,
    el(
      "p",
      null,
      "Write down or print below codes and use them as additional recovery method, when you'll lost your device with OTP codes.",
    ),
    el("pre", null, el("code", null, generatedCodes.join("\n"))),
    el(
      "p",
      null,
      FormButton("Submit", {
        onClick: async (event) => {
          event.preventDefault();

          let err = await api.newRecovery({
            "name": name,
            "codes": generatedCodes,
          });
          if (err) {
            errContainer.innerText = "Failed to add recovery codes to account.";
            return;
          }

          render(
            targetForm,
            el(
              "p",
              null,
              el(
                "strong",
                null,
                "Successfully added Recovery Codes two factor method to account.",
              ),
            ),
          );

          // Run onAdd hook after successfully submitting
          // new recovery codes.
          onAdd();
        },
      }),
      FormButton("Cancel", {
        onClick: (event) => {
          event.preventDefault();
          empty(targetForm);
        },
      }),
    ),
  ];
};

function mount({ targetForm, targetButton, onAdd }) {
  targetButton.addEventListener("click", async () => {
    render(
      targetForm,
      Recovery({
        onAdd: onAdd,
        targetForm: targetForm,
        targetButton: targetButton,
      }),
    );
  });
}

export { mount };
