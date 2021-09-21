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
  const name = FormInput({
    label: "Enter name of device with OTP codes.",
    forTag: "otp-names",
    props: {
      type: "text",
      name: "otp-names",
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
      },
      name: "otp-validate",
    },
  });

  return [
    el("h4", null, "One Time Codes"),
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
        onClick: (event) => {
          event.preventDefault();
          console.log("Hello OTP!");
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
