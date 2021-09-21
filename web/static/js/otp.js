import { el, empty, render } from "/static/js/utils.js";
import * as api from "/static/js/api.js";

const twoFactorForm = document.getElementById("two-factor-form");
const addOTPButton = document.getElementById("add-otp");

const FormInput = ({ label, forTag, props }) =>
  el(
    "p",
    null,
    el("label", { "for": forTag }, label),
    el("br", null, null),
    el("input", props),
  );

const ImageOTP = (src) => el("img", { src: src, style: "width:200px;" }, null);

const FormButton = (desc, props) => el("button", props, desc);

const OTP = (options) => {
  let state = {
    name: "",
    code: "",
  };

  const errContainer = el("strong", null, "");
  const clearErr = () => {
    errContainer.innerText = "";
  };

  const name = FormInput({
    label: "Enter name of device with OTP codes.",
    forTag: "otp-names",
    props: {
      type: "text",
      name: "otp-names",
      onInput: (e) => {
        state.name = e.currentTarget.value;
        clearErr();
      },
    },
  });

  const validateCode = FormInput({
    label: "Enter current code to validete process.",
    forTag: "otp-validate",
    props: {
      type: "number",
      onInput: (e) => {
        if (e.currentTarget.value.length > 6) {
          e.currentTarget.value = e.currentTarget.value.slice(0, 6);
        }
        state.code = e.currentTarget.value;
        clearErr();
      },
      name: "otp-validate",
    },
  });

  return [
    el("h4", null, "One Time Codes"),
    el("p", null, errContainer),
    name,
    el(
      "p",
      null,
      "Scan below image with your device or use secret phrase with your authenticator application.",
    ),
    ImageOTP(options.image),
    el("p", null, "Secret phrase is: ", el("strong", null, options.secret)),
    validateCode,
    el(
      "p",
      null,
      FormButton("Submit", {
        onClick: async (event) => {
          event.preventDefault();
          clearErr();

          let err = await api.newOTP({
            "name": state.name,
            "secret": options.secret,
            "code": state.code,
          });
          if (err) {
            errContainer.innerText = "Failed to add OTP to account.";
            return;
          }
          addOTPButton.disabled = false;
          render(
            twoFactorForm,
            el(
              "p",
              null,
              el(
                "strong",
                null,
                "Successfully added OTP authenticator method to account.",
              ),
            ),
          );
        },
      }),
      FormButton("Cancel", {
        onClick: () => {
          event.preventDefault();
          empty(twoFactorForm);
          addOTPButton.disabled = false;
        },
      }),
    ),
  ];
};

addOTPButton.addEventListener("click", async () => {
  let [options, err] = await api.optionsOTP();
  if (err) {
    render(twoFactorForm, el("strong", null, "Failed to add OTP to account."));
  }
  render(twoFactorForm, OTP(options));
  addOTPButton.disabled = true;
});
