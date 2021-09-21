import { withErr } from "/static/js/utils.js";

class AuthorizationRequiredError extends Error {
  constructor(message) {
    super(message);
    this.name = "AuthorizationRequiredError";
  }
}

async function who() {
  let [res, err] = await withErr(fetch("/who", {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
  }));
  if (err) {
    return [null, err];
  }
  if (!res.ok) {
    err = new AuthorizationRequiredError("User is not authorized.");
    return [null, err];
  }
  let [parsed, jsonErr] = await withErr(res.json());
  if (jsonErr) {
    return [null, err];
  }
  return [parsed, null];
}

async function updatePassword(userID, { oldPass, newPass }) {
  let [res, err] = await withErr(fetch(`/api/v1/users/${userID}/password`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    redirect: "follow",
    body: JSON.stringify({
      "old": oldPass,
      "new": newPass,
    }),
  }));
  if (err) {
    return [null, err];
  }
  return [res, null];
}

async function optionsOTP() {
  let [res, errOptions] = await withErr(fetch("/api/v1/twofactor/otp/options", {
    method: "GET",
    redirect: "follow",
    credentials: "include",
  }));
  if (errOptions) {
    return [null, err];
  }
  let [jsonRes, errJson] = await withErr(res.json());
  if (errJson) {
    return [null, errJson];
  }
  return [jsonRes, null];
}

async function newOTP(body) {
  let [res, errPost] = await withErr(fetch("/api/v1/twofactor/otp", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    redirect: "follow",
    credentials: "include",
    body: JSON.stringify(body),
  }));
  if (errPost) {
    return [null, errPost];
  }
  let [jsonRes, errJson] = await withErr(res.json());
  if (errJson) {
    return [null, errJson];
  }
  return [jsonRes, null];
}

export { newOTP, optionsOTP, updatePassword, who };
